package messages

import (
	"github.com/szcvak/sps/pkg/core"
)

type ClanStreamMessage struct{}

func NewClanStreamMessage() *ClanStreamMessage {
	return &ClanStreamMessage{}
}

func (c *ClanStreamMessage) PacketId() uint16 {
	return 24311
}

func (c *ClanStreamMessage) PacketVersion() uint16 {
	return 1
}

func (c *ClanStreamMessage) Marshal() []byte {
	stream := core.NewByteStreamWithCapacity(8)

	stream.Write(core.VInt(0))
	stream.Write(core.VInt(0))

	return stream.Buffer()
}
