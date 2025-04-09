package messages

import (
	"github.com/szcvak/sps/pkg/core"
)

type LoginOkMessage struct {
	loginMessage *LoginMessage
}

func NewLoginOkMessage(loginMessage *LoginMessage) *LoginOkMessage {
	return &LoginOkMessage{
		loginMessage: loginMessage,
	}
}

func (l *LoginOkMessage) PacketId() uint16 {
	return 20104
}

func (l *LoginOkMessage) PacketVersion() uint16 {
	return 1
}

func (l *LoginOkMessage) Marshal() []byte {
	stream := core.NewByteStreamWithCapacity(64)

	stream.Write(l.loginMessage.HighId)
	stream.Write(l.loginMessage.LowId)

	stream.Write(l.loginMessage.HighId)
	stream.Write(l.loginMessage.LowId)

	stream.Write(l.loginMessage.Token)

	stream.Write("467606826913688")
	stream.Write("G:325378671")

	stream.Write(l.loginMessage.Major)
	stream.Write(l.loginMessage.Minor)
	stream.Write(l.loginMessage.Build)

	stream.Write("-dev")

	stream.Write(0)
	stream.Write(0)
	stream.Write(0)

	stream.Write(core.EmptyString)
	stream.Write(core.EmptyString)
	stream.Write(core.EmptyString)

	stream.Write(0)

	stream.Write(core.EmptyString)
	stream.Write(l.loginMessage.Region)
	stream.Write(core.EmptyString)

	stream.Write(1)

	stream.Write(core.EmptyString)
	stream.Write(core.EmptyString)
	stream.Write(core.EmptyString)

	return stream.Buffer()
}
