package messages

import (
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
	"log/slog"
)

type TeamChangeMemberSettingsMessage struct {
	card core.DataRef
}

func NewTeamChangeMemberSettingsMessage() *TeamChangeMemberSettingsMessage {
	return &TeamChangeMemberSettingsMessage{}
}

func (t *TeamChangeMemberSettingsMessage) Unmarshal(data []byte) {
	stream := core.NewByteStream(data)
	defer stream.Close()

	_, _ = stream.ReadVInt()
	t.card, _ = stream.ReadDataRef()
}

func (t *TeamChangeMemberSettingsMessage) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	if wrapper.Player.TeamId == nil {
		return
	}
	
	slog.Info("changed brawler in team", "h", t.card.F, "l", t.card.S)

	wrapper.Player.SelectedCardHigh = t.card.F
	wrapper.Player.SelectedCardLow = t.card.S
	
	tm := core.GetTeamManager()
	tm.UpdateBrawler(wrapper.Player)
	
	for _, member := range tm.Teams[*wrapper.Player.TeamId].Members {
		if member.Wrapper != nil {
			msg := NewTeamMessage(member.Wrapper)
			payload := msg.Marshal()
		
			member.Wrapper.Send(msg.PacketId(), msg.PacketVersion(), payload)
		}
	}
}
