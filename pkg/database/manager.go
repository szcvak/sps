package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/szcvak/sps/pkg/config"
	"github.com/szcvak/sps/pkg/core"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Manager struct {
	pool *pgxpool.Pool
}

func NewManager() (*Manager, error) {
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))

	if err != nil {
		return nil, fmt.Errorf("unable to create database pool: %v", err)
	}

	return &Manager{
		pool: pool,
	}, nil
}

func (m *Manager) Close() {
	if m.pool != nil {
		m.pool.Close()
	}
}

func (m *Manager) Pool() *pgxpool.Pool {
	return m.pool
}

func (m *Manager) CreateDefault() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tx, err := m.pool.Begin(ctx)

	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	ddlStatements := []struct {
		Name string
		Stmt string
	}{
		{"players table", players},
		{"player progression table", playerProgression},
		{"player brawlers table", playerBrawlers},
		{"player unlocked star powers table", playerUnlockedStarPowers},
		{"player unlocked star gadgets table", playerUnlockedGadgets},
		{"player unlocked star gears table", playerUnlockedGears},
		{"player wallet table", playerWallet},
	}

	for _, entry := range ddlStatements {
		if _, err = tx.Exec(ctx, entry.Stmt); err != nil {
			return fmt.Errorf("could not create %s: %v", entry.Name, err)
		}
	}

	err = tx.Commit(ctx)

	if err != nil {
		return fmt.Errorf("transaction commit error: %v", err)
	}

	slog.Info("created all default tables")

	return nil
}

func (m *Manager) CreatePlayer(ctx context.Context, highId int32, lowId int32, name string, token string, region string) (*core.Player, error) {
	tx, err := m.pool.Begin(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer tx.Rollback(ctx)

	var newPlayerId int64
	var createdAt time.Time

	playerInsertSQL := `
		INSERT INTO players (name, token, high_id, low_id, region, last_login)
		VALUES ($1, $2, $3, $4, $5, current_timestamp)
		RETURNING id, created_at`

	err = tx.QueryRow(ctx, playerInsertSQL, name, token, highId, lowId, region).Scan(&newPlayerId, &createdAt)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf("%w: %v", ErrAccountAlreadyExists, pgErr.ConstraintName)
		}

		return nil, fmt.Errorf("failed to insert player: %w", err)
	}

	progressionInsertSQL := `
		INSERT INTO player_progression (player_id, trophies, highest_trophies, solo_victories, duo_victories, trio_victories)
		VALUES ($1, $2, $2, 0, 0, 0)`

	_, err = tx.Exec(ctx, progressionInsertSQL, newPlayerId, config.NewPlayerTrophies)

	if err != nil {
		return nil, fmt.Errorf("failed to insert player progression for player %d: %w", newPlayerId, err)
	}

	walletInsertSQL := `INSERT INTO player_wallet (player_id, currency_id, balance) VALUES ($1, $2, $3)`
	newPlayerWallet := make(map[int32]*core.PlayerCurrency)

	for _, currencyId := range config.DefaultCurrencies {
		balance := config.DefaultCurrencyBalance[currencyId]

		_, err = tx.Exec(ctx, walletInsertSQL, newPlayerId, currencyId, balance)

		if err != nil {
			return nil, fmt.Errorf("failed to insert currency %d for player %d: %w", currencyId, newPlayerId, err)
		}

		newPlayerWallet[currencyId] = &core.PlayerCurrency{
			CurrencyId: currencyId,
			Balance:    int64(balance),
		}
	}

	var unlocked []int32

	err = json.Unmarshal([]byte(defaultUnlockedSkinsJson), &unlocked)

	if err != nil {
		return nil, fmt.Errorf("failed to parse default unlocked skins JSON: %w", err)
	}

	var defaultCardsMap map[string]int32

	err = json.Unmarshal([]byte(defaultBrawlerCards), &defaultCardsMap)

	if err != nil {
		return nil, fmt.Errorf("failed to parse default cards JSON '%s': %w", defaultBrawlerCards, err)
	}

	startingBrawler := &core.PlayerBrawler{
		BrawlerId:         config.NewPlayerStartingBrawlerId,
		Trophies:          0,
		HighestTrophies:   0,
		PowerLevel:        1,
		PowerPoints:       0,
		SelectedGadget:    nil,
		SelectedStarPower: nil,
		SelectedGear1:     nil,
		SelectedGear2:     nil,
		UnlockedSkinIds:   unlocked,
		SelectedSkinId:    int32(defaultSkinId),
		Cards:             defaultCardsMap,
	}

	brawlerInsertSQL := `
		INSERT INTO player_brawlers (
			player_id, brawler_id, trophies, highest_trophies,
			power_level, power_points,
			unlocked_skins, selected_skin, cards
		)
		VALUES ($1, $2, 0, 0, 1, 0, $3, $4, $5)`

	_, err = tx.Exec(ctx, brawlerInsertSQL,
		newPlayerId,
		config.NewPlayerStartingBrawlerId,
		defaultUnlockedSkinsJson,
		defaultSkinId,
		defaultBrawlerCards,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to insert starting brawler %d for player %d: %w", config.NewPlayerStartingBrawlerId, newPlayerId, err)
	}

	err = tx.Commit(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction for player %d: %w", newPlayerId, err)
	}

	player := core.NewPlayer()

	player.DbId = newPlayerId
	player.HighId = highId
	player.LowId = lowId
	player.Name = name
	player.Token = token
	player.Region = region
	player.CreatedAt = createdAt
	player.LastLogin = createdAt
	player.ProfileIcon = 0

	player.Trophies = config.NewPlayerTrophies
	player.HighestTrophies = config.NewPlayerTrophies
	player.SoloVictories = 0
	player.DuoVictories = 0
	player.TrioVictories = 0

	player.BattleHints = false
	player.ControlMode = 0
	player.CoinBooster = 0

	player.Wallet = newPlayerWallet

	player.Brawlers[config.NewPlayerStartingBrawlerId] = startingBrawler

	player.SetState(core.StateSession)

	slog.Info("created player", "id", newPlayerId, "name", name)

	return player, nil
}

func (m *Manager) LoadPlayerByToken(ctx context.Context, token string) (*core.Player, error) {
	conn, err := m.pool.Acquire(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %w", err)
	}

	defer conn.Release()

	player := core.NewPlayer()

	// core/progression
	stmt := `
		select
			p.id, p.control_mode, p.battle_hints, p.coin_booster, p.high_id, p.low_id, p.name, p.region, p.profile_icon, p.created_at, p.last_login, p.tutorial_state, p.coins_reward,
			pp.trophies, pp.highest_trophies, pp.solo_victories, pp.duo_victories, pp.trio_victories, pp.experience
		from players p
		join player_progression pp ON p.id = pp.player_id
		where p.token = $1`

	err = conn.QueryRow(ctx, stmt, token).Scan(
		&player.DbId, &player.ControlMode, &player.BattleHints, &player.CoinBooster, &player.HighId, &player.LowId, &player.Name, &player.Region, &player.ProfileIcon, &player.CreatedAt, &player.LastLogin, &player.TutorialState, &player.CoinsReward,
		&player.Trophies, &player.HighestTrophies, &player.SoloVictories, &player.DuoVictories, &player.TrioVictories, &player.Experience,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w for player with token %s", ErrPlayerNotFound, token)
		}

		return nil, fmt.Errorf("failed to query core/progression data for player with token %s: %w", token, err)
	}

	playerId := player.DbId

	// brawlers
	stmt = `
		select
			brawler_id, trophies, highest_trophies, power_level, power_points,
			selected_gadget, selected_star_power, selected_gear1, selected_gear2,
			unlocked_skins, selected_skin, cards
		from player_brawlers
		where player_id = $1`

	rowsBrawlers, err := conn.Query(ctx, stmt, playerId)

	if err != nil {
		return nil, fmt.Errorf("failed to query brawlers for player %d: %w", playerId, err)
	}

	defer rowsBrawlers.Close()

	for rowsBrawlers.Next() {
		b := &core.PlayerBrawler{}

		err = rowsBrawlers.Scan(
			&b.BrawlerId, &b.Trophies, &b.HighestTrophies, &b.PowerLevel, &b.PowerPoints,
			&b.SelectedGadget, &b.SelectedStarPower, &b.SelectedGear1, &b.SelectedGear2,
			&b.UnlockedSkinIds, &b.SelectedSkinId, &b.Cards,
		)

		if err != nil {
			slog.Warn("skipping row while loading player data", "playerId", playerId, "err", err)
			continue
		}

		player.Brawlers[b.BrawlerId] = b
	}

	if err = rowsBrawlers.Err(); err != nil {
		return nil, fmt.Errorf("error iterating brawler rows for player %d: %w", playerId, err)
	}

	// wallet
	stmt = `select currency_id, balance from player_wallet where player_id = $1`

	rowsWallet, err := conn.Query(ctx, stmt, playerId)

	if err != nil {
		return nil, fmt.Errorf("failed to query wallet for player %d: %w", playerId, err)
	}

	defer rowsWallet.Close()

	for rowsWallet.Next() {
		c := &core.PlayerCurrency{}

		err = rowsWallet.Scan(&c.CurrencyId, &c.Balance)

		if err != nil {
			slog.Warn("skipping row while loading player data", "playerId", playerId, "err", err)
			continue
		}

		player.Wallet[c.CurrencyId] = c
	}

	if err = rowsWallet.Err(); err != nil {
		return nil, fmt.Errorf("error iterating wallet rows for player %d: %w", playerId, err)
	}

	return player, nil
}

func (m *Manager) LoadPlayerByIds(ctx context.Context, highId int32, lowId int32) (*core.Player, error) {
	conn, err := m.pool.Acquire(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %w", err)
	}

	defer conn.Release()

	player := core.NewPlayer()

	// core/progression
	stmt := `
		select
			p.id, p.control_mode, p.battle_hints, p.coin_booster, p.high_id, p.low_id, p.name, p.region, p.profile_icon, p.created_at, p.last_login, p.tutorial_state, p.coins_reward,
			pp.trophies, pp.highest_trophies, pp.solo_victories, pp.duo_victories, pp.trio_victories, pp.experience
		from players p
		join player_progression pp ON p.id = pp.player_id
		where p.high_id = $1 and p.low_id = $2`

	err = conn.QueryRow(ctx, stmt, highId, lowId).Scan(
		&player.DbId, &player.ControlMode, &player.BattleHints, &player.CoinBooster, &player.HighId, &player.LowId, &player.Name, &player.Region, &player.ProfileIcon, &player.CreatedAt, &player.LastLogin, &player.TutorialState, &player.CoinsReward,
		&player.Trophies, &player.HighestTrophies, &player.SoloVictories, &player.DuoVictories, &player.TrioVictories, &player.Experience,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w for player with ids %d, %d", ErrPlayerNotFound, highId, lowId)
		}

		return nil, fmt.Errorf("failed to query core/progression data for player with ids %d, %d: %w", highId, lowId, err)
	}

	playerId := player.DbId

	// brawlers
	stmt = `
		select
			brawler_id, trophies, highest_trophies, power_level, power_points,
			selected_gadget, selected_star_power, selected_gear1, selected_gear2,
			unlocked_skins, selected_skin, cards
		from player_brawlers
		where player_id = $1`

	rowsBrawlers, err := conn.Query(ctx, stmt, playerId)

	if err != nil {
		return nil, fmt.Errorf("failed to query brawlers for player %d: %w", playerId, err)
	}

	defer rowsBrawlers.Close()

	for rowsBrawlers.Next() {
		b := &core.PlayerBrawler{}

		err = rowsBrawlers.Scan(
			&b.BrawlerId, &b.Trophies, &b.HighestTrophies, &b.PowerLevel, &b.PowerPoints,
			&b.SelectedGadget, &b.SelectedStarPower, &b.SelectedGear1, &b.SelectedGear2,
			&b.UnlockedSkinIds, &b.SelectedSkinId, &b.Cards,
		)

		if err != nil {
			slog.Warn("skipping row while loading player data", "playerId", playerId, "err", err)
			continue
		}

		player.Brawlers[b.BrawlerId] = b
	}

	if err = rowsBrawlers.Err(); err != nil {
		return nil, fmt.Errorf("error iterating brawler rows for player %d: %w", playerId, err)
	}

	// wallet
	stmt = `select currency_id, balance from player_wallet where player_id = $1`

	rowsWallet, err := conn.Query(ctx, stmt, playerId)

	if err != nil {
		return nil, fmt.Errorf("failed to query wallet for player %d: %w", playerId, err)
	}

	defer rowsWallet.Close()

	for rowsWallet.Next() {
		c := &core.PlayerCurrency{}

		err = rowsWallet.Scan(&c.CurrencyId, &c.Balance)

		if err != nil {
			slog.Warn("skipping row while loading player data", "playerId", playerId, "err", err)
			continue
		}

		player.Wallet[c.CurrencyId] = c
	}

	if err = rowsWallet.Err(); err != nil {
		return nil, fmt.Errorf("error iterating wallet rows for player %d: %w", playerId, err)
	}

	return player, nil
}

func (m *Manager) Exec(query string, args ...interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := m.Pool().Exec(ctx, query, args...)
	return err
}
