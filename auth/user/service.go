package user

import (
	"context"
	"crypto/sha1"
	"fmt"
)

type UseCase interface {
	ValidateUser(ctx context.Context, email, password string) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}
func (s *Service) ValidateUser(ctx context.Context, email, password string) error {
	u, err := s.repo.Get(ctx, email)
	if err != nil {
		return err
	}
	if u == nil {
		err := fmt.Errorf("invalid user")
		return err
	}
	return s.ValidatePassword(ctx, u, password)
}

func (s *Service) ValidatePassword(ctx context.Context, u *User, password string) error {
	h := sha1.New()
	h.Write([]byte(password))
	p := fmt.Sprintf("%x", h.Sum(nil))
	if p != u.Password {
		err := fmt.Errorf("invalid password")
		return err
	}
	return nil
}
