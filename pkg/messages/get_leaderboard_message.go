package messages

import (
	"context"
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

	g.local, _ = stream.ReadBool()
	g.leaderboardType, _ = stream.ReadVInt()
	g.brawler, _ = stream.ReadDataRef()
}

func (g *GetLeaderboardMessage) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	if wrapper.Player.DbId <= 0 || wrapper.Player.State() != core.StateLoggedIn {
		return
	}

	slog.Info("fetching leaderboard", "type", g.leaderboardType, "isLocal", g.local, "brawler", g.brawler)

	switch g.leaderboardType {
	case core.VInt(0): // brawler
		brawlerLeaderboard(wrapper, dbm, g)
	case core.VInt(1): // players
		playerLeaderboard(wrapper, dbm, g)
	}
}

func brawlerLeaderboard(wrapper *core.ClientWrapper, dbm *database.Manager, g *GetLeaderboardMessage) {
	instanceId := g.brawler.S
	entries, err := dbm.GetBrawlerTrophyLeaderboard(context.Background(), instanceId, 200)

	if err != nil {
		slog.Error("failed to get brawler leaderboard!", "error", err, "brawlerId", instanceId)
		return
	}

	if g.local {
		filtered := make([]database.LeaderboardPlayerEntry, 0, len(entries))
		region := wrapper.Player.Region

		for _, entry := range entries {
			if entry.Region == region {
				filtered = append(filtered, entry)
			}
		}

		entries = filtered
	}

	msg := NewLeaderboardMessage(wrapper, dbm, g)
	wrapper.Send(msg.PacketId(), msg.PacketVersion(), msg.Marshal())
}

func playerLeaderboard(wrapper *core.ClientWrapper, dbm *database.Manager, g *GetLeaderboardMessage) {
	entries, err := dbm.GetPlayerTrophyLeaderboard(context.Background(), 200)

	if err != nil {
		slog.Error("failed to get player leaderboard!", "error", err)
		return
	}

	if g.local {
		filtered := make([]database.LeaderboardPlayerEntry, 0, len(entries))
		region := wrapper.Player.Region

		for _, entry := range entries {
			if entry.Region == region {
				filtered = append(filtered, entry)
			}
		}

		entries = filtered
	}

	msg := NewLeaderboardMessage(wrapper, dbm, g)
	wrapper.Send(msg.PacketId(), msg.PacketVersion(), msg.Marshal())
}
