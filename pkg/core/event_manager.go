package core

import (
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"github.com/szcvak/sps/pkg/csv"
)

const (
	GameModeShowdown  string = "BattleRoyale"
	GameModeBounty    string = "BountyHunter"
	GameModeGemGrab   string = "CoinRush"
	GameModeHeist     string = "AttackDefend"
	GameModeBrawlBall string = "LaserBall"

	NumEventSlots = 4
)

type EventConfig struct {
	Gamemode         string
	RequiredBrawlers int32
	CoinsToClaim     int32
	BonusCoins       int32
	CoinsToWin       int32
	EventText        string
	DoubleCoins      bool
	DoubleExp        bool
}

type EventSlotSchedule struct {
	Configs  []EventConfig
	Duration time.Duration
}

type ActiveEvent struct {
	SlotIndex  int
	LocationID int32
	StartTime  time.Time
	EndTime    time.Time
	Config     EventConfig
	SeenBy     []int64
}

type EventManager struct {
	mu             sync.RWMutex
	slotData       [NumEventSlots]slotInternalData
	locationGetter func(gameMode string) []int32
	tickers        [NumEventSlots]*time.Ticker

	stopChan chan struct{}
}

type slotInternalData struct {
	schedule      EventSlotSchedule
	rotationIndex int
	currentEvent  ActiveEvent
}

var (
	eventManagerInstance *EventManager
	eventManagerOnce     sync.Once

	errInvalidSlot = errors.New("invalid event slot index")
)

func InitEventManager(schedules [NumEventSlots]EventSlotSchedule) {
	eventManagerOnce.Do(func() {
		for i, s := range schedules {
			if len(s.Configs) == 0 {
				panic(fmt.Sprintf("schedule for slot %d is empty", i))
			}

			if s.Duration <= 0 {
				panic(fmt.Sprintf("invalid duration for slot %d", i))
			}
		}

		getter := func(gamemode string) []int32 {
			ids := csv.GetLocationsByGamemode(gamemode)

			if len(ids) == 0 {
				allIds := csv.LocationIds()

				if len(allIds) == 0 {
					slog.Error("Cannot get any locations, even fallback failed!")
					return []int32{0}
				}

				slog.Warn("could not get locations for game mode, using random", "gamemode", gamemode)

				return allIds
			}

			return ids
		}

		em := &EventManager{
			locationGetter: getter,
			stopChan:       make(chan struct{}),
		}

		for i := 0; i < NumEventSlots; i++ {
			em.slotData[i] = slotInternalData{
				schedule:      schedules[i],
				rotationIndex: -1,
			}

			err := em.rotateEventForSlot(i, time.Now())

			if err != nil {
				panic(fmt.Sprintf("initial rotation failed for slot %d: %v", i, err))
			}
		}

		eventManagerInstance = em
		eventManagerInstance.startRotationLoops()

		slog.Info("event manager initialized")
	})
}

func GetEventManager() *EventManager {
	if eventManagerInstance == nil {
		panic("event manager not initialized")
	}

	return eventManagerInstance
}

func DefaultSchedules() [NumEventSlots]EventSlotSchedule {
	return [NumEventSlots]EventSlotSchedule{
		{
			Configs: []EventConfig{
				{Gamemode: GameModeShowdown, RequiredBrawlers: 0, CoinsToClaim: 0, BonusCoins: 0, CoinsToWin: 100, EventText: "Solo Showdown"},
				{Gamemode: GameModeGemGrab, RequiredBrawlers: 0, CoinsToClaim: 0, BonusCoins: 0, CoinsToWin: 100, EventText: "Gem Grab"},
			},
			Duration: 2 * time.Hour,
		},
		{
			Configs: []EventConfig{
				{Gamemode: GameModeBrawlBall, RequiredBrawlers: 0, CoinsToClaim: 0, BonusCoins: 0, CoinsToWin: 100, EventText: "Brawl Ball", DoubleExp: true},
			},
			Duration: 1 * time.Hour,
		},
		{
			Configs: []EventConfig{
				{Gamemode: GameModeHeist, RequiredBrawlers: 0, CoinsToClaim: 100, BonusCoins: 0, CoinsToWin: 100, EventText: "Heist", DoubleCoins: true},
				{Gamemode: GameModeBounty, RequiredBrawlers: 0, CoinsToClaim: 0, BonusCoins: 0, CoinsToWin: 100, EventText: "Bounty"},
			},
			Duration: 3 * time.Hour,
		},
		{
			Configs: []EventConfig{
				{Gamemode: GameModeShowdown, RequiredBrawlers: 0, CoinsToClaim: 0, BonusCoins: 0, CoinsToWin: 100, EventText: "Solo Showdown"},
			},
			Duration: 4 * time.Hour,
		},
	}
}

func (em *EventManager) startRotationLoops() {
	for i := 0; i < NumEventSlots; i++ {
		ticker := time.NewTicker(em.slotData[i].schedule.Duration)

		em.tickers[i] = ticker

		go em.slotRotationLoop(i, ticker.C)
	}
}

func (em *EventManager) Close() {
	em.mu.Lock()
	defer em.mu.Unlock()

	close(em.stopChan)

	for i := 0; i < NumEventSlots; i++ {
		if em.tickers[i] != nil {
			em.tickers[i].Stop()
		}
	}

	slog.Info("rotation loops stopped")
}

func (em *EventManager) slotRotationLoop(slotIndex int, tick <-chan time.Time) {
	slog.Info("starting rotation loop", "slot", slotIndex)

	for {
		select {
		case now := <-tick:
			slog.Info("rotation triggered by ticker", "slot", slotIndex)

			err := em.rotateEventForSlot(slotIndex, now)

			if err != nil {
				slog.Error("failed to rotate event!", "slot", slotIndex, "err", err)
			}
		case <-em.stopChan:
			slog.Info("stopping rotation loop", "slot", slotIndex)
			return
		}
	}
}

func (em *EventManager) rotateEventForSlot(slotIndex int, startTime time.Time) error {
	if slotIndex < 0 || slotIndex >= NumEventSlots {
		return errInvalidSlot
	}

	em.mu.Lock()
	defer em.mu.Unlock()

	slot := &em.slotData[slotIndex]
	schedule := &slot.schedule

	if len(schedule.Configs) == 0 {
		return fmt.Errorf("schedule for slot %d is empty", slotIndex)
	}

	slot.rotationIndex = (slot.rotationIndex + 1) % len(schedule.Configs)
	config := schedule.Configs[slot.rotationIndex]

	possibleLocations := em.locationGetter(config.Gamemode)

	var locationID int32

	if len(possibleLocations) > 0 {
		locationID = possibleLocations[rand.Intn(len(possibleLocations))]
	} else {
		slog.Error("no locations found, using fallback", "gamemode", config.Gamemode, "slot", slotIndex)
		locationID = 0
	}

	slot.currentEvent = ActiveEvent{
		SlotIndex:  slotIndex,
		LocationID: locationID,
		StartTime:  startTime,
		EndTime:    startTime.Add(schedule.Duration),
		Config:     config,
		SeenBy:     []int64{},
	}

	slog.Info("rotated event", "slot", slotIndex, "gamemode", config.Gamemode, "location", locationID, "endTime", slot.currentEvent.EndTime)

	return nil
}

func (em *EventManager) GetCurrentEventPtr(slot int32) *ActiveEvent {
	return &em.slotData[slot].currentEvent
}

func (em *EventManager) Embed(stream *ByteStream, player *Player) {
	em.mu.RLock()
	defer em.mu.RUnlock()

	now := time.Now()

	stream.Write(VInt(NumEventSlots))

	for i := 0; i < NumEventSlots; i++ {
		event := em.slotData[i].currentEvent

		stream.Write(VInt(event.SlotIndex + 1))
		stream.Write(VInt(event.Config.RequiredBrawlers))
	}

	stream.Write(VInt(NumEventSlots))

	for i := 0; i < NumEventSlots; i++ {
		event := em.slotData[i].currentEvent
		timeLeft := int32(event.EndTime.Sub(now).Seconds())

		if timeLeft < 0 {
			timeLeft = 1
		}
		
		visionState := 1
		
		for _, id := range event.SeenBy {
			if id == player.DbId {
				visionState = 2
				break
			}
		}

		stream.Write(VInt(event.SlotIndex + 1))
		stream.Write(VInt(event.SlotIndex + 1))

		stream.Write(VInt(1))
		stream.Write(VInt(timeLeft))

		stream.Write(VInt(event.Config.CoinsToClaim))
		stream.Write(VInt(event.Config.BonusCoins))
		stream.Write(VInt(event.Config.CoinsToWin))

		stream.Write(event.Config.DoubleCoins)
		stream.Write(event.Config.DoubleExp)

		stream.Write(ScId{15, event.LocationID})

		stream.Write(VInt(0))
		stream.Write(VInt(visionState)) // 1=new event, 2=seen

		stream.Write(event.Config.EventText)
		stream.Write(false)
	}

	stream.Write(VInt(NumEventSlots))

	for i := 0; i < NumEventSlots; i++ {
		slot := &em.slotData[i]
		schedule := &slot.schedule
		currentEvent := &slot.currentEvent

		nextIndexInSlotSchedule := (slot.rotationIndex + 1) % len(schedule.Configs)
		nextConfig := schedule.Configs[nextIndexInSlotSchedule]

		timeUntilNextRotation := int32(0)

		if currentEvent.EndTime.After(now) {
			timeUntilNextRotation = int32(currentEvent.EndTime.Sub(now).Seconds())
		}

		if timeUntilNextRotation <= 0 {
			timeUntilNextRotation = 1
		}

		totalDurationSeconds := int32(schedule.Duration.Seconds())
		possibleLocations := em.locationGetter(nextConfig.Gamemode)

		var locationID int32

		if len(possibleLocations) > 0 {
			locationID = possibleLocations[rand.Intn(len(possibleLocations))]
		} else {
			locationID = 0
		}

		stream.Write(VInt(i + 1))
		stream.Write(VInt(i + 1))

		stream.Write(VInt(timeUntilNextRotation))
		stream.Write(VInt(totalDurationSeconds))

		stream.Write(VInt(nextConfig.CoinsToClaim))
		stream.Write(VInt(nextConfig.BonusCoins))
		stream.Write(VInt(nextConfig.CoinsToWin))

		stream.Write(false)
		stream.Write(false)

		stream.Write(ScId{15, locationID})

		stream.Write(VInt(0))
		stream.Write(VInt(1))

		stream.Write(nextConfig.EventText)
		stream.Write(false)
	}
}
