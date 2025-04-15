package messages

import (
	"github.com/szcvak/sps/pkg/config"
	"github.com/szcvak/sps/pkg/core"
	"log/slog"
)

type TeamMessage struct {
	wrapper *core.ClientWrapper
}

func NewTeamMessage(wrapper *core.ClientWrapper) *TeamMessage {
	return &TeamMessage{
		wrapper: wrapper,
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

	if t.wrapper == nil {
		return []byte{}
	}
	
	if t.wrapper.Player == nil {
		return []byte{}
	}
	
	if t.wrapper.Player.TeamId == nil {
		slog.Error("failed to send team message!", "playerId", t.wrapper.Player.DbId, "err", "player is not in team")
		return []byte{}
	}

	tm := core.GetTeamManager()
	team := tm.Teams[*t.wrapper.Player.TeamId]

	if team == nil {
		slog.Error("team does not exist!", "teamId", *t.wrapper.Player.TeamId)
		t.wrapper.Player.TeamId = nil
		return []byte{}
	}

	brawlerTrophies := 0

	if config.MaximumRank <= 34 {
		brawlerTrophies = progressStart[config.MaximumRank-1]
	} else {
		brawlerTrophies = progressStart[33] + 50*(config.MaximumRank-1)
	}

	em := core.GetEventManager()
	event := em.GetCurrentEvent(team.Event-1)
	location := event.LocationId
	maxPlayers := event.Config.MaxPlayers

	stream.Write(core.VInt(0))           // game room type
	stream.Write(team.IsPractice)        // practice with bots
	stream.Write(core.VInt(maxPlayers))  // max players

	stream.Write(0)
	stream.Write(int32(team.Id)) // team id

	stream.Write(core.VInt(0))
	stream.Write(core.VInt(0))
	stream.Write(core.VInt(0))

	stream.Write(core.ScId{15, location})

	stream.Write(core.VInt(len(team.Members)))

	for i, member := range team.Members {
		stream.Write(member.HighId)
		stream.Write(member.LowId)

		stream.Write(member.Name)
		stream.Write(core.VInt(0))
		stream.Write(member.SelectedBrawler)
		stream.Write(member.SelectedSkin)
		stream.Write(core.VInt(brawlerTrophies))
		stream.Write(core.VInt(brawlerTrophies))
		stream.Write(core.VInt(config.MaximumRank))
		stream.Write(core.VInt(member.Status)) // 0=offline, 1=in battle, 2=other screen, 3=present, 4=in matchmake
		stream.Write(core.VInt(0))
		stream.Write(member.IsReady)
		stream.Write(core.VInt(i))
	}

	return stream.Buffer()
}
