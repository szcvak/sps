package messages

import (
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
)

type AskForAllianceDataMessage struct {
	id int32
}

func NewAskForAllianceDataMessage() *AskForAllianceDataMessage {
	return &AskForAllianceDataMessage{}
}

func (a *AskForAllianceDataMessage) Unmarshal(data []byte) {
	stream := core.NewByteStream(data)
	
	_, _ = stream.ReadInt()
	a.id, _ = stream.ReadInt()
}

func (a *AskForAllianceDataMessage) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	msg := NewAllianceDataMessage(wrapper, dbm, int64(a.id))
	wrapper.Send(msg.PacketId(), msg.PacketVersion(), msg.Marshal())
}
