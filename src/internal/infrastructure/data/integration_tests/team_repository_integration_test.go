//go:build integration

package integration_tests

import (
	"PrService/src/internal/domain"
	"PrService/src/internal/infrastructure/data/repositories"
	"context"
	"errors"
	"testing"
	"time"
)

func TestTeamRepository_Create_And_GetByName(t *testing.T) {
	ctx := context.Background()
	truncateAll(t, ctx)

	repo := repositories.NewTeamRepository(testPool)

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

	repo := repositories.NewTeamRepository(testPool)

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

	repo := repositories.NewTeamRepository(testPool)

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

	repo := repositories.NewTeamRepository(testPool)

	_, err := repo.GetByUserID(ctx, domain.UserID("ghost"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestTeamRepository_GetStats_WithData(t *testing.T) {
	ctx := context.Background()
	truncateAll(t, ctx)

	repo := repositories.NewTeamRepository(testPool)
	teamName := domain.TeamName("backend")

	insertTeam(t, ctx, teamName)

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

	now := time.Now().UTC().Truncate(time.Second)

	_, err := testPool.Exec(ctx, `
        INSERT INTO pull_requests (id, name, author_id, status, created_at, merged_at)
        VALUES
            ($1, $2, $3, 'OPEN',   $4, NULL),
            ($5, $6, $7, 'MERGED', $8, $9)
    `,
		"pr-open", "Open PR", "u1", now,
		"pr-merged", "Merged PR", "u1", now.Add(-3600*time.Second), now,
	)
	if err != nil {
		t.Fatalf("failed to insert pull requests: %v", err)
	}

	stats, err := repo.GetStats(ctx, teamName)
	if err != nil {
		t.Fatalf("GetStats returned error: %v", err)
	}

	if stats.TeamName != teamName {
		t.Fatalf("expected team_name %q, got %q", teamName, stats.TeamName)
	}
	if stats.MembersCount != 2 {
		t.Fatalf("expected MembersCount=2, got %d", stats.MembersCount)
	}
	if stats.ActiveMembersCount != 1 {
		t.Fatalf("expected ActiveMembersCount=1, got %d", stats.ActiveMembersCount)
	}
	if stats.TotalPRs != 2 {
		t.Fatalf("expected TotalPRs=2, got %d", stats.TotalPRs)
	}
	if stats.OpenPRs != 1 {
		t.Fatalf("expected OpenPRs=1, got %d", stats.OpenPRs)
	}
	if stats.MergedPRs != 1 {
		t.Fatalf("expected MergedPRs=1, got %d", stats.MergedPRs)
	}
	if stats.AvgTimeToMergeSec != 3600 {
		t.Fatalf("expected AvgTimeToMergeSec=3600, got %d", stats.AvgTimeToMergeSec)
	}
}

func TestTeamRepository_GetStats_NoUsersNoPRs(t *testing.T) {
	ctx := context.Background()
	truncateAll(t, ctx)

	repo := repositories.NewTeamRepository(testPool)
	teamName := domain.TeamName("empty-team")

	insertTeam(t, ctx, teamName)

	stats, err := repo.GetStats(ctx, teamName)
	if err != nil {
		t.Fatalf("GetStats returned error: %v", err)
	}

	if stats.MembersCount != 0 ||
		stats.ActiveMembersCount != 0 ||
		stats.TotalPRs != 0 ||
		stats.OpenPRs != 0 ||
		stats.MergedPRs != 0 ||
		stats.AvgTimeToMergeSec != 0 {
		t.Fatalf("expected all zero stats, got %+v", stats)
	}
}

func TestTeamRepository_GetStats_TeamNotFound(t *testing.T) {
	ctx := context.Background()
	truncateAll(t, ctx)

	repo := repositories.NewTeamRepository(testPool)

	_, err := repo.GetStats(ctx, domain.TeamName("unknown"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrTeamNotFound) {
		t.Fatalf("expected ErrTeamNotFound, got %v", err)
	}
}
