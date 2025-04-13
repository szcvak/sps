package messages

import (
	"github.com/szcvak/sps/pkg/config"
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/database"
	"math"
	"time"
)

type BattleEndTrioMessage struct {
	data   BattleEndData
	player *core.Player
	dbm    *database.Manager
}

func NewBattleEndTrioMessage(data BattleEndData, player *core.Player, dbm *database.Manager) *BattleEndTrioMessage {
	return &BattleEndTrioMessage{
		data:   data,
		player: player,
		dbm:    dbm,
	}
}

func (b *BattleEndTrioMessage) PacketId() uint16 {
	return 23456
}

func (b *BattleEndTrioMessage) PacketVersion() uint16 {
	return 1
}

func (b *BattleEndTrioMessage) Marshal() []byte {
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
		trophies = getTrioBattleEndTrophies(int32(b.data.BattleRank), b.player.Brawlers[b.data.Brawlers[0].CharacterId.S].Trophies)
		coins = getTrioBattleEndCoins(int32(b.data.BattleRank))
		exp = getTrioBattleEndExp(int32(b.data.BattleRank))

		starPlayerExp = 10
		boostedCoins = 0

		now := time.Now().Unix()

		if int64(b.player.CoinBooster)-now > 0 {
			boostedCoins = coins
		}
	}

	stream.Write(core.VInt(1))
	stream.Write(core.VInt(0))
	stream.Write(core.VInt(coins))
	stream.Write(core.VInt(6969))
	stream.Write(core.VInt(0))
	stream.Write(false)
	stream.Write(b.data.BattleEndType)

	stream.Write(core.VInt(trophies))
	stream.Write(core.ScId{28, b.player.ProfileIcon})
	stream.Write(b.data.IsTutorial)
	stream.Write(b.data.IsRealGame)
	stream.Write(core.VInt(50))
	stream.Write(core.VInt(boostedCoins))
	stream.Write(core.VInt(0)) /* coins doubled */

	stream.Write(b.data.PlayersAmount)

	for _, player := range b.data.Brawlers {
		stream.Write(player.Name)
		stream.Write(player.IsPlayer)
		stream.Write(player.Team != b.data.Brawlers[0].Team)
		stream.Write(player.IsPlayer) // star player
		stream.Write(core.ScId(player.CharacterId))
		stream.Write(core.ScId(player.SkinId))
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

	if b.player.AllianceId != nil {
		logError(
			b.dbm.Exec(
				"update alliances set total_trophies = total_trophies + $1 where id = $2",
				trophies, *b.player.AllianceId,
			),
		)
	}

	return stream.Buffer()
}

// --- Helper functions --- //

type trophyRangeTrio struct {
	minTrophies int32
	maxTrophies int32

	resultChanges [3]int32
}

var trophyRangesTrio = []trophyRangeTrio{
	{0, 29, [3]int32{6, 0, 0}},
	{30, 59, [3]int32{6, -1, 0}},
	{60, 99, [3]int32{6, -2, 0}},
	{100, 139, [3]int32{5, -2, 0}},
	{140, 219, [3]int32{5, -3, 0}},
	{220, 299, [3]int32{5, -4, 0}},
	{300, 499, [3]int32{5, -5, 0}},
	{500, 599, [3]int32{4, -6, 0}},
	{600, 699, [3]int32{3, -6, 0}},
	{700, 799, [3]int32{3, -7, 0}},
	{800, 899, [3]int32{2, -7, 0}},
	{900, math.MaxInt32, [3]int32{2, -8, 0}},
}

func getTrioBattleEndTrophies(result int32, currentTrophies int32) int32 {
	if result < 0 || result > 2 {
		return 0
	}

	if currentTrophies < 0 {
		currentTrophies = 0
	}

	for _, tr := range trophyRangesTrio {
		if currentTrophies >= tr.minTrophies && currentTrophies <= tr.maxTrophies {
			return tr.resultChanges[result]
		}
	}

	return 0
}

func getTrioBattleEndCoins(result int32) int32 {
	switch result {
	case 0:
		return 20
	case 1:
		return 15
	case 2:
		return 10
	}

	return 0
}

func getTrioBattleEndExp(result int32) int32 {
	switch result {
	case 0:
		return 10
	case 1:
		return 5
	case 2:
		return 0
	}

	return 0
}
