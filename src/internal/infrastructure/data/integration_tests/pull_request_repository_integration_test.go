//go:build integration

package integration_tests

import (
	"PrService/src/internal/infrastructure/data/repositories"
	"context"
	"errors"
	"testing"

	"PrService/src/internal/domain"
)

func TestPullRequestRepository_Create_And_GetByID_WithReviewers(t *testing.T) {
	ctx := context.Background()
	truncateAll(t, ctx)

	repo := repositories.NewPullRequestRepository(testPool)

	teamName := domain.TeamName("backend")
	insertTeam(t, ctx, teamName)

	author := domain.User{
		ID:       domain.UserID("author1"),
		Username: "Author",
		TeamName: teamName,
		IsActive: true,
	}
	reviewer1 := domain.User{
		ID:       domain.UserID("r1"),
		Username: "Reviewer1",
		TeamName: teamName,
		IsActive: true,
	}
	reviewer2 := domain.User{
		ID:       domain.UserID("r2"),
		Username: "Reviewer2",
		TeamName: teamName,
		IsActive: true,
	}

	insertUser(t, ctx, author)
	insertUser(t, ctx, reviewer1)
	insertUser(t, ctx, reviewer2)

	pr := &domain.PullRequest{
		ID:                domain.PullRequestID("pr-1"),
		Name:              "Add feature",
		AuthorID:          author.ID,
		Status:            domain.PullRequestStatusOpen,
		AssignedReviewers: []domain.UserID{reviewer1.ID, reviewer2.ID},
		CreatedAt:         nil,
		MergedAt:          nil,
	}

	if err := repo.Create(ctx, pr); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	got, err := repo.GetByID(ctx, pr.ID)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}

	if got.ID != pr.ID {
		t.Errorf("expected ID %q, got %q", pr.ID, got.ID)
	}
	if got.Name != pr.Name {
		t.Errorf("expected Name %q, got %q", pr.Name, got.Name)
	}
	if got.AuthorID != pr.AuthorID {
		t.Errorf("expected AuthorID %q, got %q", pr.AuthorID, got.AuthorID)
	}
	if got.Status != pr.Status {
		t.Errorf("expected Status %q, got %q", pr.Status, got.Status)
	}

	if len(got.AssignedReviewers) != 2 {
		t.Fatalf("expected 2 reviewers, got %d", len(got.AssignedReviewers))
	}

	set := map[domain.UserID]struct{}{}
	for _, rID := range got.AssignedReviewers {
		set[rID] = struct{}{}
	}
	if _, ok := set[reviewer1.ID]; !ok {
		t.Errorf("expected reviewer %q in AssignedReviewers", reviewer1.ID)
	}
	if _, ok := set[reviewer2.ID]; !ok {
		t.Errorf("expected reviewer %q in AssignedReviewers", reviewer2.ID)
	}
}

func TestPullRequestRepository_Create_NoReviewers(t *testing.T) {
	ctx := context.Background()
	truncateAll(t, ctx)

	repo := repositories.NewPullRequestRepository(testPool)

	teamName := domain.TeamName("backend")
	insertTeam(t, ctx, teamName)

	author := domain.User{
		ID:       domain.UserID("author1"),
		Username: "Author",
		TeamName: teamName,
		IsActive: true,
	}
	insertUser(t, ctx, author)

	pr := &domain.PullRequest{
		ID:                domain.PullRequestID("pr-2"),
		Name:              "Add search",
		AuthorID:          author.ID,
		Status:            domain.PullRequestStatusOpen,
		AssignedReviewers: nil,
		CreatedAt:         nil,
		MergedAt:          nil,
	}

	if err := repo.Create(ctx, pr); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	got, err := repo.GetByID(ctx, pr.ID)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}

	if len(got.AssignedReviewers) != 0 {
		t.Fatalf("expected 0 reviewers, got %d", len(got.AssignedReviewers))
	}
}

func TestPullRequestRepository_Create_DuplicateID_ReturnsErrPullRequestExists(t *testing.T) {
	ctx := context.Background()
	truncateAll(t, ctx)

	repo := repositories.NewPullRequestRepository(testPool)

	teamName := domain.TeamName("backend")
	insertTeam(t, ctx, teamName)

	author := domain.User{
		ID:       domain.UserID("author1"),
		Username: "Author",
		TeamName: teamName,
		IsActive: true,
	}
	insertUser(t, ctx, author)

	pr := &domain.PullRequest{
		ID:                domain.PullRequestID("pr-dup"),
		Name:              "First",
		AuthorID:          author.ID,
		Status:            domain.PullRequestStatusOpen,
		AssignedReviewers: nil,
	}

	if err := repo.Create(ctx, pr); err != nil {
		t.Fatalf("Create (first) failed: %v", err)
	}

	err := repo.Create(ctx, pr)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrPullRequestExists) {
		t.Fatalf("expected ErrPullRequestExists, got %v", err)
	}
}

func TestPullRequestRepository_GetByID_NotFound(t *testing.T) {
	ctx := context.Background()
	truncateAll(t, ctx)

	repo := repositories.NewPullRequestRepository(testPool)

	_, err := repo.GetByID(ctx, domain.PullRequestID("no-such-pr"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrPullRequestNotFound) {
		t.Fatalf("expected ErrPullRequestNotFound, got %v", err)
	}
}

func TestPullRequestRepository_ListByReviewer(t *testing.T) {
	ctx := context.Background()
	truncateAll(t, ctx)

	repo := repositories.NewPullRequestRepository(testPool)

	teamName := domain.TeamName("backend")
	insertTeam(t, ctx, teamName)

	author := domain.User{
		ID:       domain.UserID("author1"),
		Username: "Author",
		TeamName: teamName,
		IsActive: true,
	}
	r1 := domain.User{
		ID:       domain.UserID("r1"),
		Username: "Reviewer1",
		TeamName: teamName,
		IsActive: true,
	}
	r2 := domain.User{
		ID:       domain.UserID("r2"),
		Username: "Reviewer2",
		TeamName: teamName,
		IsActive: true,
	}
	r3 := domain.User{
		ID:       domain.UserID("r3"),
		Username: "Reviewer3",
		TeamName: teamName,
		IsActive: true,
	}

	insertUser(t, ctx, author)
	insertUser(t, ctx, r1)
	insertUser(t, ctx, r2)
	insertUser(t, ctx, r3)

	pr1 := &domain.PullRequest{
		ID:                "pr-1",
		Name:              "PR1",
		AuthorID:          author.ID,
		Status:            domain.PullRequestStatusOpen,
		AssignedReviewers: []domain.UserID{r1.ID, r2.ID},
	}
	pr2 := &domain.PullRequest{
		ID:                "pr-2",
		Name:              "PR2",
		AuthorID:          author.ID,
		Status:            domain.PullRequestStatusOpen,
		AssignedReviewers: []domain.UserID{r1.ID},
	}
	pr3 := &domain.PullRequest{
		ID:                "pr-3",
		Name:              "PR3",
		AuthorID:          author.ID,
		Status:            domain.PullRequestStatusOpen,
		AssignedReviewers: []domain.UserID{r2.ID},
	}

	if err := repo.Create(ctx, pr1); err != nil {
		t.Fatalf("Create pr1 failed: %v", err)
	}
	if err := repo.Create(ctx, pr2); err != nil {
		t.Fatalf("Create pr2 failed: %v", err)
	}
	if err := repo.Create(ctx, pr3); err != nil {
		t.Fatalf("Create pr3 failed: %v", err)
	}

	list, err := repo.ListByReviewer(ctx, r1.ID)
	if err != nil {
		t.Fatalf("ListByReviewer returned error: %v", err)
	}

	if len(list) != 2 {
		t.Fatalf("expected 2 PRs for reviewer %s, got %d", r1.ID, len(list))
	}

	byID := make(map[domain.PullRequestID]domain.PullRequest)
	for _, pr := range list {
		byID[pr.ID] = pr
	}

	got1, ok := byID["pr-1"]
	if !ok {
		t.Fatalf("expected pr-1 in result")
	}
	set1 := make(map[domain.UserID]struct{})
	for _, rid := range got1.AssignedReviewers {
		set1[rid] = struct{}{}
	}
	if _, ok := set1[r1.ID]; !ok {
		t.Errorf("expected pr-1 to have reviewer %s", r1.ID)
	}
	if _, ok := set1[r2.ID]; !ok {
		t.Errorf("expected pr-1 to have reviewer %s", r2.ID)
	}

	got2, ok := byID["pr-2"]
	if !ok {
		t.Fatalf("expected pr-2 in result")
	}
	if len(got2.AssignedReviewers) != 1 || got2.AssignedReviewers[0] != r1.ID {
		t.Errorf("expected pr-2 AssignedReviewers = [%s], got %v", r1.ID, got2.AssignedReviewers)
	}
}

func TestPullRequestRepository_Update_Success_ReplaceReviewers(t *testing.T) {
	ctx := context.Background()
	truncateAll(t, ctx)

	repo := repositories.NewPullRequestRepository(testPool)

	teamName := domain.TeamName("backend")
	insertTeam(t, ctx, teamName)

	author := domain.User{
		ID:       domain.UserID("author1"),
		Username: "Author",
		TeamName: teamName,
		IsActive: true,
	}
	r1 := domain.User{
		ID:       domain.UserID("r1"),
		Username: "Reviewer1",
		TeamName: teamName,
		IsActive: true,
	}
	r2 := domain.User{
		ID:       domain.UserID("r2"),
		Username: "Reviewer2",
		TeamName: teamName,
		IsActive: true,
	}
	r3 := domain.User{
		ID:       domain.UserID("r3"),
		Username: "Reviewer3",
		TeamName: teamName,
		IsActive: true,
	}

	insertUser(t, ctx, author)
	insertUser(t, ctx, r1)
	insertUser(t, ctx, r2)
	insertUser(t, ctx, r3)

	pr := &domain.PullRequest{
		ID:                "pr-upd",
		Name:              "Initial",
		AuthorID:          author.ID,
		Status:            domain.PullRequestStatusOpen,
		AssignedReviewers: []domain.UserID{r1.ID, r2.ID},
	}

	if err := repo.Create(ctx, pr); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	pr.Status = domain.PullRequestStatusMerged
	pr.AssignedReviewers = []domain.UserID{r2.ID, r3.ID}

	if err := repo.Update(ctx, pr); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	got, err := repo.GetByID(ctx, pr.ID)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}

	if got.Status != domain.PullRequestStatusMerged {
		t.Errorf("expected Status %q, got %q", domain.PullRequestStatusMerged, got.Status)
	}

	if len(got.AssignedReviewers) != 2 {
		t.Fatalf("expected 2 reviewers after update, got %d", len(got.AssignedReviewers))
	}

	set := make(map[domain.UserID]struct{})
	for _, rid := range got.AssignedReviewers {
		set[rid] = struct{}{}
	}

	if _, ok := set[r2.ID]; !ok {
		t.Errorf("expected updated reviewers to contain %s", r2.ID)
	}
	if _, ok := set[r3.ID]; !ok {
		t.Errorf("expected updated reviewers to contain %s", r3.ID)
	}
	if _, ok := set[r1.ID]; ok {
		t.Errorf("expected updated reviewers NOT to contain %s", r1.ID)
	}
}

func TestPullRequestRepository_Update_NotFound(t *testing.T) {
	ctx := context.Background()
	truncateAll(t, ctx)

	repo := repositories.NewPullRequestRepository(testPool)

	pr := &domain.PullRequest{
		ID:       "non-existent",
		Name:     "Ghost",
		AuthorID: "ghost-author",
		Status:   domain.PullRequestStatusOpen,
	}

	err := repo.Update(ctx, pr)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrPullRequestNotFound) {
		t.Fatalf("expected ErrPullRequestNotFound, got %v", err)
	}
}
