package messages

import "github.com/szcvak/sps/pkg/core"

type TeamLeftMessage struct {
	reason int32
}

func NewTeamLeftMessage(reason int32) *TeamLeftMessage {
	return &TeamLeftMessage {
		reason: reason,
	}
}

func (t *TeamLeftMessage) PacketId() uint16 {
	return 24125
}

func (t *TeamLeftMessage) PacketVersion() uint16 {
	return 1
}

func (t *TeamLeftMessage) Marshal() []byte {
	stream := core.NewByteStreamWithCapacity(4)
	
	stream.Write(t.reason)
	
	return stream.Buffer()
}
