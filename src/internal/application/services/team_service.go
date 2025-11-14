package services

import (
	"context"

	"PrService/src/internal/application/contracts"

	"PrService/src/internal/domain"
)

type TeamService struct {
	teamRepository domain.TeamRepository
	userRepository domain.UserRepository
	txManager      contracts.TxManager
}

func NewTeamService(
	teamRepository domain.TeamRepository,
	userRepository domain.UserRepository,
	txManager contracts.TxManager,
) *TeamService {
	return &TeamService{
		teamRepository: teamRepository,
		userRepository: userRepository,
		txManager:      txManager,
	}
}

func (s *TeamService) Create(ctx context.Context, team *domain.Team) (*domain.Team, error) {
	err := s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := s.teamRepository.Create(txCtx, team.Name); err != nil {
			return err
		}

		users := make([]domain.User, 0, len(team.Members))
		for _, member := range team.Members {
			user := member.ToUser(team.Name)
			users = append(users, user)
		}

		return s.userRepository.UpsertBatch(txCtx, users)
	})

	if err != nil {
		return nil, err
	}

	return team, nil
}

func (s *TeamService) Get(ctx context.Context, name domain.TeamName) (*domain.Team, error) {
	return s.teamRepository.GetByName(ctx, name)
}
