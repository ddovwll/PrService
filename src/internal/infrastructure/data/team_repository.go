package data

import (
	"context"

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
	q := querierFromContext(ctx, r.pool)

	const query = `
		INSERT INTO teams (name)
		VALUES ($1)
	`

	_, err := q.Exec(ctx, query, name)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrTeamAlreadyExists
		}
		return err
	}

	return nil
}

func (r *TeamRepository) GetByName(ctx context.Context, name domain.TeamName) (*domain.Team, error) {
	q := querierFromContext(ctx, r.pool)

	const teamQuery = `
		SELECT name
		FROM teams
		WHERE name = $1
	`

	var t domain.Team
	if err := q.QueryRow(ctx, teamQuery, name).Scan(&t.Name); err != nil {
		if isNoRows(err) {
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
	q := querierFromContext(ctx, r.pool)

	const query = `
		SELECT team_name
		FROM users
		WHERE id = $1
	`

	var teamName domain.TeamName
	if err := q.QueryRow(ctx, query, userID).Scan(&teamName); err != nil {
		if isNoRows(err) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	return r.GetByName(ctx, teamName)
}
