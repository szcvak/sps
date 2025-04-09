package messages

import "github.com/szcvak/sps/pkg/core"

type OwnHomeDataMessage struct {
	wrapper *core.ClientWrapper
}

func NewOwnHomeDataMessage(wrapper *core.ClientWrapper) *OwnHomeDataMessage {
	return &OwnHomeDataMessage{
		wrapper: wrapper,
	}
}

func (o *OwnHomeDataMessage) PacketId() uint16 {
	return 24101
}

func (o *OwnHomeDataMessage) PacketVersion() uint16 {
	return 1
}

func (o *OwnHomeDataMessage) Marshal() []byte {
	player := o.wrapper.Player

	if player.State() != core.StateLogin {
		return make([]byte, 0)
	}

	stream := core.NewByteStreamWithCapacity(128)

	return stream.Buffer()
}
