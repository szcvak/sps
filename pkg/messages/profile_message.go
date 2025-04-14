package messages

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/csv"
	"github.com/szcvak/sps/pkg/database"
)

type ProfileMessage struct {
	player *core.Player
	dbm    *database.Manager
}

func NewProfileMessage(player *core.Player, dbm *database.Manager) *ProfileMessage {
	return &ProfileMessage{
		player: player,
		dbm:    dbm,
	}
}

func (p *ProfileMessage) PacketId() uint16 {
	return 24113
}

func (p *ProfileMessage) PacketVersion() uint16 {
	return 1
}

func (p *ProfileMessage) Marshal() []byte {
	stream := core.NewByteStreamWithCapacity(32)

	stream.Write(core.VInt(p.player.HighId))
	stream.Write(core.VInt(p.player.LowId))

	stream.Write(p.player.Name)

	stream.Write(core.VInt(0))

	stream.Write(core.VInt(len(p.player.Brawlers)))

	for key, data := range p.player.Brawlers {
		stream.Write(core.ScId{16, int32(key)})
		stream.Write(core.VInt(0))
		stream.Write(core.VInt(data.Trophies))
		stream.Write(core.VInt(data.HighestTrophies))

		var powerLevel int32 = 0

		for card, amt := range data.Cards {
			id, err := strconv.Atoi(card)

			if err != nil {
				slog.Error("failed to convert card to int!", "err", err)
				continue
			}

			if !csv.IsCardUnlocked(id) {
				powerLevel += int32(amt)
			}
		}

		stream.Write(core.VInt(powerLevel))
	}

	// stats

	stream.Write(core.VInt(7)) // stats count

	stream.Write(core.VInt(1)) // stats index
	stream.Write(core.VInt(p.player.TrioVictories))

	stream.Write(core.VInt(2))
	stream.Write(core.VInt(p.player.Experience))

	stream.Write(core.VInt(3))
	stream.Write(core.VInt(p.player.Trophies))

	stream.Write(core.VInt(4))
	stream.Write(core.VInt(p.player.HighestTrophies))

	stream.Write(core.VInt(5))
	stream.Write(core.VInt(len(p.player.Brawlers)))

	stream.Write(core.VInt(7))
	stream.Write(core.VInt(28000000 + p.player.ProfileIcon))

	stream.Write(core.VInt(8))
	stream.Write(core.VInt(p.player.SoloVictories))

	// second pass

	stream.Write(core.VInt(2))
	stream.Write(core.VInt(p.player.Experience))

	stream.Write(core.VInt(3))
	stream.Write(core.VInt(p.player.Trophies))

	stream.Write(core.VInt(4))
	stream.Write(core.VInt(p.player.HighestTrophies))

	stream.Write(core.VInt(5))
	stream.Write(core.VInt(len(p.player.Brawlers)))

	stream.Write(core.VInt(7))
	stream.Write(core.VInt(28000000 + p.player.ProfileIcon))

	stream.Write(core.VInt(8))
	stream.Write(core.VInt(p.player.SoloVictories))

	// end

	// alliance

	if p.player.AllianceId == nil {
		stream.Write(false)
		stream.Write(core.VInt(0))
	} else {
		stream.WriteBool(true)

		a, err := p.dbm.LoadAlliance(context.Background(), *p.player.AllianceId)

		if err != nil {
			slog.Error("failed to load alliance!", "err", err)
			return stream.Buffer()
		}

		stream.Write(0)
		stream.Write(int32(a.Id))

		stream.Write(a.Name)
		stream.Write(core.ScId{8, a.BadgeId})

		stream.Write(core.VInt(a.Type))
		stream.Write(core.VInt(a.TotalMembers))
		stream.Write(core.VInt(a.TotalTrophies))
		stream.Write(core.VInt(a.RequiredTrophies))

		stream.Write(core.ScId{14, 249}) // unknown
		stream.Write(core.ScId{25, 2})   // unknown

		stream.Write(core.VInt(p.player.AllianceRole))
	}

	return stream.Buffer()
}
