package auth

import (
	"context"
	"github.com/1ncrease0/pixly/services/auth/internal/domain"
	"github.com/1ncrease0/pixly/services/auth/internal/domain/events"
	"github.com/google/uuid"
)

type UserRepo interface {
	Create(ctx context.Context, user *domain.User) error
	UserByEmail(ctx context.Context, email domain.Email) (*domain.User, error)
	//	UserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

type VerificationRepo interface {
	Save(ctx context.Context, codeHash string, userID uuid.UUID) error
	Get(ctx context.Context, codeHash string) (uuid.UUID, error)
}

type VerificationSender interface {
	SendVerification(ctx context.Context, e events.EmailVerification) error
}

type AuthService struct {
	userRepo         UserRepo
	verificationRepo VerificationRepo
	sender           VerificationSender
}

func NewAuthService(userRepo UserRepo, verificationRepo VerificationRepo, sender VerificationSender) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		verificationRepo: verificationRepo,
		sender:           sender,
	}
}

func (s *AuthService) Register(ctx context.Context, email domain.Email, name domain.Username, password domain.Password) error {
	user := domain.NewUser(uuid.New(), email, name, password, false)
	if err := s.userRepo.Create(ctx, user); err != nil {
		return err
	}

	return s.SendVerification(ctx, user)
}

func (s *AuthService) VerifyEmail(ctx context.Context, code domain.VerificationCode) error {

	return nil
}

func (s *AuthService) SendVerification(ctx context.Context, u *domain.User) error {
	code := domain.NewVerificationCode()

	if err := s.verificationRepo.Save(ctx, code.Hash(), u.ID()); err != nil {
		return err
	}
	e := events.EmailVerification{
		Email: u.Email().String(),
		Code:  code.String(),
	}
	return s.sender.SendVerification(ctx, e)
}
