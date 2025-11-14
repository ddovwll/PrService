//go:build integration

package data

import (
	"PrService/src/internal/domain"
	"context"
	"errors"
	"testing"
)

func TestTeamRepository_Create_And_GetByName(t *testing.T) {
	ctx := context.Background()
	truncateAll(t, ctx)

	repo := NewTeamRepository(testPool)

	teamName := domain.TeamName("backend")

	if err := repo.Create(ctx, teamName); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	err := repo.Create(ctx, teamName)
	if !errors.Is(err, domain.ErrTeamAlreadyExists) {
		t.Fatalf("expected ErrTeamAlreadyExists, got: %v", err)
	}

	insertUser(t, ctx, domain.User{
		ID:       domain.UserID("u1"),
		Username: "Alice",
		TeamName: teamName,
		IsActive: true,
	})
	insertUser(t, ctx, domain.User{
		ID:       domain.UserID("u2"),
		Username: "Bob",
		TeamName: teamName,
		IsActive: false,
	})

	team, err := repo.GetByName(ctx, teamName)
	if err != nil {
		t.Fatalf("GetByName returned error: %v", err)
	}

	if team.Name != teamName {
		t.Fatalf("expected team name %q, got %q", teamName, team.Name)
	}

	if len(team.Members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(team.Members))
	}

	membersByID := make(map[domain.UserID]domain.TeamMember, len(team.Members))
	for _, m := range team.Members {
		membersByID[m.ID] = m
	}

	alice, ok := membersByID["u1"]
	if !ok {
		t.Fatalf("expected member u1 in team members")
	}
	if alice.Username != "Alice" || !alice.IsActive {
		t.Fatalf("unexpected data for u1: %+v", alice)
	}

	bob, ok := membersByID["u2"]
	if !ok {
		t.Fatalf("expected member u2 in team members")
	}
	if bob.Username != "Bob" || bob.IsActive {
		t.Fatalf("unexpected data for u2: %+v", bob)
	}
}

func TestTeamRepository_GetByName_NotFound(t *testing.T) {
	ctx := context.Background()
	truncateAll(t, ctx)

	repo := NewTeamRepository(testPool)

	_, err := repo.GetByName(ctx, domain.TeamName("unknown"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrTeamNotFound) {
		t.Fatalf("expected ErrTeamNotFound, got %v", err)
	}
}

func TestTeamRepository_GetByUserID_Success(t *testing.T) {
	ctx := context.Background()
	truncateAll(t, ctx)

	repo := NewTeamRepository(testPool)

	teamName := domain.TeamName("backend")
	insertTeam(t, ctx, teamName)

	insertUser(t, ctx, domain.User{
		ID:       "u1",
		Username: "Alice",
		TeamName: teamName,
		IsActive: true,
	})
	insertUser(t, ctx, domain.User{
		ID:       "u2",
		Username: "Bob",
		TeamName: teamName,
		IsActive: true,
	})

	team, err := repo.GetByUserID(ctx, domain.UserID("u1"))
	if err != nil {
		t.Fatalf("GetByUserID returned error: %v", err)
	}

	if team.Name != teamName {
		t.Fatalf("expected team name %q, got %q", teamName, team.Name)
	}

	if len(team.Members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(team.Members))
	}
}

func TestTeamRepository_GetByUserID_UserNotFound(t *testing.T) {
	ctx := context.Background()
	truncateAll(t, ctx)

	repo := NewTeamRepository(testPool)

	_, err := repo.GetByUserID(ctx, domain.UserID("ghost"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}
