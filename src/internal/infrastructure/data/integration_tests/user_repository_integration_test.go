//go:build integration

package integration_tests

import (
	"PrService/src/internal/infrastructure/data/repositories"
	"context"
	"errors"
	"testing"

	"PrService/src/internal/domain"
)

func TestUserRepository_UpsertBatch_InsertAndGetByID(t *testing.T) {
	ctx := context.Background()
	truncateAll(t, ctx)

	repo := repositories.NewUserRepository(testPool)

	teamName := domain.TeamName("backend")
	insertTeam(t, ctx, teamName)

	user := domain.User{
		ID:       domain.UserID("u1"),
		Username: "Alice",
		TeamName: teamName,
		IsActive: true,
	}

	if err := repo.UpsertBatch(ctx, []domain.User{user}); err != nil {
		t.Fatalf("UpsertBatch(insert) failed: %v", err)
	}

	got, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}

	if got.ID != user.ID {
		t.Errorf("expected ID %q, got %q", user.ID, got.ID)
	}
	if got.Username != user.Username {
		t.Errorf("expected Username %q, got %q", user.Username, got.Username)
	}
	if got.TeamName != user.TeamName {
		t.Errorf("expected TeamName %q, got %q", user.TeamName, got.TeamName)
	}
	if got.IsActive != user.IsActive {
		t.Errorf("expected IsActive %v, got %v", user.IsActive, got.IsActive)
	}
}

func TestUserRepository_UpsertBatch_UpdateExisting(t *testing.T) {
	ctx := context.Background()
	truncateAll(t, ctx)

	repo := repositories.NewUserRepository(testPool)

	teamName := domain.TeamName("backend")
	insertTeam(t, ctx, teamName)

	original := domain.User{
		ID:       domain.UserID("u1"),
		Username: "Alice",
		TeamName: teamName,
		IsActive: true,
	}

	if err := repo.UpsertBatch(ctx, []domain.User{original}); err != nil {
		t.Fatalf("UpsertBatch(insert) failed: %v", err)
	}

	updated := domain.User{
		ID:       original.ID,
		Username: "AliceUpdated",
		TeamName: teamName,
		IsActive: false,
	}

	if err := repo.UpsertBatch(ctx, []domain.User{updated}); err != nil {
		t.Fatalf("UpsertBatch(update) failed: %v", err)
	}

	got, err := repo.GetByID(ctx, updated.ID)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}

	if got.Username != updated.Username {
		t.Errorf("expected Username %q, got %q", updated.Username, got.Username)
	}
	if got.IsActive != updated.IsActive {
		t.Errorf("expected IsActive %v, got %v", updated.IsActive, got.IsActive)
	}
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	ctx := context.Background()
	truncateAll(t, ctx)

	repo := repositories.NewUserRepository(testPool)

	_, err := repo.GetByID(ctx, domain.UserID("non-existent"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserRepository_Update_Success(t *testing.T) {
	ctx := context.Background()
	truncateAll(t, ctx)

	repo := repositories.NewUserRepository(testPool)

	teamName := domain.TeamName("backend")
	insertTeam(t, ctx, teamName)

	user := domain.User{
		ID:       domain.UserID("u1"),
		Username: "Alice",
		TeamName: teamName,
		IsActive: true,
	}

	if err := repo.UpsertBatch(ctx, []domain.User{user}); err != nil {
		t.Fatalf("UpsertBatch(insert) failed: %v", err)
	}

	user.Username = "Bob"
	user.IsActive = false

	if err := repo.Update(ctx, &user); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	got, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}

	if got.Username != "Bob" {
		t.Errorf("expected Username %q, got %q", "Bob", got.Username)
	}
	if got.IsActive != false {
		t.Errorf("expected IsActive %v, got %v", false, got.IsActive)
	}
}

func TestUserRepository_Update_NotFound(t *testing.T) {
	ctx := context.Background()
	truncateAll(t, ctx)

	repo := repositories.NewUserRepository(testPool)

	user := domain.User{
		ID:       domain.UserID("ghost"),
		Username: "Ghost",
		TeamName: domain.TeamName("backend"),
		IsActive: true,
	}

	err := repo.Update(ctx, &user)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}
