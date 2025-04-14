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
	slog.Info("changed card", "h", t.card.F, "l", t.card.S)

	wrapper.Player.SelectedCardHigh = t.card.F
	wrapper.Player.SelectedCardLow = t.card.S

	err := dbm.Exec("update players set selected_card_high = $1, selected_card_low = $2 where id = $3", t.card.F, t.card.S, wrapper.Player.DbId)

	if err != nil {
		slog.Error("failed to update player's card!", "err", err)
		return
	}

	msg := NewTeamMessage(wrapper, dbm, 2)
	wrapper.Send(msg.PacketId(), msg.PacketVersion(), msg.Marshal())
}
