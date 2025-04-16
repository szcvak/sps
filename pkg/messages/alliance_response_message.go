package messages

import "github.com/szcvak/sps/pkg/core"

type AllianceResponseMessage struct {
	response int32
}

func NewAllianceResponseMessage(response int32) *AllianceResponseMessage {
	return &AllianceResponseMessage {
		response: response,
	}
}

func (a *AllianceResponseMessage) PacketId() uint16 {
	return 24333
}

func (a *AllianceResponseMessage) PacketVersion() uint16 {
	return 1
}

func (a *AllianceResponseMessage) Marshal() []byte {
	stream := core.NewByteStreamWithCapacity(4)
	
	stream.Write(core.VInt(a.response))
	
	return stream.Buffer()
}
