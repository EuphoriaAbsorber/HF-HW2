package repository

import (
	"context"
	model "main/internal/model"
)

type RepositoryRealization struct {
}

func newRepository() RepositoryInterface {
	return &RepositoryRealization{}
}

func (ret *RepositoryRealization) CreateMarble(ctx context.Context, asset *model.Marble) error {
	return nil

}
