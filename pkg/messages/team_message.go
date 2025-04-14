package messages

import (
	"context"
	"github.com/szcvak/sps/pkg/config"
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
	"log/slog"
)

type TeamMessage struct {
	wrapper *core.ClientWrapper
	dbm     *database.Manager
	event   int32
}

func NewTeamMessage(wrapper *core.ClientWrapper, dbm *database.Manager, event int32) *TeamMessage {
	return &TeamMessage{
		wrapper: wrapper,
		dbm:     dbm,
		event:   event,
	}
}

func (t *TeamMessage) PacketId() uint16 {
	return 24124
}

func (t *TeamMessage) PacketVersion() uint16 {
	return 1
}

func (t *TeamMessage) Marshal() []byte {
	stream := core.NewByteStreamWithCapacity(64)

	if t.wrapper.Player.TeamCode == nil {
		return []byte{}
	}

	team, err := t.dbm.LoadTeam(context.Background(), *t.wrapper.Player.TeamCode)

	if err != nil {
		slog.Error("failed to load team!", "err", err)
		return []byte{}
	}

	brawlerTrophies := 0

	if config.MaximumRank <= 34 {
		brawlerTrophies = progressStart[config.MaximumRank-1]
	} else {
		brawlerTrophies = progressStart[33] + 50*(config.MaximumRank-1)
	}

	em := core.GetEventManager()
	location := em.GetCurrentEvent(t.event).LocationId

	stream.Write(core.VInt(0))    // game room type
	stream.Write(team.IsPractice) // practice with bots
	stream.Write(core.VInt(3))    // max players

	stream.Write(0)
	stream.Write(int32(team.DbId)) // team id

	stream.Write(core.VInt(0))
	stream.Write(core.VInt(0))
	stream.Write(core.VInt(0))

	stream.Write(core.ScId{15, location})

	stream.Write(core.VInt(team.TotalMembers))

	for i, member := range team.Members {
		if i == 0 {
			stream.Write(0)
			stream.Write(t.wrapper.Player.LowId)
		} else {
			stream.Write(1)
			stream.Write(int32(i))
		}

		selectedSkin := member.Brawlers[member.SelectedCardLow].SelectedSkinId

		stream.Write(member.Name)
		stream.Write(core.VInt(0))
		stream.Write(core.ScId{member.SelectedCardHigh, member.SelectedCardLow})
		stream.Write(core.ScId{29, selectedSkin})
		stream.Write(core.VInt(brawlerTrophies))
		stream.Write(core.VInt(brawlerTrophies))
		stream.Write(core.VInt(config.MaximumRank))
		stream.Write(core.VInt(member.TeamStatus))
		stream.Write(core.VInt(0))
		stream.Write(member.TeamIsReady)
		stream.Write(core.VInt(i))
	}

	return stream.Buffer()
}
