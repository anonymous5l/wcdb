package main

import (
	"bytes"
	"crypto/aes"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/anonymous5l/wcdb/protobuf"
	"github.com/h2non/filetype"
	"github.com/urfave/cli/v2"
	"google.golang.org/protobuf/proto"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var ChatCommand = &cli.Command{
	Name:   "chat",
	Usage:  "take talker chat message from decrypted Backup.db",
	Action: actionChat,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "db",
			Usage:   "decrypted Backup.db file path",
			Value:   "Backup.db",
			Aliases: []string{"d"},
		},
		&cli.StringFlag{
			Name:     "resource",
			Usage:    "BAK_0_XXX folder path",
			Required: true,
			Aliases:  []string{"r"},
		},
		&cli.StringFlag{
			Name:     "talker",
			Usage:    "dump chat list talker",
			Required: true,
			Aliases:  []string{"t"},
		},
		&cli.StringFlag{
			Name:     "pass",
			Usage:    "decrypt media resource file chunk key",
			Required: true,
			Aliases:  []string{"p"},
		},
		&cli.BoolFlag{
			Name:    "media",
			Usage:   "dump media file output directory ./res/<Talker>",
			Value:   false,
			Aliases: []string{"m"},
		},
	},
}

var fds sync.Map

func releaseFds() {
	fds.Range(func(key, value any) bool {
		fd := value.(*os.File)
		fd.Close()
		return true
	})
}

func aesDecrypt(pass, data []byte, usePadding bool) ([]byte, error) {
	// aes-128-ecb only take front 16
	cipher, err := aes.NewCipher(pass[:16])
	if err != nil {
		return nil, err
	}
	size := cipher.BlockSize()
	for bs, be := 0, size; bs < len(data); bs, be = bs+size, be+size {
		cipher.Decrypt(data[bs:be], data[bs:be])
	}

	if usePadding {
		// pkcs7 un pad
		isPad := true
		pad := int(data[len(data)-1])
		for i := 1; i < pad; i++ {
			if int(data[len(data)-1-i]) != pad {
				isPad = false
				break
			}
		}
		if isPad {
			data = data[:len(data)-pad]
		}
	}

	return data, nil
}

func getFd(path string) *os.File {
	if o, ok := fds.Load(path); ok {
		return o.(*os.File)
	}

	o, err := os.Open(path)
	if err != nil {
		return nil
	}
	fds.Store(path, o)
	return o
}

func init() {
	filetype.AddMatcher(filetype.AddType("aud", ""), func(i []byte) bool {
		if bytes.Compare(i[:10], []byte{0x02, 0x23, 0x21, 0x53, 0x49, 0x4C, 0x4B, 0x5F, 0x56, 0x33}) == 0 {
			return true
		}
		return false
	})
	filetype.AddMatcher(filetype.AddType("doc", ""), func(i []byte) bool {
		if bytes.Compare(i[:8], []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}) == 0 {
			if len(i) > 512 {
				if bytes.Compare(i[512:516], []byte{0xec, 0xa5, 0xc1, 0x00}) == 0 {
					return true
				}
			}
		}
		return false
	})
	filetype.AddMatcher(filetype.AddType("xls", ""), func(i []byte) bool {
		if bytes.Compare(i[:8], []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}) == 0 {
			if len(i) > 512 {
				if bytes.Compare(i[512:516], []byte{0xfd, 0xff, 0xff, 0xff}) == 0 && i[516] == 0 {
					return true
				} else if bytes.Compare(i[512:516], []byte{0xfd, 0xff, 0xff, 0xff}) == 0 && i[516] == 2 {
					return true
				} else if bytes.Compare(i[512:520], []byte{0x09, 0x08, 0x10, 0x00, 0x00, 0x06, 0x05, 0x00}) == 0 {
					return true
				}
			}
		}
		return false
	})
	filetype.AddMatcher(filetype.AddType("ppt", ""), func(i []byte) bool {
		if bytes.Compare(i[:8], []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}) == 0 {
			if len(i) > 512 {
				if bytes.Compare(i[512:516], []byte{0xa0, 0x46, 0x1d, 0xf0}) == 0 {
					return true
				} else if bytes.Compare(i[512:516], []byte{0x00, 0x6e, 0x1e, 0xf0}) == 0 {
					return true
				} else if bytes.Compare(i[512:516], []byte{0x0F, 0x00, 0xE8, 0x03}) == 0 {
					return true
				} else if bytes.Compare(i[512:516], []byte{0xfd, 0xff, 0xff, 0xff}) == 0 && i[517] == 0 && i[518] == 0 {
					return true
				}
			}
		}
		return false
	})
}

func getExt(data []byte) string {
	t, err := filetype.Match(data)
	if err != nil {
		return ""
	}

	return "." + t.Extension
}

func dumpFile(db *BackupDB, resource, pass, dir, id string) error {
	media, err := db.MsgMedia(id)
	if err != nil {
		return err
	}

	file, err := db.FileSegment(media.MediaId)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer([]byte{})
	for i := 0; i < len(file); i++ {
		f := file[i]
		fd := getFd(filepath.Join(resource, f.FileName))
		if _, err = fd.Seek(f.OffSet, io.SeekStart); err != nil {
			return err
		}
		chunk := make([]byte, f.Length, f.Length)
		n, err := fd.Read(chunk)
		if err != nil {
			return err
		}
		if chunk, err = aesDecrypt([]byte(pass), chunk[:n], i == len(file)-1); err != nil {
			return err
		}
		buf.Write(chunk)
	}

	totalBuf := buf.Bytes()

	ext := getExt(totalBuf)

	o, err := os.Create(filepath.Join(dir, id) + ext)
	if err != nil {
		return err
	}
	defer o.Close()

	if _, err = o.Write(totalBuf); err != nil {
		return err
	}

	return nil
}

func printGroupMessage(strs *bytes.Buffer, msg XmlMessage) (err error) {
	if msg.AppMsg.RecordItem == nil {
		strs.WriteString("[CombineMessage]")
		return nil
	}

	var dataList *XmlRecordMessageDataList

	if strings.HasPrefix(msg.AppMsg.RecordItem.Value, "<recorditem>") {
		var combine XmlMessageRecordItem
		if err = xml.Unmarshal([]byte(msg.AppMsg.RecordItem.Value), &combine); err != nil {
			return
		}
		dataList = combine.RecordInfo.DataList
	} else if strings.HasPrefix(msg.AppMsg.RecordItem.Value, "<recordinfo>") {
		var combine XmlRecordMessage
		if err = xml.Unmarshal([]byte(msg.AppMsg.RecordItem.Value), &combine); err != nil {
			return
		}
		dataList = combine.DataList
	} else {
		fmt.Println(msg.AppMsg.RecordItem.Value)
	}

	strs.WriteString("[CombineMessage\n")
	items := dataList.DataItems
	for i := 0; i < len(items); i++ {
		item := items[i]
		strs.WriteString(item.SourceName)
		strs.WriteString(" : ")
		if item.DataDesc != "" {
			strs.WriteString(item.DataDesc)
		} else {
			strs.WriteString(item.DataTitle)
		}
		strs.WriteString("\n")
	}
	strs.WriteString("]")

	return nil
}

func actionChat(ctx *cli.Context) error {
	dbName := ctx.String("db")
	resource := ctx.String("resource")
	talker := ctx.String("talker")
	pass := Pass(ctx.String("pass"))
	media := ctx.Bool("media")
	if !pass.Valid() {
		return ErrInvalidPassKey
	}

	defer releaseFds()

	db, err := NewBackupDB(dbName)
	if err != nil {
		return err
	}
	defer db.Close()

	ids, err := db.Name2ID()
	if err != nil {
		return err
	}

	getTalkerId := func(name string) int {
		for i := 0; i < len(ids); i++ {
			if ids[i].UsrName == name {
				return i + 1
			}
		}
		return -1
	}

	talkerId := getTalkerId(talker)

	if talkerId < 0 {
		return errors.New("no record")
	}

	strs := bytes.NewBufferString("")

	msgs, err := db.MsgSegment(talkerId)
	if err != nil {
		return err
	}

	for i := 0; i < len(msgs); i++ {
		msg := msgs[i]

		bakFilename := filepath.Join(resource, msg.FilePath)

		bakFile := getFd(bakFilename)

		if _, err = bakFile.Seek(msg.OffSet, io.SeekStart); err != nil {
			return err
		}

		var dataSize int

		data := make([]byte, msg.Length, msg.Length)

		if dataSize, err = bakFile.Read(data); err != nil {
			return err
		}
		if dataSize != len(data) {
			return errors.New("invalid BAK file")
		}

		if data, err = aesDecrypt([]byte(pass), data, true); err != nil {
			return err
		}

		resourcePath := filepath.Join("res", talker)
		if media {
			if _, err = os.Stat(resourcePath); err != nil {
				if os.IsNotExist(err) {
					if err = os.MkdirAll(resourcePath, 0755); err != nil {
						return err
					}
				}
			}
		}

		var pMessages protobuf.BakChatMsgList
		if err = proto.Unmarshal(data, &pMessages); err != nil {
			return err
		}

		strs.Reset()
		for i := 0; i < int(pMessages.GetCount()); i++ {
			message := pMessages.GetList()[i]

			if strs.Len() > 0 {
				strs.WriteString("\n")
			}

			if media {
				medias := message.GetMediaId()
				for j := 0; j < len(medias); j++ {
					id := medias[j]
					if err = dumpFile(db, resource, string(pass), resourcePath, id.GetStr()); err != nil {
						return err
					}
				}
			}

			msgId := message.GetNewMsgId()

			strs.WriteString(fmt.Sprintf("%20d", msgId))
			strs.WriteString(" | ")

			millTime := time.UnixMilli(message.GetClientMsgMillTime())

			if message.FromUserName.GetStr() == talker {
				strs.WriteString("\x1B[1;37m(")
				strs.WriteString(millTime.Format("2006-01-02 15:04:05"))
				strs.WriteString(") -> : ")
			} else {
				strs.WriteString("\x1B[1;32m(")
				strs.WriteString(millTime.Format("2006-01-02 15:04:05"))
				strs.WriteString(") <- : ")
			}

			// 10000 be contacts notify message
			// 1 text message
			// 3 image
			// 34 voice message
			// 47 emoji
			// 62 short video
			// 50 voip message
			// 48 location
			// 76 qq music
			// 3 netease music
			// 4 red book
			// 42 name card
			// 19 group join refer message
			// 49 composite message
			//   type - 6 - file
			//   type - 57 - refer message
			//   type - 33 - applet
			//   type - 36 - app share
			//   type - 17 - realtime location share
			//   type - 2000 - money transfer
			//   type - 2001 - lucky money
			// 62 tickle
			msgType := message.GetType()

			content := message.GetContent().GetStr()

			switch msgType {
			case 42:
				var xmlMessage XmlNameCard
				if err = xml.Unmarshal([]byte(content), &xmlMessage); err != nil {
					return err
				}
				strs.WriteString("[NameCard: ")
				strs.WriteString(xmlMessage.NickName)
				strs.WriteString("]")
			case 49:
				var xmlMessage XmlMessage
				if err = xml.Unmarshal([]byte(content), &xmlMessage); err != nil {
					return err
				}
				if xmlMessage.AppMsg == nil {
					break
				}
				switch xmlMessage.AppMsg.Type {
				case 19:
					if err = printGroupMessage(strs, xmlMessage); err != nil {
						return err
					}
				case 62:
					strs.WriteString("[Tickle]")
				case 57:
					strs.WriteString(xmlMessage.AppMsg.Title)
					strs.WriteString(" [Refer: ")
					strs.WriteString(xmlMessage.AppMsg.ReferMsg.DisplayName)
					strs.WriteString(":")
					strs.WriteString(xmlMessage.AppMsg.ReferMsg.Content)
					strs.WriteString("]")
				case 33:
					strs.WriteString("[Applet: ")
					strs.WriteString(xmlMessage.AppMsg.Title)
					strs.WriteString("]")
				case 36:
					strs.WriteString("[App: ")
					strs.WriteString(xmlMessage.AppMsg.Title)
					strs.WriteString("]")
				case 4, 5:
					strs.WriteString("[Link: ")
					strs.WriteString(xmlMessage.AppMsg.Title)
					strs.WriteString("]")
				case 76, 3:
					strs.WriteString("[Music: ")
					strs.WriteString(xmlMessage.AppMsg.Title)
					strs.WriteString("]")
				case 6:
					strs.WriteString("[File: ")
					strs.WriteString(xmlMessage.AppMsg.Title)
					strs.WriteString("]")
				case 2001, 2000:
					strs.WriteString("[" + xmlMessage.AppMsg.Title + "]")
				default:
					strs.WriteString(content)
				}
			case 48:
				var xmlMessage XmlMessage
				if err = xml.Unmarshal([]byte(content), &xmlMessage); err != nil {
					return err
				}
				strs.WriteString("[Location: ")
				strs.WriteString(xmlMessage.Location.Label)
				strs.WriteString("]")
			case 47:
				var xmlMessage XmlMessage
				if err = xml.Unmarshal([]byte(content), &xmlMessage); err != nil {
					return err
				}
				strs.WriteString("[Emoji: ")
				strs.WriteString(xmlMessage.Emoji.MD5)
				strs.WriteString("]")
			case 34:
				if strings.HasPrefix(content, "<msg>") {
					var xmlMessage XmlMessage
					if err = xml.Unmarshal([]byte(content), &xmlMessage); err != nil {
						return err
					}

					strs.WriteString("[Voice: ")
					strs.WriteString(strconv.FormatFloat(float64(xmlMessage.Voice.VoiceLength)/1000, 'f', 1, 64))
					strs.WriteByte('s')
					strs.WriteString("]")
				} else {
					strs.WriteString("[Voice]")
				}
			case 43:
				var xmlMessage XmlMessage
				if err = xml.Unmarshal([]byte(content), &xmlMessage); err != nil {
					return err
				}

				strs.WriteString("[Video: ")
				strs.WriteString(strconv.FormatInt(int64(xmlMessage.Video.PlayLength), 10))
				strs.WriteByte('s')
				strs.WriteString("]")
			case 50:
				if strings.HasPrefix(content, "<voipinvitemsg>") {
					var voip XmlOldVoIP
					if err = xml.Unmarshal([]byte("<xml>"+content+"</xml>"), &voip); err != nil {
						return err
					}

					if voip.VoIPInviteMsg.InviteType == 0 {
						strs.WriteString("[VideoCall: ")
					} else if voip.VoIPInviteMsg.InviteType == 1 {
						strs.WriteString("[VoiceCall: ")
					}

					switch voip.VoIPLocalInfo.WordingType {
					case 4:
						strs.WriteString(strconv.FormatInt(int64(voip.VoIPLocalInfo.Duration), 10))
						strs.WriteString("s")
					case 2:
						strs.WriteString("Canceled")
					case 3:
						strs.WriteString("Aborted")
					case 1: // FIXME maybe don't known
						strs.WriteString("Timeout")
					}

					strs.WriteString("]")
				} else {

					var voip XmlVoIP
					if err = xml.Unmarshal([]byte(content), &voip); err != nil {
						return err
					}

					if voip.VoIPBubbleMsg.RoomType == 0 {
						strs.WriteString("[VideoCall: ")
					} else if voip.VoIPBubbleMsg.RoomType == 1 {
						strs.WriteString("[VoiceCall: ")
					}

					strs.WriteString(voip.VoIPBubbleMsg.Msg)
					strs.WriteString("]")
				}
			case 3:
				var xmlMessage XmlMessage
				if err = xml.Unmarshal([]byte(content), &xmlMessage); err != nil {
					return err
				}

				strs.WriteString("[Image: ")
				strs.WriteString(xmlMessage.Image.MD5)
				strs.WriteString("]")
			case 1, 10000:
				strs.WriteString(content)
			default:
				strs.WriteString(strconv.FormatUint(uint64(msgType), 10))
				strs.WriteString(":")
				strs.WriteString(content)
			}

			strs.WriteString("\u001B[0m")
		}

		fmt.Println(strs.String())
	}

	return nil
}
