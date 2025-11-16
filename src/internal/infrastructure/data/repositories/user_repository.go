package repositories

import (
	"context"

	"PrService/src/internal/infrastructure/data"

	"PrService/src/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		pool: pool,
	}
}

func (r *UserRepository) UpsertBatch(ctx context.Context, users []domain.User) error {
	if len(users) == 0 {
		return nil
	}

	q := data.QuerierFromContext(ctx, r.pool)

	const query = `
		INSERT INTO users (id, username, team_name, is_active)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			username  = EXCLUDED.username,
			team_name = EXCLUDED.team_name,
			is_active = EXCLUDED.is_active
	`

	for _, u := range users {
		if _, err := q.Exec(ctx, query,
			u.ID,
			u.Username,
			u.TeamName,
			u.IsActive,
		); err != nil {
			return err
		}
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	q := data.QuerierFromContext(ctx, r.pool)

	const query = `
		SELECT id, username, team_name, is_active
		FROM users
		WHERE id = $1
	`

	var u domain.User
	if err := q.QueryRow(ctx, query, id).Scan(
		&u.ID,
		&u.Username,
		&u.TeamName,
		&u.IsActive,
	); err != nil {
		if data.IsNoRows(err) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	return &u, nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	q := data.QuerierFromContext(ctx, r.pool)

	const query = `
		UPDATE users
		SET username = $2,
		    team_name = $3,
		    is_active = $4
		WHERE id = $1
	`

	tag, err := q.Exec(ctx, query,
		user.ID,
		user.Username,
		user.TeamName,
		user.IsActive,
	)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}
