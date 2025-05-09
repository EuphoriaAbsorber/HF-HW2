package repository

import (
	"context"
	model "main/internal/model"
)

type RepositoryInterface interface {
	CreateMarble(ctx context.Context, asset *model.Marble) error
}

type Service struct {
	Repository RepositoryInterface
}

func NewService() *Service {
	return &Service{
		Repository: newRepository(),
	}
}
