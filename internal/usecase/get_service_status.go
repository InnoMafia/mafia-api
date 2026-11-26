package usecase

import (
	"context"
	"time"
)

type StatusRepository interface {
	Ping(context.Context) error
}

type GetServiceStatus struct {
	repository StatusRepository
}

func NewGetServiceStatus(repository StatusRepository) *GetServiceStatus {
	return &GetServiceStatus{repository: repository}
}

func (u *GetServiceStatus) Execute(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return u.repository.Ping(ctx) == nil
}
