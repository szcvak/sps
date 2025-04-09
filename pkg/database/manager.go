package database

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/szcvak/sps/pkg/core"
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

		fmt.Printf("created %s\n", entry.Name)
	}

	err = tx.Commit(ctx)

	if err != nil {
		return fmt.Errorf("transaction commit error: %v", err)
	}

	fmt.Println("created all default tables")

	return nil
}

func (m *Manager) CreatePlayer(ctx context.Context, highId int32, lowId int32, name string, token string, region string) error {
	tx, err := m.pool.Begin(ctx)

	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer tx.Rollback(ctx)

	var newPlayerId int64

	// players
	stmt := `
		insert into players (name, token, high_id, low_id, region)
		values ($1, $2, $3, $4, $5)
		returning id`
	err = tx.QueryRow(ctx, stmt, name, token, highId, lowId, region).Scan(&newPlayerId)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("%w: %v", ErrAccountAlreadyExists, pgErr.ConstraintName)
		}

		return fmt.Errorf("failed to insert player: %w", err)
	}

	// player progression
	stmt = `
		insert into player_progression (player_id)
		values ($1)`
	_, err = tx.Exec(ctx, stmt, newPlayerId)

	if err != nil {
		return fmt.Errorf("failed to insert player progression for player %d: %w", newPlayerId, err)
	}

	// player wallet
	stmt = `insert into player_wallet (player_id, currency_id, balance) VALUES ($1, $2, $3)`

	for _, currencyId := range defaultCurrencies {
		_, err = tx.Exec(ctx, stmt, newPlayerId, currencyId, defaultCurrencyBalance[currencyId])

		if err != nil {
			return fmt.Errorf("failed to insert currency %d for player %d: %w", currencyId, newPlayerId, err)
		}
	}

	// brawlers
	stmt = `
		insert into player_brawlers (
			player_id, brawler_id, trophies, highest_trophies,
			power_level, power_points,
			unlocked_skins, selected_skin
		)
		values ($1, $2, 0, 0, 1, 0, $3, $4)`

	_, err = tx.Exec(ctx, stmt,
		newPlayerId,
		defaultStartingBrawlerId,
		defaultUnlockedSkinsJson,
		defaultSkinId,
	)

	if err != nil {
		return fmt.Errorf("failed to insert starting brawler %d for player %d: %w", defaultStartingBrawlerId, newPlayerId, err)
	}

	err = tx.Commit(ctx)

	if err != nil {
		return fmt.Errorf("failed to commit transaction for player %d: %w", newPlayerId, err)
	}

	fmt.Printf("created player %s (%d)\n", name, newPlayerId)

	return nil
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
			p.id, p.high_id, p.low_id, p.name, p.region, p.profile_icon, p.created_at, p.last_login,
			pp.trophies, pp.highest_trophies, pp.solo_victories, pp.duo_victories, pp.trio_victories
		from players p
		join player_progression pp ON p.id = pp.player_id
		where p.token = $1`

	err = conn.QueryRow(ctx, stmt, token).Scan(
		&player.DbId, &player.HighId, &player.LowId, &player.Name, &player.Region, &player.ProfileIcon, &player.CreatedAt, &player.LastLogin,
		&player.Trophies, &player.HighestTrophies, &player.SoloVictories, &player.DuoVictories, &player.TrioVictories,
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
			unlocked_skins, selected_skin
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
			&b.UnlockedSkinIds, &b.SelectedSkinId,
		)

		if err != nil {
			fmt.Printf("error scanning brawler row for player %d: %v\n", playerId, err)
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
			fmt.Printf("error scanning currency row for player %d: %v\n", playerId, err)
			continue
		}

		player.Wallet[c.CurrencyId] = c
	}

	if err = rowsWallet.Err(); err != nil {
		return nil, fmt.Errorf("error iterating wallet rows for player %d: %w", playerId, err)
	}

	fmt.Printf("loaded data for player %s (%d)\n", player.Name, playerId)

	return player, nil
}
