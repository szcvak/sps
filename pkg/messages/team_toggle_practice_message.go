package messages

import (
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
)

type TeamTogglePracticeMessage struct{}

func NewTeamTogglePracticeMessage() *TeamTogglePracticeMessage {
	return &TeamTogglePracticeMessage{}
}

func (t *TeamTogglePracticeMessage) Unmarshal(_ []byte) {}

func (t *TeamTogglePracticeMessage) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	if wrapper.Player.TeamId == nil {
		return
	}

	tm := core.GetTeamManager()
	tm.TogglePractice(wrapper.Player)

	for _, member := range tm.Teams[*wrapper.Player.TeamId].Members {
		if member.Wrapper != nil {
			msg := NewTeamMessage(member.Wrapper)
			payload := msg.Marshal()
		
			member.Wrapper.Send(msg.PacketId(), msg.PacketVersion(), payload)
		}
	}
}
