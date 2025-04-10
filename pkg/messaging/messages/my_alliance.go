package messages

import (
	"github.com/szcvak/sps/pkg/core"
)

type MyAllianceMessage struct{}

func NewMyAllianceMessage() *MyAllianceMessage {
	return &MyAllianceMessage{}
}

func (m *MyAllianceMessage) PacketId() uint16 {
	return 24399
}

func (m *MyAllianceMessage) PacketVersion() uint16 {
	return 1
}

func (m *MyAllianceMessage) Marshal() []byte {
	stream := core.NewByteStreamWithCapacity(4)

	stream.Write(core.VInt(0))
	stream.Write(false)

	return stream.Buffer()
}
