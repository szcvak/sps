package messages

import (
	"log/slog"
	"context"
	
	"github.com/szcvak/sps/pkg/hub"
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
)

type AllianceLeaveMessage struct{}

func NewAllianceLeaveMessage() *AllianceLeaveMessage {
	return &AllianceLeaveMessage{}
}

func (a *AllianceLeaveMessage) Unmarshal(_ []byte) {}

func (a *AllianceLeaveMessage) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	if wrapper.Player.AllianceId == nil {
		return
	}

	oldAlliance := *wrapper.Player.AllianceId
	
	deleted, err := dbm.RemoveAllianceMember(context.Background(), wrapper.Player, oldAlliance)

	if err != nil {
		slog.Error("failed to remove alliance member!", "err", err)
		return
	}
	
	h := hub.GetHub()
	h.UpdateAllianceMembership(wrapper, wrapper.Player.AllianceId, nil)
	
	msg := NewAllianceEventMessage(80)
	wrapper.Send(msg.PacketId(), msg.PacketVersion(), msg.Marshal())
	
	msg2 := NewMyAllianceMessage(wrapper, dbm)
	wrapper.Send(msg2.PacketId(), msg2.PacketVersion(), msg2.Marshal())
	
	if !deleted {
		message, err := dbm.AddAllianceMessage(context.Background(), oldAlliance, wrapper.Player, 44, "", wrapper.Player)
	
		if err != nil {
			slog.Error("failed to send alliance message!", "err", err)
			return
		}
	
		serverMsg := NewAllianceChatServerMessage(*message, oldAlliance)
		h.BroadcastToAlliance(oldAlliance, serverMsg)
	}
}
