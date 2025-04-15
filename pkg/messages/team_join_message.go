package messages

import (
	"log/slog"
	
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
)

type TeamJoinMessage struct {
	teamHigh core.VInt
	teamId core.VInt
	teamType core.VInt
}

func NewTeamJoinMessage() *TeamJoinMessage {
	return &TeamJoinMessage{}
}

func (t *TeamJoinMessage) Unmarshal(data []byte) {
	stream := core.NewByteStream(data)
	defer stream.Close()

	t.teamHigh, _ = stream.ReadVInt()
	t.teamId, _ = stream.ReadVInt()
	t.teamType, _ = stream.ReadVInt()
}

func (t *TeamJoinMessage) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	if wrapper.Player.TeamId != nil {
		return
	}
	
	tm := core.GetTeamManager()
	_, exists := tm.Teams[int32(t.teamId)]
	
	if !exists {
		slog.Error("player tried to join unknown team!", "playerId", wrapper.Player.DbId, "teamId", t.teamId)
		return
	}
	
	tm.JoinTeam(wrapper, int32(t.teamId))
	
	if wrapper.Player.TeamId == nil {
		return
	}
	
	for _, member := range tm.Teams[*wrapper.Player.TeamId].Members {
		if member.Wrapper != nil {
			msg := NewTeamMessage(member.Wrapper)
			payload := msg.Marshal()
		
			member.Wrapper.Send(msg.PacketId(), msg.PacketVersion(), payload)
		}
	}
	
	msg := NewTeamStreamMessage(wrapper, true)
	wrapper.Send(msg.PacketId(), msg.PacketVersion(), msg.Marshal())
}
