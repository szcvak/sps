package messages

import (
	"log/slog"

	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
)

type GetLeaderboardMessage struct {
	local           bool
	leaderboardType core.VInt
	brawler         core.DataRef
}

func NewGetLeaderboardMessage() *GetLeaderboardMessage {
	return &GetLeaderboardMessage{}
}

func (g *GetLeaderboardMessage) Unmarshal(data []byte) {
	stream := core.NewByteStream(data)

	local, _ := stream.ReadVInt()
	g.leaderboardType, _ = stream.ReadVInt()
	g.brawler, _ = stream.ReadDataRef()

	g.local = local == 1
}

func (g *GetLeaderboardMessage) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	if wrapper.Player.DbId <= 0 || wrapper.Player.State() != core.StateLoggedIn {
		return
	}

	slog.Info("fetching leaderboard", "type", g.leaderboardType, "isLocal", g.local, "brawler", g.brawler)

	msg := NewLeaderboardMessage(wrapper, dbm, g)
	wrapper.Send(msg.PacketId(), msg.PacketVersion(), msg.Marshal())
}
