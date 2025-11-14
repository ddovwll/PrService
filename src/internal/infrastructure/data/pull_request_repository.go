package data

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"PrService/src/internal/domain"
)

type PullRequestRepository struct {
	pool *pgxpool.Pool
}

func NewPullRequestRepository(pool *pgxpool.Pool) *PullRequestRepository {
	return &PullRequestRepository{pool: pool}
}

func (r *PullRequestRepository) Create(ctx context.Context, pr *domain.PullRequest) error {
	q := querierFromContext(ctx, r.pool)

	const insertPR = `
		INSERT INTO pull_requests (
			id, name, author_id, status, created_at, merged_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := q.Exec(ctx, insertPR,
		pr.ID,
		pr.Name,
		pr.AuthorID,
		pr.Status,
		pr.CreatedAt,
		pr.MergedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrPullRequestExists
		}
		return err
	}

	if len(pr.AssignedReviewers) > 0 {
		if err := r.replaceAssignedReviewers(ctx, q, pr.ID, pr.AssignedReviewers); err != nil {
			return err
		}
	}

	return nil
}

func (r *PullRequestRepository) GetByID(ctx context.Context, id domain.PullRequestID) (*domain.PullRequest, error) {
	q := querierFromContext(ctx, r.pool)

	const query = `
		SELECT id, name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE id = $1
	`

	var pr domain.PullRequest
	if err := q.QueryRow(ctx, query, id).Scan(
		&pr.ID,
		&pr.Name,
		&pr.AuthorID,
		&pr.Status,
		&pr.CreatedAt,
		&pr.MergedAt,
	); err != nil {
		if isNoRows(err) {
			return nil, domain.ErrPullRequestNotFound
		}
		return nil, err
	}

	reviewers, err := r.loadAssignedReviewers(ctx, q, pr.ID)
	if err != nil {
		return nil, err
	}
	pr.AssignedReviewers = reviewers

	return &pr, nil
}

func (r *PullRequestRepository) ListByReviewer(
	ctx context.Context,
	reviewerID domain.UserID,
) ([]domain.PullRequest, error) {
	q := querierFromContext(ctx, r.pool)

	const query = `
		SELECT pr.id, pr.name, pr.author_id, pr.status, pr.created_at, pr.merged_at
		FROM pull_requests pr
		JOIN pull_request_reviewers prr
		  ON prr.pull_request_id = pr.id
		WHERE prr.reviewer_id = $1
	`

	rows, err := q.Query(ctx, query, reviewerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.PullRequest

	for rows.Next() {
		var pr domain.PullRequest
		if err := rows.Scan(
			&pr.ID,
			&pr.Name,
			&pr.AuthorID,
			&pr.Status,
			&pr.CreatedAt,
			&pr.MergedAt,
		); err != nil {
			return nil, err
		}

		reviewers, err := r.loadAssignedReviewers(ctx, q, pr.ID)
		if err != nil {
			return nil, err
		}
		pr.AssignedReviewers = reviewers

		result = append(result, pr)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return result, nil
}

func (r *PullRequestRepository) Update(ctx context.Context, pr *domain.PullRequest) error {
	q := querierFromContext(ctx, r.pool)

	const updatePR = `
		UPDATE pull_requests
		SET
			name      = $2,
			author_id = $3,
			status    = $4,
			created_at = $5,
			merged_at  = $6
		WHERE id = $1
	`

	tag, err := q.Exec(ctx, updatePR,
		pr.ID,
		pr.Name,
		pr.AuthorID,
		pr.Status,
		pr.CreatedAt,
		pr.MergedAt,
	)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrPullRequestNotFound
	}

	return r.replaceAssignedReviewers(ctx, q, pr.ID, pr.AssignedReviewers)
}

func (r *PullRequestRepository) replaceAssignedReviewers(
	ctx context.Context,
	q pgxQuerier,
	prID domain.PullRequestID,
	reviewers []domain.UserID,
) error {
	const deleteQuery = `
		DELETE FROM pull_request_reviewers
		WHERE pull_request_id = $1
	`

	if _, err := q.Exec(ctx, deleteQuery, prID); err != nil {
		return err
	}

	if len(reviewers) == 0 {
		return nil
	}

	const insertQuery = `
		INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id)
		VALUES ($1, $2)
	`

	for _, reviewerID := range reviewers {
		if _, err := q.Exec(ctx, insertQuery, prID, reviewerID); err != nil {
			return err
		}
	}

	return nil
}

func (r *PullRequestRepository) loadAssignedReviewers(
	ctx context.Context,
	q pgxQuerier,
	prID domain.PullRequestID,
) ([]domain.UserID, error) {
	const query = `
		SELECT reviewer_id
		FROM pull_request_reviewers
		WHERE pull_request_id = $1
	`

	rows, err := q.Query(ctx, query, prID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviewers []domain.UserID
	for rows.Next() {
		var id domain.UserID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, id)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return reviewers, nil
}
