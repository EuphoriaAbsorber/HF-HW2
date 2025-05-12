package repository

import (
	"context"
	model "main/internal/model"
	"time"
)

type RepositoryInterface interface {
	CreateMarble(ctx context.Context, marble *model.Marble) error
	ReadMarble(ctx context.Context, name string) ([]byte, error)
	DeleteMarble(ctx context.Context, marbleJSON model.Marble) error
	TransferMarble(ctx context.Context, marbleToTransfer model.Marble) error
	GetMarbleHistory(ctx context.Context, id string) (map[string]model.Marble, map[string]time.Time, error)
}

type Service struct {
	Repository RepositoryInterface
}

func NewService() *Service {
	return &Service{
		Repository: newRepository(),
	}
}
