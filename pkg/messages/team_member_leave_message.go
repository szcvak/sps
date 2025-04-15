package messages

import (
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
)

type TeamMemberLeaveMessage struct{}

func NewTeamMemberLeaveMessage() *TeamMemberLeaveMessage {
	return &TeamMemberLeaveMessage{}
}

func (t *TeamMemberLeaveMessage) Unmarshal(_ []byte) {}

func (t *TeamMemberLeaveMessage) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	if wrapper.Player.TeamId == nil {
		return
	}
	
	oldId := *wrapper.Player.TeamId
	
	tm := core.GetTeamManager()
	tm.LeaveTeam(wrapper.Player)
	
	msg := NewTeamLeftMessage(core.TeamLeftReasonLeft)
	wrapper.Send(msg.PacketId(), msg.PacketVersion(), msg.Marshal())
	
	data := tm.Teams[oldId]
	
	if data == nil {
		return
	}
	
	for _, member := range tm.Teams[oldId].Members {
		if member.Wrapper != nil {
			msg := NewTeamMessage(member.Wrapper)
			payload := msg.Marshal()
		
			member.Wrapper.Send(msg.PacketId(), msg.PacketVersion(), payload)
		}
	}
}
