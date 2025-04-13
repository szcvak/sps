package messages

import (
	"context"
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
	"log/slog"
	"time"
)

type LeaderboardMessage struct {
	wrapper *core.ClientWrapper
	dbm     *database.Manager
	g       *GetLeaderboardMessage
}

func NewLeaderboardMessage(wrapper *core.ClientWrapper, dbm *database.Manager, g *GetLeaderboardMessage) *LeaderboardMessage {
	return &LeaderboardMessage{
		wrapper: wrapper,
		dbm:     dbm,
		g:       g,
	}
}

func (l *LeaderboardMessage) PacketId() uint16 {
	return 24403
}

func (l *LeaderboardMessage) PacketVersion() uint16 {
	return 1
}

func (l *LeaderboardMessage) Marshal() []byte {
	stream := core.NewByteStreamWithCapacity(512)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	const limit = 200

	playerRank := int32(0)

	stream.Write(l.g.leaderboardType)

	if l.g.leaderboardType == 0 {
		stream.Write(core.ScId(l.g.brawler))
	} else {
		stream.Write(core.VInt(0))
	}

	if l.g.local {
		stream.Write(l.wrapper.Player.Region)
	} else {
		stream.Write(core.EmptyString)
	}

	switch l.g.leaderboardType {
	case 0:
		entries, err := l.dbm.GetBrawlerTrophyLeaderboard(ctx, l.g.brawler.S, limit)

		if err != nil {
			slog.Error("failed to fetch brawler leaderboard!", "error", err, "brawlerID", l.g.brawler.S)
			stream.Write(core.VInt(0))
		} else {
			if l.g.local {
				filtered := make([]database.LeaderboardPlayerEntry, 0, len(entries))

				for _, entry := range entries {
					if entry.Region == l.wrapper.Player.Region {
						filtered = append(filtered, entry)
					}
				}

				entries = filtered
			}

			stream.Write(core.VInt(len(entries)))

			for i, entry := range entries {
				if entry.DbId == l.wrapper.Player.DbId {
					playerRank = int32(i + 1)
				}

				stream.Write(core.VInt(entry.PlayerHighID))
				stream.Write(core.VInt(entry.PlayerLowID))
				stream.Write(core.VInt(i + 1))
				stream.Write(core.VInt(entry.Trophies))
				stream.Write(true)

				stream.Write(entry.Name)
				stream.Write("Test")

				playerLevel := int32(1)
				stream.Write(core.VInt(playerLevel))

				stream.Write(core.ScId{F: 28, S: entry.ProfileIcon})
			}
		}
	case 1:
		entries, err := l.dbm.GetPlayerTrophyLeaderboard(ctx, limit)

		if err != nil {
			slog.Error("failed to fetch player leaderboard!", "error", err)
			stream.Write(core.VInt(0))
		} else {
			if l.g.local {
				filtered := make([]database.LeaderboardPlayerEntry, 0, len(entries))

				for _, entry := range entries {
					if entry.Region == l.wrapper.Player.Region {
						filtered = append(filtered, entry)
					}
				}

				entries = filtered
			}

			stream.Write(core.VInt(len(entries)))

			for i, entry := range entries {
				if entry.DbId == l.wrapper.Player.DbId {
					playerRank = int32(i + 1)
				}

				stream.Write(core.VInt(entry.PlayerHighID))
				stream.Write(core.VInt(entry.PlayerLowID))
				stream.Write(core.VInt(i + 1))
				stream.Write(core.VInt(entry.Trophies))
				stream.Write(true)

				stream.Write(entry.Name)
				stream.Write("Test")

				playerLevel := int32(1)
				stream.Write(core.VInt(playerLevel))

				stream.Write(core.ScId{F: 28, S: entry.ProfileIcon})
			}
		}
	}

	stream.Write(core.VInt(0))
	stream.Write(core.VInt(playerRank))
	stream.Write(core.VInt(0))

	stream.Write(core.VInt(0))
	stream.Write(l.wrapper.Player.Region)

	return stream.Buffer()
}

// --- Helper functions --- //

func getPlayerLevel(experience int32) int32 {
	level := int32(0)

	for i, reqExp := range core.RequiredExp {
		if experience >= int32(reqExp) {
			level = int32(i + 1)
		} else {
			break
		}
	}

	return level
}
