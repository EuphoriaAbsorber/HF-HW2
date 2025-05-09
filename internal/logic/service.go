package logic

import (
	"context"
	"fmt"
	"main/internal/model"
	rep "main/internal/repository"
)

type Service struct {
	repository *rep.Service
}

func NewService(repository *rep.Service) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) CreateMarble(ctx context.Context, marble *model.Marble) error {
	// item := &model.Asset{
	// 	ID:             id,
	// 	AppraisedValue: appraisedValue,
	// 	Color:          color,
	// 	Size:           size,
	// 	Owner:          owner,
	// }

	err := s.repository.Repository.CreateMarble(ctx, marble)
	if err != nil {
		return fmt.Errorf("failed to create marble: %v", err)
	}

	return nil
}
