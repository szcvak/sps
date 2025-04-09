package messages

import (
	"fmt"

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
	l.SystemLanguage, _ = stream.ReadVInt()

	l.Region, _ = stream.ReadString()

	l.unmarshalled = true
}

func (l *LoginMessage) Process(wrapper *core.ClientWrapper) {
	if !l.Unmarshalled() {
		return
	}

	fmt.Printf("%d, %d, %s, %d.%d.%d, %s, %s, %s, %d, %s", l.HighId, l.LowId, l.Token, l.Major, l.Minor, l.Build, l.FingerprintSha, l.DeviceUuid, l.DeviceIdentifier, l.SystemLanguage, l.Region)

	msg := NewLoginOkMessage(l)
	wrapper.Send(msg.PacketId(), msg.PacketVersion(), msg.Marshal())
}
