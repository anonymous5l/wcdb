package main

import (
	"database/sql"
	"encoding/xml"
)

type Config struct {
	Key       string
	Reserved0 int
	Buf       []byte
	Reserved1 int
	Reserved2 string
}

type MsgFileSegment struct {
	MapKey      int
	InnerOffSet int
	Length      int
	TotalLen    int
	OffSet      int64
	Reserved1   int
	FileName    string
	Reserved2   sql.NullString
	Reserved3   sql.NullString
	Reserved4   int
}

type Session struct {
	Talker    string
	EndTime   int64
	TotalSize int64
	NickName  string
	Reserved0 int
	Reserved1 int
	Reserved2 sql.NullString
	Reserved3 sql.NullString
	StartTime int64
	Reserved5 sql.NullString
}

type Name2ID struct {
	UsrName string
}

type MsgSegment struct {
	TalkerId  int
	StartTime int
	EndTime   int
	OffSet    int64
	Length    int
	UsrName   string
	Status    int
	Reserved1 int
	FilePath  string
	SegmentId string
	Reserved2 sql.NullString
	Reserved3 sql.NullString
}

type MsgMedia struct {
	TalkerId     int
	MediaId      int
	MsgSegmentId int64
	SrvId        int
	MD5          sql.NullString
	Talker       string
	MediaIdStr   string
	Reserved0    int
	Reserved1    int
	Reserved2    sql.NullString
}

type cdata struct {
	Value string `xml:",cdata"`
}

type XmlAppAttach struct {
	XMLName        xml.Name `xml:"appattach"`
	TotalLen       int      `xml:"totallen"`
	AttachId       string   `xml:"attachid"`
	EmoticonMD5    string   `xml:"emoticonmd5"`
	FileExt        string   `xml:"fileext"`
	CDNAttachURL   string   `xml:"cdnattachurl,omitempty"`
	CDNThumbURL    string   `xml:"cdnthumburl,omitempty"`
	CDNThumbMD5    string   `xml:"cdnthumbmd5,omitempty"`
	CDNThumbLength int      `xml:"cdnthumblength,omitempty"`
	CDNThumbWidth  int      `xml:"cdnthumbwidth,omitempty"`
	CDNThumbHeight int      `xml:"cdnthumbheight,omitempty"`
	CDNThumbAESKey string   `xml:"cdnthumbaeskey"`
	AESKey         string   `xml:"aeskey"`
	EncryVer       int      `xml:"encryver"`
	FileKey        string   `xml:"filekey,omitempty"`
	CDNThumbCRC    uint32   `xml:"-"`
}

type XmlAppMessage struct {
	XMLName           xml.Name      `xml:"appmsg"`
	AppId             string        `xml:"appid,attr"`
	SdkVer            int           `xml:"sdkver,attr"`
	Title             string        `xml:"title"`
	Desc              string        `xml:"des"`
	Action            string        `xml:"action"`
	Type              int           `xml:"type"`
	ShowType          int           `xml:"showtype"`
	SoundType         int           `xml:"soundtype"`
	MediaTagName      string        `xml:"mediatagname"`
	MessageExt        string        `xml:"messageext"`
	MessageAction     string        `xml:"messageaction"`
	Content           string        `xml:"content"`
	ContentAttr       int           `xml:"contentattr"`
	URL               string        `xml:"url"`
	LowURL            string        `xml:"lowurl"`
	DataURL           string        `xml:"dataurl"`
	LowDataURL        string        `xml:"lowdataurl"`
	SongAlbumURL      string        `xml:"songalbumurl,omitempty"`
	SongLyric         string        `xml:"soinglyric,omitempty"`
	AppAttach         *XmlAppAttach `xml:"appattach,omitempty"`
	ExtInfo           string        `xml:"extinfo"`
	SourceUserName    string        `xml:"sourceusername"`
	SourceDisplayName string        `xml:"sourcedisplayname"`
	ThumbURL          string        `xml:"thumburl"`
	MD5               string        `xml:"md5"`
	StatExtStr        string        `xml:"Statextstr"`
	XMLFullLen        uint32        `xml:"xmlfulllen,omitempty"`
	DirectShare       int           `xml:"directshare,omitempty"`
	RecordItem        *cdata        `xml:"recorditem,omitempty"`
}

type XmlAppInfo struct {
	XMLName xml.Name `xml:"appinfo"`
	Version int      `xml:"version"`
	AppName string   `xml:"appname"`
}

type XmlVoice struct {
	XMLName      xml.Name `xml:"voicemsg"`
	EndFlag      int      `xml:"endflag,attr"`
	CancelFlag   int      `xml:"cancelflag,attr"`
	ForwardFlag  int      `xml:"forwardflag,attr"`
	VoiceFormat  int      `xml:"voiceformat,attr"`
	VoiceLength  int      `xml:"voicelength,attr"`
	Length       int      `xml:"length,attr"`
	BufferId     string   `xml:"bufid,attr"`
	AESKey       string   `xml:"aeskey,attr"`
	VoiceURL     string   `xml:"voiceurl,attr"`
	VoiceMD5     string   `xml:"voicemd5,attr"`
	ClientMsgId  string   `xml:"clientmsgid,attr"`
	FromUserName string   `xml:"fromusername,attr"`
}

type XmlEmoji struct {
	XMLName           xml.Name `xml:"emoji"`
	FromUsername      string   `xml:"fromusername,attr"`
	ToUsername        string   `xml:"tousername,attr"`
	Type              int      `xml:"type,attr"`
	IdBuffer          string   `xml:"idbuffer,attr"`
	MD5               string   `xml:"md5,attr"`
	Len               int      `xml:"len,attr"`
	ProductId         string   `xml:"productid,attr"`
	AndroidMD5        string   `xml:"androidmd5,attr"`
	AndroidLen        int      `xml:"androidlen,attr"`
	S60V3MD5          string   `xml:"s60v3md5,attr"`
	S60V3Len          int      `xml:"s60v3len,attr"`
	S60V5MD5          string   `xml:"s60v5md5,attr"`
	S60V5Len          int      `xml:"s60v5len,attr"`
	CDNURL            string   `xml:"cdnurl,attr"`
	DesignerId        string   `xml:"designerid,attr"`
	ThumbURL          string   `xml:"thumburl,attr"`
	EncryptURL        string   `xml:"encrypturl,attr"`
	AESKey            string   `xml:"aeskey,attr"`
	ExternURL         string   `xml:"externurl,attr"`
	ExternMD5         string   `xml:"externmd5,attr"`
	Width             int      `xml:"width,attr"`
	Height            int      `xml:"height,attr"`
	TPURL             string   `xml:"tpurl,attr"`
	TPAuthKey         string   `xml:"tpauthkey,attr"`
	AttachedText      string   `xml:"attachedtext,attr"`
	AttachedTextColor string   `xml:"attachedtextcolor,attr"`
	LenSId            string   `xml:"lensid,attr"`
	EmojiAttr         string   `xml:"emojiattr,attr"`
	LinkId            string   `xml:"linkid,attr"`
}

type XmlImage struct {
	XMLName        xml.Name `xml:"img"`
	CDNBigImgURL   string   `xml:"cdnbigimgurl,attr"`
	HDLength       int      `xml:"hdlength,attr"`
	CDNHDHeight    int      `xml:"cdnhdheight,attr"`
	Length         int      `xml:"length,attr"`
	CDNThumbAESKey string   `xml:"cdnthumbaeskey,attr"`
	MD5            string   `xml:"md5,attr"`
	CDNHDWidth     int      `xml:"cdnhdwidth,attr"`
	CDNThumbWidth  int      `xml:"cdnthumbwidth,attr"`
	CDNThumbHeight int      `xml:"cdnthumbheight,attr"`
	AESKey         string   `xml:"aeskey,attr"`
	CDNMidWidth    int      `xml:"cdnmidwidth,attr"`
	CDNMidHeight   int      `xml:"cdnmidheight,attr"`
	CDNThumbLength int      `xml:"cdnthumblength,attr"`
	EncryptVer     int      `xml:"encryver,attr"`
	CDNMidImgURL   string   `xml:"cdnmidimgurl,attr"`
	CDNThumbURL    string   `xml:"cdnthumburl,attr"`
	FileKey        string   `xml:"filekey,attr"`
}

type XmlVideo struct {
	XMLName           xml.Name `xml:"videomsg"`
	ClientMsgId       string   `xml:"clientmsgid,attr"`
	PlayLength        int      `xml:"playlength,attr"`
	Length            int      `xml:"length,attr"`
	Type              int      `xml:"type,attr"`
	FromUserName      string   `xml:"fromusername,attr"`
	AESKey            string   `xml:"aeskey,attr"`
	CDNVideoURL       string   `xml:"cdnvideourl,attr"`
	CDNThumbURL       string   `xml:"cdnthumburl,attr"`
	CDNThumbLength    int      `xml:"cdnthumblength,attr"`
	CDNThumbWidth     int      `xml:"cdnthumbwidth,attr"`
	CDNThumbHeight    int      `xml:"cdnthumbheight,attr"`
	CDNThumbAESKey    string   `xml:"cdnthumbaeskey,attr"`
	EncryptVer        int      `xml:"encryver,attr"`
	FileParam         string   `xml:"fileparam,attr"`
	MD5               string   `xml:"md5,attr"`
	NewMD5            string   `xml:"newmd5,attr"`
	IsPlaceHolder     int      `xml:"isplaceholder,attr"`
	RawLength         int      `xml:"rawlength,attr"`
	CDNRawVideoURL    string   `xml:"cdnrawvideourl,attr"`
	CDNRawVideoAESKey string   `xml:"cdnrawvideoaeskey,attr"`
}

type XmlLocation struct {
	XMLName xml.Name `xml:"location"`
	X       float64  `xml:"x,attr"`
	Y       float64  `xml:"y,attr"`
	Scale   int      `xml:"scale,attr"`
	Label   string   `xml:"label,attr"`
	POIName string   `xml:"poiname,attr"`
	InfoURL string   `xml:"infourl,attr"`
	MapType string   `xml:"maptype,attr"`
	POIId   string   `xml:"poiid,attr"`
}

type XmlNameCard struct {
	XMLName                 xml.Name `xml:"msg"`
	BigHeadImgURL           string   `xml:"bigheadimgurl,attr"`
	SmallHeadImgURL         string   `xml:"smallheadimgurl,attr"`
	UserName                string   `xml:"username,attr"`
	NickName                string   `xml:"nickname,attr"`
	FullPY                  string   `xml:"fullpy,attr"`
	ShortPY                 string   `xml:"shortpy,attr"`
	Alias                   string   `xml:"alias,attr"`
	ImageStatus             int      `xml:"imagestatus,attr"`
	Scene                   int      `xml:"scene,attr"`
	Province                string   `xml:"province,attr"`
	City                    string   `xml:"city,attr"`
	Sign                    string   `xml:"sign,attr"`
	Sex                     int      `xml:"sex,attr"`
	CertFlag                int      `xml:"certflag"`
	CertInfo                string   `xml:"certinfo"`
	BrandIconURL            string   `xml:"brandIconUrl"`
	BrandHomeURL            string   `xml:"brandHomeUrl,attr"`
	BrandSubscriptConfigURL string   `xml:"brandSubscriptConfigUrl,attr"`
	BrandFlags              int      `xml:"brandFlags,attr"`
	RegionCode              string   `xml:"regionCode,attr"`
	AntiSpamTicket          string   `xml:"antispamticket,attr"`
}

type XmlMessage struct {
	XMLName      xml.Name       `xml:"msg"`
	AppMsg       *XmlAppMessage `xml:"appmsg,omitempty"`
	FromUserName string         `xml:"fromusername"`
	Scene        int            `xml:"scene"`
	AppInfo      *XmlAppInfo    `xml:"appinfo,omitempty"`
	CommentURL   string         `xml:"commenturl"`
	Voice        *XmlVoice      `xml:"voicemsg,omitempty"`
	Emoji        *XmlEmoji      `xml:"emoji,omitempty"`
	Image        *XmlImage      `xml:"img,omitempty"`
	Video        *XmlVideo      `xml:"videomsg,omitempty"`
	Location     *XmlLocation   `xml:"location,omitempty"`
}

type XmlVoIPBubble struct {
	XMLName    xml.Name `xml:"VoIPBubbleMsg"`
	Msg        string   `xml:"msg"`
	RoomType   int      `xml:"room_type"`
	RedDot     bool     `xml:"red_dot"`
	RoomId     int      `xml:"roomid"`
	RoomKey    string   `xml:"roomkey"`
	InviteId   int      `xml:"inviteid"`
	MsgType    int      `xml:"msg_type"`
	Timestamp  int64    `xml:"timestamp"`
	Identity   string   `xml:"identity"`
	Duration   int      `xml:"duration"`
	InviteId64 int64    `xml:"inviteid64"`
	Business   int      `xml:"business"`
}

type XmlVoIP struct {
	XMLName       xml.Name       `xml:"voipmsg"`
	Type          string         `xml:"type,attr"`
	VoIPBubbleMsg *XmlVoIPBubble `xml:"VoIPBubbleMsg,omitempty"`
}

type XmlVoIPInviteMsg struct {
	XMLName    xml.Name `xml:"voipinvitemsg"`
	RoomId     int      `xml:"roomid"`
	Key        string   `xml:"key"`
	Status     int      `xml:"status"`
	InviteType int      `xml:"invitetype"`
}

type XmlVoIPExtInfo struct {
	XMLName  xml.Name `xml:"voipextinfo"`
	RecvTime int64    `xml:"recvtime"`
}

type XmlVoIPLocalInfo struct {
	XMLName     xml.Name `xml:"voiplocalinfo"`
	WordingType int      `xml:"wordingtype"`
	Duration    int      `xml:"duration"`
}

type XmlOldVoIP struct {
	VoIPInviteMsg XmlVoIPInviteMsg `xml:"voipinvitemsg"`
	VoIPExtInfo   XmlVoIPExtInfo   `xml:"voipextinfo"`
	VoIPLocalInfo XmlVoIPLocalInfo `xml:"voiplocalinfo"`
}
