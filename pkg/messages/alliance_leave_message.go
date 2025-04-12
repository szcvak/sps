package messages

import (
	"log/slog"
	"context"
	
	"github.com/szcvak/sps/pkg/hub"
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
)

// TODO: abstract alliance join and leave

type AllianceLeaveMessage struct{}

func NewAllianceLeaveMessage() *AllianceLeaveMessage {
	return &AllianceLeaveMessage{}
}

func (a *AllianceLeaveMessage) Unmarshal(_ []byte) {}

func (a *AllianceLeaveMessage) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	if wrapper.Player.AllianceId == nil {
		return
	}

	err := dbm.Exec(
		"update alliances set total_trophies = total_trophies - $1 where id = $2",
		wrapper.Player.Trophies, *wrapper.Player.AllianceId + 1,
	)
	
	if err != nil {
		slog.Error("failed to update alliances!", "err", err)
		return
	}
	
	err = dbm.Exec(
		"delete from alliance_members where player_id = $1",
		wrapper.Player.DbId,
	)
	
	if err != nil {
		slog.Error("failed to update alliance_members!", "err", err)
		return
	}
	
	alliance, err := dbm.LoadAlliance(context.Background(), *wrapper.Player.AllianceId)
	
	if err != nil {
		slog.Error("failed to load alliance!", "err", err)
		return
	}
	
	wrapper.Player.AllianceRole = 0
	wrapper.Player.AllianceId = nil
	
	msg := NewAllianceEventMessage(80)
	wrapper.Send(msg.PacketId(), msg.PacketVersion(), msg.Marshal())
	
	msg2 := NewMyAllianceMessage(wrapper, dbm)
	wrapper.Send(msg2.PacketId(), msg2.PacketVersion(), msg2.Marshal())
	
	if alliance.TotalMembers > 1 {
		message, err := dbm.AddAllianceMessage(context.Background(), *wrapper.Player.AllianceId, wrapper.Player, 4, "", wrapper.Player)
	
		if err != nil {
			slog.Error("failed to send alliance message!", "err", err)
			return
		}
	
		serverMsg := NewAllianceChatServerMessage(*message, *wrapper.Player.AllianceId)
	
		messageHub := hub.GetHub()
		messageHub.BroadcastToAlliance(*wrapper.Player.AllianceId, serverMsg)
	}
}
