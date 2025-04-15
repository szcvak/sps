package messages

import (
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
)

type TeamMemberStatusMessage struct {
	status core.VInt
}

func NewTeamMemberStatusMessage() *TeamMemberStatusMessage {
	return &TeamMemberStatusMessage{}
}

func (t *TeamMemberStatusMessage) Unmarshal(data []byte) {
	stream := core.NewByteStream(data)
	defer stream.Close()

	t.status, _ = stream.ReadVInt()
}

func (t *TeamMemberStatusMessage) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	if wrapper.Player.TeamId == nil {
		return
	}

	tm := core.GetTeamManager()
	tm.SetStatus(wrapper.Player, int16(t.status))
	
	for _, member := range tm.Teams[*wrapper.Player.TeamId].Members {
		if member.Wrapper != nil {
			msg := NewTeamMessage(member.Wrapper)
			payload := msg.Marshal()
		
			member.Wrapper.Send(msg.PacketId(), msg.PacketVersion(), payload)
		}
	}
}
