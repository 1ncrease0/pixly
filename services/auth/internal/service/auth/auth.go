package auth

import (
	"context"
	"github.com/1ncrease0/pixly/pkg/jwt"
	"github.com/1ncrease0/pixly/services/auth/internal/domain"
	"github.com/1ncrease0/pixly/services/auth/internal/domain/events"
	"github.com/google/uuid"
	"time"
)

type UserRepo interface {
	Create(ctx context.Context, user *domain.User) error
	UserByEmail(ctx context.Context, email domain.Email) (*domain.User, error)
	UserByID(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	Verify(ctx context.Context, userID uuid.UUID) error
}

type SessionRepo interface {
	CreateSession(ctx context.Context, session *domain.Session) error
	Session(ctx context.Context, tokenHash string) (*domain.Session, error)
	DeleteSession(ctx context.Context, tokenHash string) error
}

type VerificationRepo interface {
	Save(ctx context.Context, codeHash string, userID uuid.UUID) error
	Get(ctx context.Context, codeHash string) (uuid.UUID, error)
	Delete(ctx context.Context, codeHash string) error
}

type VerificationSender interface {
	SendVerification(ctx context.Context, e events.EmailVerification) error
}

type AuthService struct {
	userRepo         UserRepo
	verificationRepo VerificationRepo
	sender           VerificationSender
	sessionRepo      SessionRepo
	jwtManger        *jwt.Manager
}

func NewAuthService(
	userRepo UserRepo,
	verificationRepo VerificationRepo,
	sender VerificationSender,
	sessionRepo SessionRepo,
	m *jwt.Manager,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		verificationRepo: verificationRepo,
		sender:           sender,
		sessionRepo:      sessionRepo,
		jwtManger:        m,
	}
}

func (s *AuthService) Register(ctx context.Context, email domain.Email, name domain.Username, password domain.Password) error {
	user := domain.NewUser(uuid.New(), email, name, password, false)
	if err := s.userRepo.Create(ctx, user); err != nil {
		return err
	}

	return s.sendVerification(ctx, user)
}

func (s *AuthService) VerifyEmail(ctx context.Context, code domain.VerificationCode) error {
	userID, err := s.verificationRepo.Get(ctx, code.Hash())
	if err != nil {
		return err
	}
	if err := s.userRepo.Verify(ctx, userID); err != nil {
		return err
	}

	_ = s.verificationRepo.Delete(ctx, code.Hash())

	return nil
}

func (s *AuthService) Login(ctx context.Context, email domain.Email, password domain.Password) (*domain.TokenPair, error) {
	u, err := s.userRepo.UserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if !u.IsVerified() {
		return nil, domain.ErrUserNotVerified
	}

	if u.PasswordHash() != password.Hash() {
		return nil, domain.ErrInvalidPassword
	}

	access, err := s.jwtManger.NewJWT(u.ID())
	if err != nil {
		return nil, err
	}
	refresh, err := s.jwtManger.NewRefreshToken()
	if err != nil {
		return nil, err
	}
	tokens := &domain.TokenPair{
		Access:  access,
		Refresh: refresh,
	}

	session := domain.Session{
		UserID:    u.ID(),
		TokenHash: jwt.HashRefreshToken(refresh),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(s.jwtManger.RefreshTTL()),
	}
	if err := s.sessionRepo.CreateSession(ctx, &session); err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	tokenHash := jwt.HashRefreshToken(refreshToken)

	session, err := s.sessionRepo.Session(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.UserByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}

	if !user.IsVerified() {
		return nil, domain.ErrUserNotVerified
	}

	accessToken, err := s.jwtManger.NewJWT(user.ID())
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.jwtManger.NewRefreshToken()
	if err != nil {
		return nil, err
	}

	_ = s.sessionRepo.DeleteSession(ctx, tokenHash)

	newSession := domain.Session{
		UserID:    user.ID(),
		TokenHash: jwt.HashRefreshToken(newRefreshToken),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(s.jwtManger.RefreshTTL()),
	}
	if err := s.sessionRepo.CreateSession(ctx, &newSession); err != nil {
		return nil, err
	}

	return &domain.TokenPair{
		Access:  accessToken,
		Refresh: newRefreshToken,
	}, nil
}

func (s *AuthService) ResendVerification(ctx context.Context, email domain.Email) error {
	user, err := s.userRepo.UserByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user.IsVerified() {
		return domain.ErrUserAlreadyVerified
	}
	return s.sendVerification(ctx, user)
}

func (s *AuthService) sendVerification(ctx context.Context, u *domain.User) error {
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

func (s *AuthService) User(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.UserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !user.IsVerified() {
		return nil, domain.ErrUserNotVerified
	}

	return user, nil
}
