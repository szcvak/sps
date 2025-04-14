package database

import (
	"context"
	"database/sql"
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

type LeaderboardPlayerEntry struct {
	DbId         int64  `db:"id"`
	Name         string `db:"name"`
	Trophies     int32  `db:"trophies"`
	ProfileIcon  int32  `db:"profile_icon"`
	Region       string `db:"region"`
	PlayerHighId int32  `db:"high_id"`
	PlayerLowId  int32  `db:"low_id"`
	AllianceId   *int64 `db:"alliance_id"`
	PlayerExperience   int32  `db:"experience"`
}

type LeaderboardAllianceEntry struct {
	DbId          int64  `db:"id"`
	Name          string `db:"name"`
	BadgeId       int32  `db:"badge_id"`
	Type          int16  `db:"type"`
	TotalTrophies int32  `db:"total_trophies"`
	TotalMembers  int32
}

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
		{"alliances", alliances},
		{"alliance members", allianceMembers},
		{"alliance messages", allianceMessages},
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
		insert into players (name, token, high_id, low_id, region, last_login)
		values ($1, $2, $3, $4, $5, current_timestamp)
		returning id, created_at`

	err = tx.QueryRow(ctx, playerInsertSQL, name, token, highId, lowId, region).Scan(&newPlayerId, &createdAt)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf("%w: %v", ErrAccountAlreadyExists, pgErr.ConstraintName)
		}

		return nil, fmt.Errorf("failed to insert player: %w", err)
	}

	progressionInsertSQL := `
		insert into player_progression (player_id, trophies, highest_trophies, solo_victories, duo_victories, trio_victories)
		values ($1, $2, $2, 0, 0, 0)`

	_, err = tx.Exec(ctx, progressionInsertSQL, newPlayerId, config.NewPlayerTrophies)

	if err != nil {
		return nil, fmt.Errorf("failed to insert player progression for player %d: %w", newPlayerId, err)
	}

	walletInsertSQL := `insert into player_wallet (player_id, currency_id, balance) values ($1, $2, $3)`
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
		insert into player_brawlers (
			player_id, brawler_id, trophies, highest_trophies,
			power_level, power_points,
			unlocked_skins, selected_skin, cards
		)
		values ($1, $2, 0, 0, 1, 0, $3, $4, $5)`

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

func (m *Manager) LoadPlayerByIds(ctx context.Context, high int32, low int32) (*core.Player, error) {
	conn, err := m.pool.Acquire(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %w", err)
	}

	defer conn.Release()

	player := core.NewPlayer()

	var allianceId sql.NullInt64
	var allianceRole sql.NullInt16

	// core/progression
	stmt := `
		select
			p.id, p.control_mode, p.battle_hints, p.coin_booster, p.high_id, p.low_id, p.name, p.region, p.profile_icon, p.created_at, p.last_login, p.tutorial_state, p.coins_reward, p.coin_doubler,
			pp.trophies, pp.highest_trophies, pp.solo_victories, pp.duo_victories, pp.trio_victories, pp.experience,
			am.alliance_id, am.role
		from players p
		join player_progression pp on p.id = pp.player_id
		left join alliance_members am on p.id = am.player_id
		where p.high_id = $1 and p.low_id = $2`

	err = conn.QueryRow(ctx, stmt, high, low).Scan(
		&player.DbId, &player.ControlMode, &player.BattleHints, &player.CoinBooster, &player.HighId, &player.LowId, &player.Name, &player.Region, &player.ProfileIcon, &player.CreatedAt, &player.LastLogin, &player.TutorialState, &player.CoinsReward, &player.CoinDoubler,
		&player.Trophies, &player.HighestTrophies, &player.SoloVictories, &player.DuoVictories, &player.TrioVictories, &player.Experience,
		&allianceId, &allianceRole,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w", ErrPlayerNotFound)
		}

		return nil, fmt.Errorf("failed to query core/progression data: %w", err)
	}

	if allianceId.Valid {
		player.AllianceId = &allianceId.Int64
	} else {
		player.AllianceId = nil
	}

	if allianceRole.Valid {
		player.AllianceRole = allianceRole.Int16
	} else {
		player.AllianceRole = 0
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

func (m *Manager) LoadPlayerByToken(ctx context.Context, token string) (*core.Player, error) {
	conn, err := m.pool.Acquire(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %w", err)
	}

	defer conn.Release()

	player := core.NewPlayer()

	var allianceId sql.NullInt64
	var allianceRole sql.NullInt16

	// core/progression
	stmt := `
		select
			p.id, p.control_mode, p.battle_hints, p.coin_booster, p.high_id, p.low_id, p.name, p.region, p.profile_icon, p.created_at, p.last_login, p.tutorial_state, p.coins_reward, p.coin_doubler,
			pp.trophies, pp.highest_trophies, pp.solo_victories, pp.duo_victories, pp.trio_victories, pp.experience,
			am.alliance_id, am.role
		from players p
		join player_progression pp on p.id = pp.player_id
		left join alliance_members am on p.id = am.player_id
		where p.token = $1`

	err = conn.QueryRow(ctx, stmt, token).Scan(
		&player.DbId, &player.ControlMode, &player.BattleHints, &player.CoinBooster, &player.HighId, &player.LowId, &player.Name, &player.Region, &player.ProfileIcon, &player.CreatedAt, &player.LastLogin, &player.TutorialState, &player.CoinsReward, &player.CoinDoubler,
		&player.Trophies, &player.HighestTrophies, &player.SoloVictories, &player.DuoVictories, &player.TrioVictories, &player.Experience,
		&allianceId, &allianceRole,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w for player with token %s", ErrPlayerNotFound, token)
		}

		return nil, fmt.Errorf("failed to query core/progression data for player with token %s: %w", token, err)
	}

	if allianceId.Valid {
		player.AllianceId = &allianceId.Int64
	} else {
		player.AllianceId = nil
	}

	if allianceRole.Valid {
		player.AllianceRole = allianceRole.Int16
	} else {
		player.AllianceRole = 0
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

func (m *Manager) LoadAlliance(ctx context.Context, allianceId int64) (*core.Alliance, error) {
	conn, err := m.pool.Acquire(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %w", err)
	}

	defer conn.Release()

	a := core.NewAlliance(allianceId)

	stmt := `
        select
            name, description, badge_id, type, required_trophies,
            total_trophies, creator_id, region
        from alliances
        where id = $1`

	var creatorId sql.NullInt64

	err = conn.QueryRow(ctx, stmt, allianceId).Scan(
		&a.Name, &a.Description, &a.BadgeId, &a.Type, &a.RequiredTrophies,
		&a.TotalTrophies, &creatorId, &a.Region,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to query alliance core data for id %d: %w", allianceId, err)
	}

	if creatorId.Valid {
		a.CreatorId = &creatorId.Int64
	}

	stmt = `
        select
            am.player_id, am.role,
			p.name, p.low_id, p.profile_icon,
			pp.experience, pp.trophies
        from alliance_members am
		join players p on am.player_id = p.id
		join player_progression pp on p.id = pp.player_id
        where am.alliance_id = $1
		order by am.role desc, pp.trophies desc`

	rows, err := conn.Query(ctx, stmt, allianceId)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	a.Members = make([]core.AllianceMember, 0)

	for rows.Next() {
		member := core.AllianceMember{}

		err = rows.Scan(
			&member.PlayerId,
			&member.Role,
			&member.Name,
			&member.LowId,
			&member.ProfileIcon,
			&member.Experience,
			&member.Trophies,
		)

		if err != nil {
			slog.Error("failed to scan alliance member row, skipping", "allianceId", allianceId, "err", err)
			continue
		}

		a.Members = append(a.Members, member)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	a.TotalMembers = int32(len(a.Members))

	return a, nil
}

func (m *Manager) GetAlliances(ctx context.Context) ([]*core.Alliance, error) {
	conn, err := m.pool.Acquire(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection for getting alliances: %w", err)
	}

	defer conn.Release()

	stmt := `
		select
			a.id, a.name, a.description, a.badge_id, a.type,
			a.required_trophies, a.total_trophies,
			a.creator_id, a.region,
			coalesce(mc.member_count, 0) as current_member_count
		from alliances a
		left join (
			select alliance_id, count(*) as member_count
			from alliance_members
			group by alliance_id
		) mc on a.id = mc.alliance_id
		order by a.total_trophies desc, a.name asc`

	rows, err := conn.Query(ctx, stmt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []*core.Alliance{}, nil
		}

		return nil, fmt.Errorf("failed to query alliances with counts: %w", err)
	}

	defer rows.Close()

	alliancesList := make([]*core.Alliance, 0)

	for rows.Next() {
		a := core.NewAlliance(0)

		var creatorId sql.NullInt64
		var memberCount int32

		err := rows.Scan(
			&a.Id,
			&a.Name,
			&a.Description,
			&a.BadgeId,
			&a.Type,
			&a.RequiredTrophies,
			&a.TotalTrophies,
			&creatorId,
			&a.Region,
			&memberCount,
		)

		if err != nil {
			slog.Error("failed to scan alliance row with count, skipping", "err", err)
			continue
		}

		if creatorId.Valid {
			a.CreatorId = &creatorId.Int64
		} else {
			a.CreatorId = nil
		}

		a.TotalMembers = memberCount
		a.Members = make([]core.AllianceMember, 0)

		alliancesList = append(alliancesList, a)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating alliance rows with count: %w", err)
	}

	return alliancesList, nil
}

func (m *Manager) LoadAllianceMessages(ctx context.Context, allianceId int64, limit int) ([]core.AllianceMessage, error) {
	if limit <= 0 {
		limit = 50
	}

	conn, err := m.pool.Acquire(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %w", err)
	}

	defer conn.Release()

	stmt := `
        select
            id, alliance_id, player_id,
            player_high_id, player_low_id, player_name, player_role,
            player_icon,
            message_type, message_content,
            target_id, target_name,
            created_at
        from alliance_messages
        where alliance_id = $1
        order by created_at desc
        limit $2`

	rows, err := conn.Query(ctx, stmt, allianceId, limit)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []core.AllianceMessage{}, nil
		}

		return nil, fmt.Errorf("failed to query alliance messages for id %d: %w", allianceId, err)
	}

	defer rows.Close()

	messages := make([]core.AllianceMessage, 0, limit)

	for rows.Next() {
		var msg core.AllianceMessage

		var dbPlayerId sql.NullInt64
		var dbContent sql.NullString
		var dbTargetId sql.NullInt64
		var dbTargetName sql.NullString

		err := rows.Scan(
			&msg.Id, &msg.AllianceId, &dbPlayerId,
			&msg.PlayerHighId, &msg.PlayerLowId, &msg.PlayerName, &msg.PlayerRole,
			&msg.PlayerIcon,
			&msg.Type, &dbContent,
			&dbTargetId, &dbTargetName,
			&msg.Timestamp,
		)

		if err != nil {
			slog.Error("failed to scan alliance message row, skipping", "allianceId", allianceId, "err", err)
			continue
		}

		if dbPlayerId.Valid {
			msg.PlayerId = &dbPlayerId.Int64
		}

		if dbContent.Valid {
			msg.Content = dbContent.String
		}

		if dbTargetId.Valid {
			msg.TargetId = &dbTargetId.Int64
		}

		if dbTargetName.Valid {
			msg.TargetName = dbTargetName.String
		}

		messages = append(messages, msg)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating alliance message rows for id %d: %w", allianceId, err)
	}

	reverseMessages(messages)

	return messages, nil
}

func (m *Manager) AddAllianceMessage(ctx context.Context, allianceId int64, sender *core.Player, msgType int16, content string, targetPlayer *core.Player) (*core.AllianceMessage, error) {
	conn, err := m.pool.Acquire(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %w", err)
	}

	defer conn.Release()

	var dbPlayerId sql.NullInt64
	var dbPlayerHighId sql.NullInt32
	var dbPlayerLowId sql.NullInt32
	var dbPlayerName sql.NullString
	var dbPlayerRole sql.NullInt16
	var dbPlayerIcon sql.NullInt32

	if sender != nil {
		dbPlayerId.Valid = true
		dbPlayerId.Int64 = sender.DbId

		dbPlayerHighId.Valid = true
		dbPlayerHighId.Int32 = sender.HighId

		dbPlayerLowId.Valid = true
		dbPlayerLowId.Int32 = sender.LowId

		dbPlayerName.Valid = true
		dbPlayerName.String = sender.Name

		dbPlayerRole.Valid = true
		dbPlayerRole.Int16 = sender.AllianceRole

		dbPlayerIcon.Valid = true
		dbPlayerIcon.Int32 = sender.ProfileIcon
	}

	var dbContent sql.NullString

	if content != "" {
		dbContent.Valid = true
		dbContent.String = content
	}

	var dbTargetId sql.NullInt64
	var dbTargetName sql.NullString

	if targetPlayer != nil {
		dbTargetId.Valid = true
		dbTargetId.Int64 = targetPlayer.DbId

		dbTargetName.Valid = true
		dbTargetName.String = targetPlayer.Name
	}

	stmt := `
        insert into alliance_messages (
            alliance_id, player_id,
            player_high_id, player_low_id, player_name, player_role,
            player_icon,
            message_type, message_content,
            target_id, target_name
        ) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        returning id, created_at`

	var newMessageId int64
	var newMessageTimestamp time.Time

	err = conn.QueryRow(ctx, stmt,
		allianceId,
		dbPlayerId,
		dbPlayerHighId,
		dbPlayerLowId,
		dbPlayerName,
		dbPlayerRole,
		dbPlayerIcon,
		msgType,
		dbContent,
		dbTargetId,
		dbTargetName,
	).Scan(&newMessageId, &newMessageTimestamp)

	if err != nil {
		senderId := int64(0)

		if sender != nil {
			senderId = sender.DbId
		}

		slog.Error("failed to insert alliance message!", "allianceId", allianceId, "senderId", senderId, "type", msgType, "err", err)

		return nil, fmt.Errorf("failed to insert alliance message: %w", err)
	}

	persistedMsg := &core.AllianceMessage{
		Id:         newMessageId,
		AllianceId: allianceId,
		Type:       msgType,
		Content:    dbContent.String,
		Timestamp:  newMessageTimestamp,
	}

	if dbPlayerId.Valid {
		persistedMsg.PlayerId = &dbPlayerId.Int64
		persistedMsg.PlayerHighId = dbPlayerHighId.Int32
		persistedMsg.PlayerLowId = dbPlayerLowId.Int32
		persistedMsg.PlayerName = dbPlayerName.String
		persistedMsg.PlayerRole = dbPlayerRole.Int16
		persistedMsg.PlayerIcon = dbPlayerIcon.Int32
	}

	if dbTargetId.Valid {
		persistedMsg.TargetId = &dbTargetId.Int64
		persistedMsg.TargetName = dbTargetName.String
	}

	return persistedMsg, nil
}

func (m *Manager) AddAllianceMember(ctx context.Context, player *core.Player, allianceId int64) error {
	tx, err := m.pool.Begin(ctx)

	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback(ctx)
			panic(r)
		}

		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	_, err = tx.Exec(
		ctx,
		"update alliances set total_trophies = total_trophies + $1 where id = $2",
		player.Trophies, allianceId,
	)

	if err != nil {
		return fmt.Errorf("failed to increase alliance trophies: %w", err)
	}

	_, err = tx.Exec(
		ctx,
		"insert into alliance_members (alliance_id, player_id) values ($1, $2)",
		allianceId, player.DbId,
	)

	if err != nil {
		return fmt.Errorf("failed to insert into alliance_members: %w", err)
	}

	err = tx.Commit(ctx)

	if err != nil {
		return fmt.Errorf("failed to commit transaction for alliance member %d: %w", player.DbId, err)
	}

	player.AllianceId = new(int64)
	*player.AllianceId = allianceId

	player.AllianceRole = 1

	slog.Info("player joined an alliance", "playerId", player.DbId, "allianceId", allianceId)

	return nil
}

func (m *Manager) RemoveAllianceMember(ctx context.Context, player *core.Player, allianceId int64) (bool, error) {
	tx, err := m.pool.Begin(ctx)

	if err != nil {
		return false, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer tx.Rollback(ctx)

	_, err = tx.Exec(
		ctx,
		"update alliances set total_trophies = total_trophies - $1 where id = $2",
		player.Trophies, allianceId,
	)

	if err != nil {
		return false, fmt.Errorf("failed to decrease alliance trophies: %w", err)
	}

	_, err = tx.Exec(
		ctx,
		"delete from alliance_members where player_id = $1",
		player.DbId,
	)

	if err != nil {
		return false, fmt.Errorf("failed to delete member from alliance: %w", err)
	}

	var members int32

	err = tx.QueryRow(
		ctx,
		"select count(*) from alliance_members where alliance_id = $1",
		allianceId,
	).Scan(&members)

	if err != nil {
		return false, fmt.Errorf("failed to count members: %w", err)
	}

	deleted := false

	if members <= 0 {
		err = m.DeleteAlliance(ctx, allianceId, tx)

		if err != nil {
			return false, fmt.Errorf("failed to delete alliance: %w", err)
		}

		deleted = true
	}

	err = tx.Commit(ctx)

	if err != nil {
		return deleted, fmt.Errorf("failed to commit transaction for alliance member %d: %w", player.DbId, err)
	}

	player.AllianceId = nil
	player.AllianceRole = 0

	slog.Info("player left an alliance", "playerId", player.DbId, "allianceId", allianceId)

	return deleted, nil
}

func (m *Manager) DeleteAlliance(ctx context.Context, allianceId int64, tx pgx.Tx) error {
	var err error

	newTx := false

	if tx == nil {
		tx, err = m.pool.Begin(ctx)

		if err != nil {
			return fmt.Errorf("failed to begin transaction for delete alliance: %w", err)
		}

		newTx = true

		defer func() {
			r := recover()

			if err != nil || r != nil {
				_ = tx.Rollback(ctx)
				if r != nil {
					panic(r)
				}
			}
		}()
	}

	_, err = tx.Exec(
		ctx,
		"delete from alliance_messages where alliance_id = $1",
		allianceId,
	)

	if err != nil {
		return fmt.Errorf("failed to delete alliance messages: %w", err)
	}

	_, err = tx.Exec(
		ctx,
		"delete from alliances where id = $1",
		allianceId,
	)

	if err != nil {
		return fmt.Errorf("failed to delete alliance: %w", err)
	}

	if newTx {
		err = tx.Commit(ctx)

		if err != nil {
			return fmt.Errorf("failed to commit transaction for alliance: %w", err)
		}
	}

	slog.Info("deleted an alliance", "allianceId", allianceId)

	return nil
}

func (m *Manager) CreateAlliance(ctx context.Context, name string, description string, badge int32, allianceType int32, requiredTrophies int32, creator *core.Player) error {
	tx, err := m.pool.Begin(ctx)

	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback(ctx)
			panic(r)
		}

		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var newId int64

	err = tx.QueryRow(
		ctx,
		"insert into alliances (name, description, badge_id, type, required_trophies, total_trophies, creator_id, region) values ($1, $2, $3, $4, $5, $6, $7, $8) returning id",
		name, description, badge, allianceType, requiredTrophies, creator.Trophies, creator.DbId, creator.Region).Scan(&newId)

	if err != nil {
		return fmt.Errorf("failed to create alliance: %w", err)
	}

	_, err = tx.Exec(
		ctx,
		"insert into alliance_members (alliance_id, player_id, role) values ($1, $2, $3)",
		newId, creator.DbId, 2,
	)

	if err != nil {
		return fmt.Errorf("failed to insert alliance member: %w", err)
	}

	err = tx.Commit(ctx)

	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	creator.AllianceId = new(int64)
	*creator.AllianceId = newId

	creator.AllianceRole = 2

	slog.Info("created an alliance", "allianceId", newId)

	return nil
}

func (m *Manager) GetPlayerTrophyLeaderboard(ctx context.Context, limit int, region *string) ([]LeaderboardPlayerEntry, error) {
	query := ""
	
	var rows pgx.Rows
	var err error
	
	if region == nil {
		query = `
	        select
	            p.id, p.name, pp.trophies, p.profile_icon, p.region, p.high_id, p.low_id, pp.experience,
	            am.alliance_id
	        from players p
	        join player_progression pp on p.id = pp.player_id
	        left join alliance_members am on p.id = am.player_id
	        order by pp.trophies desc
	        limit $1`
	
		rows, err = m.pool.Query(ctx, query, limit)
	} else {
		query = `
	        select
	            p.id, p.name, pp.trophies, p.profile_icon, p.region, p.high_id, p.low_id, pp.experience,
	            am.alliance_id
	        from players p
	        join player_progression pp on p.id = pp.player_id
	        left join alliance_members am on p.id = am.player_id
			where p.region = $2
	        order by pp.trophies desc
	        limit $1`
	
		rows, err = m.pool.Query(ctx, query, limit, *region)
	}

	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	defer rows.Close()

	entries := make([]LeaderboardPlayerEntry, 0, limit)

	for rows.Next() {
		var entry LeaderboardPlayerEntry
		var allianceId sql.NullInt64

		err = rows.Scan(
			&entry.DbId, &entry.Name, &entry.Trophies, &entry.ProfileIcon, &entry.Region, &entry.PlayerHighId, &entry.PlayerLowId, &entry.PlayerExperience, &allianceId,
		)

		if err != nil {
			slog.Error("scan failed", "error", err)
			continue
		}
		
		if allianceId.Valid {
			entry.AllianceId = &allianceId.Int64
		} else {
			entry.AllianceId = nil
		}

		entries = append(entries, entry)
	}

	if err = rows.Err(); err != nil {
		return entries, fmt.Errorf("row iteration error: %w", err)
	}

	return entries, nil
}

func (m *Manager) GetBrawlerTrophyLeaderboard(ctx context.Context, brawlerId int32, limit int, region *string) ([]LeaderboardPlayerEntry, error) {
	query := ""
	
	var rows pgx.Rows
	var err error
	
	if region == nil {
		query = `
	        select
	            p.id, p.name, pb.trophies, p.profile_icon, p.region, p.high_id, p.low_id, pp.experience,
	            am.alliance_id
	        from players p
	        join player_brawlers pb on p.id = pb.player_id
	        join player_progression pp on p.id = pp.player_id
	        left join alliance_members am on p.id = am.player_id
	        where pb.brawler_id = $1
	        order by pb.trophies desc
	        limit $2`
		rows, err = m.pool.Query(ctx, query, brawlerId, limit)
	} else {
		query = `
	        select
	            p.id, p.name, pb.trophies, p.profile_icon, p.region, p.high_id, p.low_id, pp.experience,
	            am.alliance_id
	        from players p
	        join player_brawlers pb on p.id = pb.player_id
	        join player_progression pp on p.id = pp.player_id
	        left join alliance_members am on p.id = am.player_id
	        where pb.brawler_id = $1 and p.region = $3
	        order by pb.trophies desc
	        limit $2`
		rows, err = m.pool.Query(ctx, query, brawlerId, limit, *region)
	}

	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	defer rows.Close()

	entries := make([]LeaderboardPlayerEntry, 0, limit)

	for rows.Next() {
		var entry LeaderboardPlayerEntry
		var allianceId sql.NullInt64

		err = rows.Scan(
			&entry.DbId, &entry.Name, &entry.Trophies, &entry.ProfileIcon, &entry.Region, &entry.PlayerHighId, &entry.PlayerLowId, &entry.PlayerExperience, &allianceId,
		)

		if err != nil {
			slog.Error("scan failed", "error", err, "brawlerID", brawlerId)
			continue
		}
		
		if allianceId.Valid {
			entry.AllianceId = &allianceId.Int64
		} else {
			entry.AllianceId = nil
		}

		entries = append(entries, entry)
	}

	if err = rows.Err(); err != nil {
		return entries, fmt.Errorf("row iteration error: %w", err)
	}

	return entries, nil
}

func (m *Manager) GetAllianceTrophyLeaderboard(ctx context.Context, limit int, region *string) ([]LeaderboardAllianceEntry, error) {
	query := ""
	
	var rows pgx.Rows
	var err error
	
	if region == nil {
		query = `
	        select
				a.id, a.name, a.badge_id, a.type,
				a.total_trophies,
				coalesce(mc.member_count, 0) as current_member_count
			from alliances a
			left join (
				select alliance_id, count(*) as member_count
				from alliance_members
				group by alliance_id
			) mc on a.id = mc.alliance_id
			order by a.total_trophies
			limit $1`
		rows, err = m.pool.Query(ctx, query, limit)
	} else {
		query = `
	        select
				a.id, a.name, a.badge_id, a.type,
				a.total_trophies,
				coalesce(mc.member_count, 0) as current_member_count
			from alliances a
			left join (
				select alliance_id, count(*) as member_count
				from alliance_members
				group by alliance_id
			) mc on a.id = mc.alliance_id
			where a.region = $2
			order by a.total_trophies
			limit $1`
		rows, err = m.pool.Query(ctx, query, limit, *region)
	}

	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	defer rows.Close()

	entries := make([]LeaderboardAllianceEntry, 0, limit)

	for rows.Next() {
		var entry LeaderboardAllianceEntry


		err = rows.Scan(
			&entry.DbId, &entry.Name, &entry.BadgeId, &entry.Type, &entry.TotalTrophies,
			&entry.TotalMembers,
		)

		if err != nil {
			slog.Error("scan failed", "err", err)
			continue
		}
		
		entries = append(entries, entry)
	}

	if err = rows.Err(); err != nil {
		return entries, fmt.Errorf("row iteration error: %w", err)
	}

	return entries, nil
}

func reverseMessages(s []core.AllianceMessage) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
