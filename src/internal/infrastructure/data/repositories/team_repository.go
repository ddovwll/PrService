package repositories

import (
	"context"

	"PrService/src/internal/infrastructure/data"

	"github.com/jackc/pgx/v5/pgxpool"

	"PrService/src/internal/domain"
)

type TeamRepository struct {
	pool *pgxpool.Pool
}

func NewTeamRepository(pool *pgxpool.Pool) *TeamRepository {
	return &TeamRepository{pool: pool}
}

func (r *TeamRepository) Create(ctx context.Context, name domain.TeamName) error {
	q := data.QuerierFromContext(ctx, r.pool)

	const query = `
		INSERT INTO teams (name)
		VALUES ($1)
	`

	_, err := q.Exec(ctx, query, name)
	if err != nil {
		if data.IsUniqueViolation(err) {
			return domain.ErrTeamAlreadyExists
		}
		return err
	}

	return nil
}

func (r *TeamRepository) GetByName(ctx context.Context, name domain.TeamName) (*domain.Team, error) {
	q := data.QuerierFromContext(ctx, r.pool)

	const teamQuery = `
		SELECT name
		FROM teams
		WHERE name = $1
	`

	var t domain.Team
	if err := q.QueryRow(ctx, teamQuery, name).Scan(&t.Name); err != nil {
		if data.IsNoRows(err) {
			return nil, domain.ErrTeamNotFound
		}
		return nil, err
	}

	const membersQuery = `
		SELECT id, username, is_active
		FROM users
		WHERE team_name = $1
	`

	rows, err := q.Query(ctx, membersQuery, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []domain.TeamMember
	for rows.Next() {
		var m domain.TeamMember
		if err := rows.Scan(&m.ID, &m.Username, &m.IsActive); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	t.Members = members

	return &t, nil
}

func (r *TeamRepository) GetByUserID(ctx context.Context, userID domain.UserID) (*domain.Team, error) {
	q := data.QuerierFromContext(ctx, r.pool)

	const query = `
		SELECT team_name
		FROM users
		WHERE id = $1
	`

	var teamName domain.TeamName
	if err := q.QueryRow(ctx, query, userID).Scan(&teamName); err != nil {
		if data.IsNoRows(err) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	return r.GetByName(ctx, teamName)
}

func (r *TeamRepository) GetStats(ctx context.Context, name domain.TeamName) (*domain.TeamStats, error) {
	q := data.QuerierFromContext(ctx, r.pool)

	const checkTeamQuery = `
		SELECT 1
		FROM teams
		WHERE name = $1
	`
	var dummy int
	if err := q.QueryRow(ctx, checkTeamQuery, name).Scan(&dummy); err != nil {
		if data.IsNoRows(err) {
			return nil, domain.ErrTeamNotFound
		}
		return nil, err
	}

	const statsQuery = `
		WITH team_users AS (
			SELECT id, is_active
			FROM users
			WHERE team_name = $1
		),
		team_prs AS (
			SELECT pr.id, pr.status, pr.created_at, pr.merged_at
			FROM pull_requests pr
			JOIN team_users tu ON tu.id = pr.author_id
		)
		SELECT
			(SELECT COUNT(*) FROM team_users)                                                  AS members_count,
			(SELECT COUNT(*) FROM team_users WHERE is_active)                                  AS active_members_count,
			(SELECT COUNT(*) FROM team_prs)                                                    AS total_prs,
			(SELECT COUNT(*) FROM team_prs WHERE status = 'OPEN')                              AS open_prs,
			(SELECT COUNT(*) FROM team_prs WHERE status = 'MERGED')                            AS merged_prs,
			COALESCE(
				(SELECT AVG(EXTRACT(EPOCH FROM (merged_at - created_at))) FROM team_prs WHERE status = 'MERGED'),
				0
			) AS avg_time_to_merge_seconds
	`

	var (
		membersCount       int
		activeMembersCount int
		totalPRs           int
		openPRs            int
		mergedPRs          int
		avgSec             float64
	)

	if err := q.QueryRow(ctx, statsQuery, name).Scan(
		&membersCount,
		&activeMembersCount,
		&totalPRs,
		&openPRs,
		&mergedPRs,
		&avgSec,
	); err != nil {
		return nil, err
	}

	stats := &domain.TeamStats{
		TeamName:           name,
		MembersCount:       membersCount,
		ActiveMembersCount: activeMembersCount,
		TotalPRs:           totalPRs,
		OpenPRs:            openPRs,
		MergedPRs:          mergedPRs,
		AvgTimeToMergeSec:  int64(avgSec),
	}

	return stats, nil
}
