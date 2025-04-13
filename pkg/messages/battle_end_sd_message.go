package messages

import (
	"github.com/szcvak/sps/pkg/config"
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/csv"
	"github.com/szcvak/sps/pkg/database"
	"log/slog"
	"math"
	"time"
)

type BattleEndSdMessage struct {
	data   BattleEndData
	player *core.Player
	dbm    *database.Manager
}

func NewBattleEndSdMessage(data BattleEndData, player *core.Player, dbm *database.Manager) *BattleEndSdMessage {
	return &BattleEndSdMessage{
		data:   data,
		player: player,
		dbm:    dbm,
	}
}

func (b *BattleEndSdMessage) PacketId() uint16 {
	return 23456
}

func (b *BattleEndSdMessage) PacketVersion() uint16 {
	return 1
}

func (b *BattleEndSdMessage) Marshal() []byte {
	stream := core.NewByteStreamWithCapacity(256)

	player := b.player
	data := b.data

	playerIndex := -1
	charId := int32(-1)

	for i, entry := range data.Brawlers {
		if entry.IsPlayer {
			playerIndex = i
			charId = entry.CharacterId.S

			break
		}
	}

	var playerBrawler *core.PlayerBrawler

	if playerIndex != -1 {
		if pbData, ok := player.Brawlers[charId]; ok {
			playerBrawler = pbData
		}
	}

	if playerBrawler == nil {
		slog.Error("failed to find player brawler data", "playerId", player.DbId, "charId", charId)
		playerBrawler = &core.PlayerBrawler{Trophies: 0, HighestTrophies: 0}
	}

	var (
		trophies      int32 = 0
		coins         int32 = 0
		exp           int32 = 0
		starPlayerExp int32 = 0
		boostedCoins  int32 = 0
		doubledCoins  int32 = 0
	)

	if data.IsRealGame {
		trophies = getSdBattleEndTrophies(int32(data.BattleRank), playerBrawler.Trophies)
		coins = getSdBattleEndCoins(int32(data.BattleRank))
		exp = getSdBattleEndExp(int32(data.BattleRank))

		if playerIndex != -1 && data.Brawlers[playerIndex].IsPlayer {
			starPlayerExp = 10
		}

		doubledCoins = coins

		if coins > player.CoinDoubler {
			doubledCoins = player.CoinDoubler
		}

		now := time.Now().Unix()

		if int64(player.CoinBooster)-now > 0 {
			boostedCoins = coins
		}
	}

	stream.Write(core.VInt(5)) // 5 = showdown
	stream.Write(core.VInt(0))
	stream.Write(core.VInt(coins))
	stream.Write(core.VInt(6969))
	stream.Write(core.VInt(0))
	stream.Write(false)
	stream.Write(data.BattleRank)

	stream.Write(core.VInt(trophies))
	stream.Write(core.ScId{28, player.ProfileIcon})
	stream.Write(data.IsTutorial)
	stream.Write(data.IsRealGame)
	stream.Write(core.VInt(50))
	stream.Write(core.VInt(boostedCoins))
	stream.Write(core.VInt(doubledCoins))

	stream.Write(core.VInt(data.PlayersAmount))

	for _, pData := range data.Brawlers {
		stream.Write(pData.Name)
		stream.Write(pData.IsPlayer)

		isEnemy := false

		if playerIndex != -1 && pData.Team != data.Brawlers[playerIndex].Team {
			isEnemy = true
		}

		stream.Write(isEnemy)

		isStarPlayer := pData.IsPlayer

		stream.Write(isStarPlayer)

		charId := pData.CharacterId.S
		cardId, found := csv.GetCardForCharacter(charId)

		if !found {
			slog.Warn("failed to find card id", "charId", charId)
			stream.Write(core.ScId{16, 0})
		} else {
			stream.Write(core.ScId{16, cardId})
		}

		stream.Write(core.ScId{pData.SkinId.F, pData.SkinId.S})
		stream.Write(core.VInt(0))

		powerLevel := int32(0)

		if pData.IsPlayer {
			powerLevel = pData.PowerLevel - 1

			if powerLevel < 0 {
				powerLevel = 0
			}
		}

		stream.Write(core.VInt(powerLevel))
	}

	stream.Write(core.VInt(2))
	stream.Write(core.VInt(0))
	stream.Write(core.VInt(exp))
	stream.Write(core.VInt(8))
	stream.Write(core.VInt(starPlayerExp))

	stream.Write(core.VInt(0))

	stream.Write(core.VInt(2))
	stream.Write(core.VInt(1))
	stream.Write(core.VInt(playerBrawler.Trophies))
	stream.Write(core.VInt(playerBrawler.HighestTrophies))
	stream.Write(core.VInt(5))
	stream.Write(core.VInt(player.Experience))
	stream.Write(core.VInt(player.Experience))

	stream.Write(true)
	core.EmbedMilestones(stream)

	if data.IsRealGame {
		player.Trophies += trophies
		player.HighestTrophies = max(player.Trophies, player.HighestTrophies)
		player.Experience += exp + starPlayerExp

		if walletCoin, ok := player.Wallet[config.CurrencyCoins]; ok {
			walletCoin.Balance += int64(coins + boostedCoins + doubledCoins)
		}

		player.CoinsReward = coins + boostedCoins + doubledCoins

		if charId < 0 {
			playerBrawler.Trophies += trophies
			playerBrawler.HighestTrophies = max(playerBrawler.Trophies, playerBrawler.HighestTrophies)

			logError(
				b.dbm.Exec(
					"update player_progression set trophies = $1, highest_trophies = $2, experience = $3 where player_id = $4",
					player.Trophies, player.HighestTrophies, player.Experience, player.DbId,
				),
			)

			if walletCoin, ok := player.Wallet[config.CurrencyCoins]; ok {
				logError(
					b.dbm.Exec(
						"update player_wallet set balance = $1 where player_id = $2 and currency_id = $3",
						walletCoin.Balance, player.DbId, config.CurrencyCoins,
					),
				)
			}

			logError(
				b.dbm.Exec(
					"update player_brawlers set trophies = $1, highest_trophies = $2 where player_id = $3 and brawler_id = $4",
					playerBrawler.Trophies, playerBrawler.HighestTrophies, player.DbId, playerBrawler.BrawlerId,
				),
			)

			logError(
				b.dbm.Exec(
					"update players set coin_doubler = $1, coins_reward = $2 where id = $3",
					player.CoinDoubler, player.CoinsReward, player.DbId,
				),
			)

			if player.AllianceId != nil {
				logError(
					b.dbm.Exec(
						"update alliances set total_trophies = total_trophies + $1 where id = $2",
						trophies, *player.AllianceId,
					),
				)
			}
		}
	}

	return stream.Buffer()
}

// --- Helper functions --- //

func logError(err error) {
	if err == nil {
		return
	}

	slog.Error("failed to update data!", "err", err)
}

type trophyRangeSd struct {
	minTrophies int32
	maxTrophies int32

	rankChanges [10]int32
}

var trophyRangesSd = []trophyRangeSd{
	{0, 29, [10]int32{8, 7, 6, 5, 5, 4, 3, 2, 1, 0}},
	{30, 59, [10]int32{8, 7, 6, 4, 3, 2, 0, -1, -2, -4}},
	{60, 99, [10]int32{8, 7, 5, 4, 3, 1, 0, -2, -3, -4}},
	{100, 139, [10]int32{7, 6, 5, 3, 2, 1, -1, -2, -3, -4}},
	{140, 219, [10]int32{7, 6, 4, 3, 2, 0, -1, -3, -4, -5}},
	{220, 299, [10]int32{7, 6, 4, 3, 1, 0, -2, -3, -5, -6}},
	{300, 419, [10]int32{7, 6, 4, 2, 1, -1, -2, -4, -6, -7}},
	{420, 499, [10]int32{5, 4, 3, 1, 0, -2, -3, -4, -6, -7}},
	{500, 599, [10]int32{5, 3, 2, 1, -1, -2, -3, -4, -6, -7}},
	{600, 699, [10]int32{4, 3, 1, 0, -1, -1, -3, -5, -6, -7}},
	{700, 799, [10]int32{4, 2, 1, -1, -2, -3, -4, -5, -6, -8}},
	{800, 899, [10]int32{3, 2, 0, -1, -2, -3, -4, -5, -7, -8}},
	{900, math.MaxInt32, [10]int32{3, 1, -1, -2, -3, -4, -5, -6, -7, -8}},
}

func getSdBattleEndTrophies(rank int32, currentTrophies int32) int32 {
	if rank < 1 || rank > 10 {
		return 0
	}

	if currentTrophies < 0 {
		currentTrophies = 0
	}

	for _, tr := range trophyRangesSd {
		if currentTrophies >= tr.minTrophies && currentTrophies <= tr.maxTrophies {
			return tr.rankChanges[rank-1]
		}
	}

	return 0
}

func getSdBattleEndCoins(rank int32) int32 {
	switch rank {
	case 1:
		return 28
	case 2:
		return 22
	case 3:
		return 16
	case 4:
		return 12
	case 5:
		return 8
	case 6:
		return 6
	case 7:
		return 4
	case 8:
		return 2
	case 9:
		return 1
	case 10:
		return 1
	case 0:
		return 34
	}

	return 0
}

func getSdBattleEndExp(rank int32) int32 {
	switch rank {
	case 1:
		return 12
	case 2:
		return 9
	case 3:
		return 6
	case 4:
		return 5
	case 5:
		return 4
	case 6:
		return 2
	case 7:
		return 1
	case 8:
		return 0
	case 9:
		return 0
	case 10:
		return 0
	case 0:
		return 15
	}

	return 0
}
