package messages

import (
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
)

type TeamPostAdMessage struct{}

func NewTeamPostAdMessage() *TeamPostAdMessage {
	return &TeamPostAdMessage{}
}

func (t *TeamPostAdMessage) Unmarshal(_ []byte) {}

func (t *TeamPostAdMessage) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	if wrapper.Player.TeamId == nil {
		return
	}

	tm := core.GetTeamManager()
	tm.TogglePostAd(wrapper.Player)

	for _, member := range tm.Teams[*wrapper.Player.TeamId].Members {
		if member.Wrapper != nil {
			msg := NewTeamMessage(member.Wrapper)
			payload := msg.Marshal()
		
			member.Wrapper.Send(msg.PacketId(), msg.PacketVersion(), payload)
		}
	}
}
