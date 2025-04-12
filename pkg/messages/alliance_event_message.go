package messages

import (
	"github.com/szcvak/sps/pkg/core"
)

type AllianceEventMessage struct {
	event core.VInt
}

func NewAllianceEventMessage(event int32) *AllianceEventMessage {
	return &AllianceEventMessage {
		event: core.VInt(event),
	}
}

func (a *AllianceEventMessage) PacketId() uint16 {
	return 24333
}

func (a *AllianceEventMessage) PacketVersion() uint16 {
	return 1
}

func (a *AllianceEventMessage) Marshal() []byte {
	stream := core.NewByteStreamWithCapacity(4)

	stream.Write(a.event)

	return stream.Buffer()
}
