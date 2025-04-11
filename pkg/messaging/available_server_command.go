package messaging

import (
	"github.com/szcvak/sps/pkg/core"
)

type AvailableServerCommandMessage struct {
	id      int
	payload interface{}
}

func NewAvailableServerCommandMessage(id int, payload interface{}) *AvailableServerCommandMessage {
	return &AvailableServerCommandMessage{
		id:      id,
		payload: payload,
	}
}

func (a *AvailableServerCommandMessage) PacketId() uint16 {
	return 24111
}

func (a *AvailableServerCommandMessage) PacketVersion() uint16 {
	return 1
}

func (a *AvailableServerCommandMessage) Marshal() []byte {
	stream := core.NewByteStreamWithCapacity(32)

	factory, exists := ServerCommands[a.id]

	if !exists {
		return []byte{}
	}

	stream.Write(core.VInt(a.id))

	msg := factory(a.payload)
	msg.Marshal(stream)

	return stream.Buffer()
}
