package messages

import (
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
)

type TeamKickMessage struct {
	highId core.VInt
	lowId core.VInt
}

func NewTeamKickMessage() *TeamKickMessage {
	return &TeamKickMessage{}
}

func (t *TeamKickMessage) Unmarshal(data []byte) {
	stream := core.NewByteStream(data)
	defer stream.Close()
	
	t.highId, _ = stream.ReadVInt()
	t.lowId, _ = stream.ReadVInt()
}

func (t *TeamKickMessage) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	if wrapper.Player.TeamId == nil {
		return
	}
	
	if int32(t.highId) == wrapper.Player.HighId && int32(t.lowId) == wrapper.Player.LowId {
		return
	}
	
	oldId := *wrapper.Player.TeamId
	
	tm := core.GetTeamManager()
	w := tm.Kick(wrapper.Player, int32(t.highId), int32(t.lowId))
	
	if w != nil {
		msg := NewTeamLeftMessage(core.TeamLeftReasonKicked)
		w.Send(msg.PacketId(), msg.PacketVersion(), msg.Marshal())
		
		tm.AddMessageExtra(wrapper.Player, 4, 1, w.Player.Name, int32(t.lowId))
	}
	
	for _, member := range tm.Teams[oldId].Members {
		if member.Wrapper != nil {
			msg := NewTeamMessage(member.Wrapper)
			payload := msg.Marshal()
		
			member.Wrapper.Send(msg.PacketId(), msg.PacketVersion(), payload)
			
			if w != nil {
				msg2 := NewTeamStreamMessage(member.Wrapper, false)
				member.Wrapper.Send(msg2.PacketId(), msg2.PacketVersion(), msg2.Marshal())
			}
		}
	}
}
