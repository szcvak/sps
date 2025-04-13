package messages

import (
	"github.com/szcvak/sps/pkg/core"
	"github.com/szcvak/sps/pkg/csv"
	"github.com/szcvak/sps/pkg/database"
	"log/slog"
	"strconv"
)

type AskForBattleEndMessage struct {
	data   BattleEndData
	stream *core.ByteStream
}

type heroEntry struct {
	CharacterId core.DataRef
	SkinId      core.DataRef
	Team        core.VInt
	IsPlayer    bool
	Name        string
	PowerLevel  int32
}

type BattleEndData struct {
	BattleEndType core.VInt
	BattleTime    core.VInt
	BattleRank    core.VInt

	CsvId core.VInt

	Location      core.VInt
	PlayersAmount core.VInt

	Brawlers []heroEntry

	IsRealGame bool
	IsTutorial bool
}

func NewAskForBattleEndMessage() *AskForBattleEndMessage {
	return &AskForBattleEndMessage{}
}

func (a *AskForBattleEndMessage) Unmarshal(data []byte) {
	stream := core.NewByteStream(data)

	a.data = BattleEndData{}

	a.data.BattleEndType, _ = stream.ReadVInt()
	a.data.BattleTime, _ = stream.ReadVInt()
	a.data.BattleRank, _ = stream.ReadVInt()

	a.data.CsvId, _ = stream.ReadVInt()
	a.data.Location, _ = stream.ReadVInt()
	a.data.PlayersAmount, _ = stream.ReadVInt()

	a.data.Brawlers = []heroEntry{}

	for i := 0; i < int(a.data.PlayersAmount); i++ {
		a.data.Brawlers = append(a.data.Brawlers, heroEntry{
			CharacterId: discardError(stream.ReadDataRef()),
			SkinId:      discardError(stream.ReadDataRef()),
			Team:        discardError(stream.ReadVInt()),
			IsPlayer:    discardError(stream.ReadBool()),
			Name:        discardError(stream.ReadString()),
			PowerLevel:  0,
		})
	}

	// deferred loading because player data is not available here
	a.stream = stream
}

func (a *AskForBattleEndMessage) Process(wrapper *core.ClientWrapper, dbm *database.Manager) {
	player := wrapper.Player
	charId := int32(-1)

	playerIndex := -1

	for i, entry := range a.data.Brawlers {
		if entry.IsPlayer {
			charId = entry.CharacterId.S
			playerIndex = i

			break
		}
	}

	if playerIndex == -1 {
		slog.Error("failed to find self in battle end data!", "playerId", player.DbId)
		return
	}

	if brawlerData, ok := player.Brawlers[charId]; ok {
		powerLevel := int32(0)

		for cardStr, amt := range brawlerData.Cards {
			id, err := strconv.Atoi(cardStr)

			if err != nil {
				slog.Error("failed to convert card str!", "card", cardStr, "playerId", player.DbId)
				continue
			}

			if !csv.IsCardUnlocked(id) {
				powerLevel += amt
			}
		}

		a.data.Brawlers[playerIndex].PowerLevel = powerLevel
	} else {
		slog.Warn("player data for brawler not found", "playerId", player.DbId, "charId", charId)
		a.data.Brawlers[playerIndex].PowerLevel = 1
	}

	a.data.IsRealGame = player.TutorialState != 1
	a.data.IsTutorial = !a.data.IsRealGame

	if a.data.BattleRank != 0 {
		msg := NewBattleEndSdMessage(a.data, player, dbm)
		wrapper.Send(msg.PacketId(), msg.PacketVersion(), msg.Marshal())
	} else {
		// msg := NewBattleEndTrioMessage(a.data, player, dbm)
		// wrapper.Send(msg.PacketId(), msg.PacketVersion(), msg.Marshal())
	}
}

// --- Helper functions --- //

func discardError[T any](value T, _ error) T {
	return value
}
