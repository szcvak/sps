package messages

import (
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/messaging"
)

type LoginFailedMessage struct {
	loginMessage *LoginMessage
	text         string
	reason       messaging.LoginFailedReason
}

func NewLoginFailedMessage(loginMessage *LoginMessage, text string, reason messaging.LoginFailedReason) *LoginFailedMessage {
	return &LoginFailedMessage{
		loginMessage: loginMessage,
		text:         text,
		reason:       reason,
	}
}

func (l *LoginFailedMessage) PacketId() uint16 {
	return 20103
}

func (l *LoginFailedMessage) PacketVersion() uint16 {
	return 1
}

func (l *LoginFailedMessage) Marshal() []byte {
	stream := core.NewByteStreamWithCapacity(64)

	stream.Write(int32(l.reason))

	stream.Write(l.loginMessage.FingerprintSha)

	stream.Write("prod.sps.q4.lol:9339")
	stream.Write("https://game-assets.brawlstarsgame.com")
	stream.Write("https://github.com/szcvak/sps")

	stream.Write(l.text)

	stream.Write(0)
	stream.Write(false)

	stream.Write(core.EmptyString)
	stream.Write(core.EmptyString)

	stream.Write(0)
	stream.Write(3)

	stream.Write(core.EmptyString)
	stream.Write(core.EmptyString)

	stream.Write(0)
	stream.Write(0)

	stream.Write(false)
	stream.Write(false)

	return stream.Buffer()
}
