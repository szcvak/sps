package messages

import (
	"github.com/szcvak/sps/pkg/config"
	"github.com/szcvak/sps/pkg/core"
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
	stream := core.NewByteStreamWithCapacity(32)

	var (
		trophies      int32
		coins         int32
		exp           int32
		starPlayerExp int32
		boostedCoins  int32
	)

	if !b.data.IsRealGame {
		trophies = 0
		coins = 0
		exp = 0
		starPlayerExp = 0
		boostedCoins = 0
	} else {
		trophies = getSdBattleEndTrophies(int32(b.data.BattleRank), b.player.Brawlers[b.data.Brawlers[0].CharacterId.S].Trophies)
		coins = getSdBattleEndCoins(int32(b.data.BattleRank))
		exp = getSdBattleEndExp(int32(b.data.BattleRank))

		starPlayerExp = 10
		boostedCoins = 0

		now := time.Now().Unix()

		if int64(b.player.CoinBooster)-now > 0 {
			boostedCoins = coins
		}
	}

	stream.Write(core.VInt(5)) // battle end mode, 5=showdown, other=3v3
	stream.Write(core.VInt(0)) // 1+= "all coins collected"
	stream.Write(core.VInt(coins))
	stream.Write(core.VInt(6969))
	stream.Write(core.VInt(0))
	stream.Write(false)
	stream.Write(b.data.BattleRank)

	stream.Write(core.VInt(trophies))
	stream.Write(core.ScId{28, b.player.ProfileIcon})
	stream.Write(b.data.IsTutorial)
	stream.Write(b.data.IsRealGame)
	stream.Write(core.VInt(50)) // coin booster %
	stream.Write(core.VInt(boostedCoins))
	stream.Write(0) // coins doubled

	stream.Write(b.data.PlayersAmount)

	for _, player := range b.data.Brawlers {
		stream.Write(player.Name)
		stream.Write(player.IsPlayer)
		stream.Write(player.Team != b.data.Brawlers[0].Team)
		stream.Write(player.IsPlayer) // star player
		stream.Write(core.ScId{player.CharacterId.F, player.CharacterId.S})
		stream.Write(core.ScId{player.SkinId.F, player.SkinId.S})
		stream.Write(core.VInt(0)) // brawler trophies
		stream.Write(core.VInt(player.PowerLevel - 1))
	}

	stream.Write(core.VInt(2))
	stream.Write(core.VInt(0))
	stream.Write(core.VInt(exp))
	stream.Write(core.VInt(8))
	stream.Write(core.VInt(starPlayerExp))

	stream.Write(core.VInt(0))

	currentBrawlerTrophies := b.player.Brawlers[b.data.Brawlers[0].CharacterId.S].Trophies
	currentBrawlerHighest := b.player.Brawlers[b.data.Brawlers[0].CharacterId.S].HighestTrophies

	stream.Write(core.VInt(2))
	stream.Write(core.VInt(1))
	stream.Write(core.VInt(currentBrawlerTrophies))
	stream.Write(core.VInt(currentBrawlerHighest))
	stream.Write(core.VInt(5))
	stream.Write(core.VInt(b.player.Experience))
	stream.Write(core.VInt(b.player.Experience))

	stream.Write(true)

	core.EmbedMilestones(stream)

	b.player.Trophies += trophies
	b.player.Experience += exp
	b.player.Wallet[config.CurrencyCoins].Balance += int64(coins + boostedCoins) /* coins doubled */
	b.player.CoinsReward += coins + boostedCoins
	b.player.Brawlers[b.data.Brawlers[0].CharacterId.S].Trophies += trophies
	b.player.Brawlers[b.data.Brawlers[0].CharacterId.S].HighestTrophies = max(b.player.Brawlers[b.data.Brawlers[0].CharacterId.S].Trophies, b.player.Brawlers[b.data.Brawlers[0].CharacterId.S].HighestTrophies)

	brawler := b.player.Brawlers[b.data.Brawlers[0].CharacterId.S]

	logError(
		b.dbm.Exec(
			"update player_progression set trophies = $1, highest_trophies = $2, experience = $3 where player_id = $4",
			b.player.Trophies, b.player.HighestTrophies, b.player.Experience, b.player.DbId,
		),
	)

	logError(
		b.dbm.Exec(
			"update player_wallet set balance = $1 where player_id = $2 and currency_id = $3",
			b.player.Wallet[config.CurrencyCoins].Balance, b.player.DbId, config.CurrencyCoins,
		),
	)

	logError(
		b.dbm.Exec(
			"update player_brawlers set trophies = $1, highest_trophies = $2 where player_id = $3 and brawler_id = $4",
			brawler.Trophies, brawler.HighestTrophies, b.player.DbId, brawler.BrawlerId,
		),
	)

	logError(
		b.dbm.Exec(
			"update players set coins_reward = $1 where id = $2",
			b.player.CoinsReward, b.player.DbId,
		),
	)

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
