package logic

import (
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
