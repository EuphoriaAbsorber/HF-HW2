package repository

import (
	"context"
	model "main/internal/model"
)

type RepositoryInterface interface {
	CreateMarble(ctx context.Context, marble *model.Marble) error
	ReadMarble(ctx context.Context, name string) ([]byte, error)
	DeleteMarble(ctx context.Context, marbleJSON model.Marble) error
}

type Service struct {
	Repository RepositoryInterface
}

func NewService() *Service {
	return &Service{
		Repository: newRepository(),
	}
}
