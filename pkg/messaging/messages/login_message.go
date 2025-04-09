package messages

import (
	"github.com/szcvak/sps/pkg/core"
)

type LoginMessage struct {
	HighId int32
	LowId  int32

	Token string

	Major, Minor, Build int32

	FingerprintSha   string
	DeviceUuid       string
	DeviceIdentifier string
	Region           string

	SystemLanguage int32

	unmarshalled bool
}

func NewLoginMessage() *LoginMessage {
	return &LoginMessage{}
}

func (l *LoginMessage) Unmarshalled() bool {
	return l.unmarshalled
}

func (l *LoginMessage) Unmarshal(payload []byte) {
	stream := core.NewByteStream(payload)

	l.HighId, _ = stream.ReadInt()
	l.LowId, _ = stream.ReadInt()

	l.Token, _ = stream.ReadString()

	l.Major, _ = stream.ReadInt()
	l.Minor, _ = stream.ReadInt()
	l.Build, _ = stream.ReadInt()

	l.FingerprintSha, _ = stream.ReadString()
	_, _ = stream.ReadString()

	l.DeviceUuid, _ = stream.ReadString()
	_, _ = stream.ReadString()

	l.DeviceIdentifier, _ = stream.ReadString()

	lang, _ := stream.ReadVInt()
	l.SystemLanguage = int32(lang)

	l.Region, _ = stream.ReadString()

	l.unmarshalled = true
}

func (l *LoginMessage) Process(wrapper *core.ClientWrapper) {
	if !l.Unmarshalled() {
		return
	}

	wrapper.Player.Token = l.Token

	wrapper.Player.HighId = l.HighId
	wrapper.Player.LowId = l.LowId

	wrapper.Player.LoggedIn = true

	msg := NewLoginOkMessage(l)
	wrapper.Send(msg.PacketId(), msg.PacketVersion(), msg.Marshal())
}
