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
	stream := core.NewByteStreamWithCapacity(64)

	switch l.g.leaderboardType {
	case 0:
		l.brawlerLeaderboard(stream)
	case 1:
		l.playerLeaderboard(stream)
	case 2:
		l.allianceLeaderboard(stream)
	default:
		slog.Error("received unhandled leaderboard type!", "type", l.g.leaderboardType)
	}
	
	return stream.Buffer()
}

// --- Helper functions --- //

func (l *LeaderboardMessage) playerLeaderboard(stream *core.ByteStream) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	var region *string = new(string)
	
	if l.g.local {
		*region = l.wrapper.Player.Region
	} else {
		region = nil
	}
	
	entries, err := l.dbm.GetPlayerTrophyLeaderboard(ctx, 50, region)
	
	if err != nil {
		slog.Error("failed to fetch player leaderboard!", "err", err)
		return
	}
	
	alliances := make(map[int64]*core.Alliance)
	playerIdx := 0
	
	stream.Write(core.VInt(1))
	stream.Write(core.VInt(0))
	
	if l.g.local {
		stream.Write(l.wrapper.Player.Region)
	} else {
		stream.Write(core.EmptyString)
	}
	
	stream.WriteVInt(core.VInt(len(entries)))
	
	for i, entry := range entries {
		if entry.PlayerHighId == l.wrapper.Player.HighId && entry.PlayerLowId == l.wrapper.Player.LowId {
			playerIdx = i+1
		}
		
		stream.Write(core.VInt(entry.PlayerHighId))
		stream.Write(core.VInt(entry.PlayerLowId))
		
		stream.Write(core.VInt(1))
		
		stream.Write(core.VInt(entry.Trophies))
		
		stream.Write(true)
		stream.Write(entry.Name)
		
		if entry.AllianceId != nil {
			alliance, exists := alliances[*entry.AllianceId]
			
			if !exists {
				data, err := l.dbm.LoadAlliance(ctx, *entry.AllianceId)
				
				if err != nil {
					slog.Error("failed to get alliance!", "err", err)
					stream.Write("")
				} else {
					alliances[*entry.AllianceId] = data
					alliance = data
				}
			}
			
			if alliance != nil {
				stream.Write(alliance.Name)
			}
		} else {
			stream.Write("")
		}
		
		stream.Write(core.VInt(getPlayerLevel(entry.PlayerExperience)))
		stream.Write(core.DataRef{28, entry.ProfileIcon})
		stream.Write(false)
	}
	
	stream.Write(core.VInt(0))
	stream.Write(core.VInt(playerIdx))
	stream.Write(core.VInt(0))
	stream.Write(core.VInt(0))
	
	stream.Write(l.wrapper.Player.Region)
}

func (l *LeaderboardMessage) brawlerLeaderboard(stream *core.ByteStream) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	var region *string = new(string)
	
	if l.g.local {
		*region = l.wrapper.Player.Region
	} else {
		region = nil
	}
	
	entries, err := l.dbm.GetBrawlerTrophyLeaderboard(ctx, l.g.brawler.S, 50, region)
	
	if err != nil {
		slog.Error("failed to fetch brawler leaderboard!", "err", err)
		return
	}
	
	alliances := make(map[int64]*core.Alliance)
	playerIdx := 0
	
	stream.Write(core.VInt(0))
	stream.Write(core.DataRef{16, l.g.brawler.S})
	
	if l.g.local {
		stream.Write(l.wrapper.Player.Region)
	} else {
		stream.Write(core.EmptyString)
	}
	
	stream.WriteVInt(core.VInt(len(entries)))
	
	for i, entry := range entries {
		if entry.PlayerHighId == l.wrapper.Player.HighId && entry.PlayerLowId == l.wrapper.Player.LowId {
			playerIdx = i+1
		}
		
		stream.Write(core.VInt(entry.PlayerHighId))
		stream.Write(core.VInt(entry.PlayerLowId))
		
		stream.Write(core.VInt(1))
		
		stream.Write(core.VInt(entry.Trophies))
		
		stream.Write(true)
		stream.Write(entry.Name)
		
		if entry.AllianceId != nil {
			alliance, exists := alliances[*entry.AllianceId]
			
			if !exists {
				data, err := l.dbm.LoadAlliance(ctx, *entry.AllianceId)
				
				if err != nil {
					slog.Error("failed to get alliance!", "err", err)
					stream.Write("")
				} else {
					alliances[*entry.AllianceId] = data
					alliance = data
				}
			}
			
			if alliance != nil {
				stream.Write(alliance.Name)
			}
		} else {
			stream.Write("")
		}
		
		stream.Write(core.VInt(getPlayerLevel(entry.PlayerExperience)))
		stream.Write(core.DataRef{28, entry.ProfileIcon})
		stream.Write(false)
	}
	
	stream.Write(core.VInt(0))
	stream.Write(core.VInt(playerIdx))
	stream.Write(core.VInt(0))
	stream.Write(core.VInt(0))
	
	stream.Write(l.wrapper.Player.Region)
}

func (l *LeaderboardMessage) allianceLeaderboard(stream *core.ByteStream) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	var region *string = new(string)
	
	if l.g.local {
		*region = l.wrapper.Player.Region
	} else {
		region = nil
	}
	
	entries, err := l.dbm.GetAllianceTrophyLeaderboard(ctx, 50, region)
	
	if err != nil {
		slog.Error("failed to fetch alliance leaderboard!", "err", err)
		return
	}
	
	stream.Write(core.VInt(2))
	stream.Write(core.VInt(0))
	
	if l.g.local {
		stream.Write(l.wrapper.Player.Region)
	} else {
		stream.Write(core.EmptyString)
	}
	
	stream.WriteVInt(core.VInt(len(entries)))
	
	for _, entry := range entries {
		stream.Write(core.VInt(0))
		stream.Write(core.VInt(entry.DbId))
		
		stream.Write(core.VInt(1))
		
		stream.Write(core.VInt(entry.TotalTrophies))
		
		stream.Write(false)
		stream.Write(true)
		
		stream.Write(entry.Name)
		stream.Write(core.VInt(entry.TotalMembers))
		stream.Write(core.DataRef{8, entry.BadgeId})
	}
	
	stream.Write(core.VInt(0))
	stream.Write(core.VInt(0))
	stream.Write(core.VInt(0))
	stream.Write(core.VInt(0))
	
	stream.Write(l.wrapper.Player.Region)
}

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
