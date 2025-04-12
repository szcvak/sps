package messages

import (
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
)

type AskForJoinableAlliancesMessage struct{}

func NewAskForJoinableAlliancesMessage() *AskForJoinableAlliancesMessage {
	return &AskForJoinableAlliancesMessage{}
}

func (a *AskForJoinableAlliancesMessage) Unmarshal(_ []byte) {}

func (a *AskForJoinableAlliancesMessage) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	msg := NewJoinableAlliancesMessage(dbm)
	wrapper.Send(msg.PacketId(), msg.PacketVersion(), msg.Marshal())
}
