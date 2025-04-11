package messages

import (
	"strconv"
	"log/slog"
	
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/csv"
)

type ProfileMessage struct {
	player *core.Player
}

func NewProfileMessage(player *core.Player) *ProfileMessage {
	return &ProfileMessage {
		player: player,
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
		
		stream.Write(core.VInt(powerLevel - 1))
	}
	
	// stats
	
	stream.Write(core.VInt(7)) // stats count
	
	stream.Write(core.VInt(1)) // stats index
	stream.Write(core.VInt(p.player.TrioVictories))
	
	stream.Write(core.VInt(2))
	stream.Write(core.VInt(0)) // experience
	
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
	stream.Write(core.VInt(0)) // experience
	
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
	
	stream.Write(true) // is in alliance
	stream.Write(0) // high id
	stream.Write(1) // low id
	stream.Write("The best") // name
	stream.Write(core.ScId{8, 19}) // icon
	
	stream.Write(core.VInt(1)) // type
	stream.Write(core.VInt(1)) // members
	stream.Write(core.VInt(0)) // trophies
	stream.Write(core.VInt(0)) // required trophies
	
	stream.Write(core.ScId{14, 249}) // unknown
	stream.Write(core.ScId{25, 2}) // unknown
	
	return stream.Buffer()
}
