package main

import (
	"bytes"
	"crypto/aes"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/anonymous5l/wcdb/protobuf"
	"github.com/urfave/cli/v2"
	"google.golang.org/protobuf/proto"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
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

func aesDecrypt(pass, data []byte) ([]byte, error) {
	// aes-128-ecb only take front 16
	cipher, err := aes.NewCipher(pass[:16])
	if err != nil {
		return nil, err
	}
	size := cipher.BlockSize()
	for bs, be := 0, size; bs < len(data); bs, be = bs+size, be+size {
		cipher.Decrypt(data[bs:be], data[bs:be])
	}

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

func getExt(data []byte) string {
	var ext string
	if bytes.Compare(data[:4], []byte{0xff, 0xd8, 0xff, 0xe0}) == 0 {
		ext = ".jpg"
	}
	if bytes.Compare(data[:4], []byte{0xff, 0xd8, 0xff, 0xe1}) == 0 {
		ext = ".jpeg"
	}
	if bytes.Compare(data[:4], []byte{0x89, 0x50, 0x4E, 0x47}) == 0 {
		ext = ".png"
	}
	if bytes.Compare(data[:4], []byte{0x00, 0x00, 0x00, 0x14}) == 0 {
		ext = ".mov"
	}
	if bytes.Compare(data[:4], []byte{0x00, 0x00, 0x00, 0x20}) == 0 {
		ext = ".mp4"
	}
	if bytes.Compare(data[:4], []byte{0x00, 0x00, 0x00, 0x1c}) == 0 {
		ext = ".mp4"
	}
	if bytes.Compare(data[:4], []byte{0x52, 0x61, 0x72, 0x21}) == 0 {
		ext = ".rar"
	}
	if bytes.Compare(data[:2], []byte{0x50, 0x4B}) == 0 {
		ext = ".zip"
	}
	if bytes.Compare(data[3:10], []byte{0x53, 0x49, 0x4C, 0x4B, 0x5F, 0x56, 0x33}) == 0 {
		ext = ".aud"
	}
	if bytes.Compare(data[:4], []byte{0x25, 0x50, 0x44, 0x46}) == 0 {
		ext = ".pdf"
	}
	return ext
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
		if chunk, err = aesDecrypt([]byte(pass), chunk[:n]); err != nil {
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

func actionChat(ctx *cli.Context) error {
	dbName := ctx.String("db")
	resource := ctx.String("resource")
	talker := ctx.String("talker")
	pass := Pass(ctx.String("pass"))
	media := ctx.Bool("media")
	if !pass.Valid() {
		return ErrInvalidPassKey
	}

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

	msg, err := db.MsgSegment(talkerId)
	if err != nil {
		return err
	}

	bakFilename := filepath.Join(resource, msg.FilePath)
	bakFile, err := os.Open(bakFilename)
	if err != nil {
		return err
	}
	if _, err = bakFile.Seek(msg.OffSet, io.SeekStart); err != nil {
		return err
	}
	data := make([]byte, msg.Length, msg.Length)
	dataSize, err := bakFile.Read(data)
	if err != nil {
		return err
	}
	if dataSize != len(data) {
		return errors.New("invalid BAK file")
	}

	if data, err = aesDecrypt([]byte(pass), data); err != nil {
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

	strs := bytes.NewBufferString("")
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

		if message.FromUserName.GetStr() == talker {
			strs.WriteString("->:")
		} else {
			strs.WriteString("  <-:")
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
		// 42 name card
		// 49 composite message
		//   type - 6 - file
		//   type - 33 - applet
		//   type - 36 - app share
		//   type - 17 - realtime location share
		//   type - 2000 - money transfer
		//   type - 2001 - lucky money
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
			case 33:
				strs.WriteString("[Applet: ")
				strs.WriteString(xmlMessage.AppMsg.Title)
				strs.WriteString("]")
			case 36:
				strs.WriteString("[App: ")
				strs.WriteString(xmlMessage.AppMsg.Title)
				strs.WriteString("]")
			case 5:
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
			var xmlMessage XmlMessage
			if err = xml.Unmarshal([]byte(content), &xmlMessage); err != nil {
				return err
			}

			strs.WriteString("[Voice: ")
			strs.WriteString(strconv.FormatFloat(float64(xmlMessage.Voice.VoiceLength)/1000, 'f', 1, 64))
			strs.WriteByte('s')
			strs.WriteString("]")
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
	}

	fmt.Println(strs.String())

	return nil
}
