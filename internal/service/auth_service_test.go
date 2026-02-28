package service

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

	"github.com/yohnnn/booking_service/internal/models"
	"github.com/yohnnn/booking_service/internal/service/mocks"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestAuthService_Register(t *testing.T) {
	type mockBehavior func(repo *mocks.MockUserRepository)

	tests := []struct {
		name         string
		email        string
		password     string
		mockBehavior mockBehavior
		wantErr      bool
		checkResult  func(t *testing.T, user *models.User)
	}{
		{
			name:     "success",
			email:    "test@example.com",
			password: "password123",
			mockBehavior: func(repo *mocks.MockUserRepository) {
				repo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, user *models.User) error {
						user.ID = uuid.New()
						user.CreatedAt = time.Now()
						return nil
					})
			},
			wantErr: false,
			checkResult: func(t *testing.T, user *models.User) {
				t.Helper()
				assert.Equal(t, "test@example.com", user.Email)
				assert.NotEqual(t, uuid.Nil, user.ID)
				err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("password123"))
				assert.NoError(t, err)
			},
		},
		{
			name:     "user already exists",
			email:    "existing@example.com",
			password: "password123",
			mockBehavior: func(repo *mocks.MockUserRepository) {
				repo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(models.ErrAlreadyExists)
			},
			wantErr: true,
		},
		{
			name:     "repository error",
			email:    "test@example.com",
			password: "password123",
			mockBehavior: func(repo *mocks.MockUserRepository) {
				repo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userRepo := mocks.NewMockUserRepository(ctrl)
			tt.mockBehavior(userRepo)

			s := NewAuthService(testLogger(), userRepo, "test-secret", 24*time.Hour)

			user, err := s.Register(context.Background(), tt.email, tt.password)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, user)
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	type mockBehavior func(repo *mocks.MockUserRepository)

	tests := []struct {
		name         string
		email        string
		password     string
		mockBehavior mockBehavior
		wantErr      bool
		wantToken    bool
	}{
		{
			name:     "success",
			email:    "test@example.com",
			password: "password123",
			mockBehavior: func(repo *mocks.MockUserRepository) {
				repo.EXPECT().
					GetByEmail(gomock.Any(), "test@example.com").
					Return(&models.User{
						ID:           uuid.New(),
						Email:        "test@example.com",
						PasswordHash: string(hashedPassword),
					}, nil)
			},
			wantErr:   false,
			wantToken: true,
		},
		{
			name:     "user not found",
			email:    "notfound@example.com",
			password: "password123",
			mockBehavior: func(repo *mocks.MockUserRepository) {
				repo.EXPECT().
					GetByEmail(gomock.Any(), "notfound@example.com").
					Return(nil, models.ErrNotFound)
			},
			wantErr: true,
		},
		{
			name:     "wrong password",
			email:    "test@example.com",
			password: "wrongpassword",
			mockBehavior: func(repo *mocks.MockUserRepository) {
				repo.EXPECT().
					GetByEmail(gomock.Any(), "test@example.com").
					Return(&models.User{
						ID:           uuid.New(),
						Email:        "test@example.com",
						PasswordHash: string(hashedPassword),
					}, nil)
			},
			wantErr: true,
		},
		{
			name:     "repository error",
			email:    "test@example.com",
			password: "password123",
			mockBehavior: func(repo *mocks.MockUserRepository) {
				repo.EXPECT().
					GetByEmail(gomock.Any(), "test@example.com").
					Return(nil, assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userRepo := mocks.NewMockUserRepository(ctrl)
			tt.mockBehavior(userRepo)

			s := NewAuthService(testLogger(), userRepo, "test-secret", 24*time.Hour)

			token, err := s.Login(context.Background(), tt.email, tt.password)
			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, token)
				return
			}

			require.NoError(t, err)
			if tt.wantToken {
				assert.NotEmpty(t, token)
			}
		})
	}
}

func TestAuthService_ParseToken(t *testing.T) {
	secretKey := "test-secret"
	s := NewAuthService(testLogger(), nil, secretKey, 24*time.Hour)

	userID := uuid.New()
	user := &models.User{ID: userID, Email: "test@example.com"}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.DefaultCost)
	user.PasswordHash = string(hashedPassword)

	tests := []struct {
		name    string
		token   func() string
		wantID  uuid.UUID
		wantErr bool
	}{
		{
			name: "valid token",
			token: func() string {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()
				repo := mocks.NewMockUserRepository(ctrl)
				repo.EXPECT().GetByEmail(gomock.Any(), "test@example.com").Return(user, nil)
				svc := NewAuthService(testLogger(), repo, secretKey, 24*time.Hour)
				token, _ := svc.Login(context.Background(), "test@example.com", "pass")
				return token
			},
			wantID:  userID,
			wantErr: false,
		},
		{
			name: "invalid token",
			token: func() string {
				return "invalid.token.string"
			},
			wantErr: true,
		},
		{
			name: "token with wrong secret",
			token: func() string {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()
				repo := mocks.NewMockUserRepository(ctrl)
				repo.EXPECT().GetByEmail(gomock.Any(), "test@example.com").Return(user, nil)
				svc := NewAuthService(testLogger(), repo, "wrong-secret", 24*time.Hour)
				token, _ := svc.Login(context.Background(), "test@example.com", "pass")
				return token
			},
			wantErr: true,
		},
		{
			name: "expired token",
			token: func() string {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()
				repo := mocks.NewMockUserRepository(ctrl)
				repo.EXPECT().GetByEmail(gomock.Any(), "test@example.com").Return(user, nil)
				svc := NewAuthService(testLogger(), repo, secretKey, -1*time.Hour)
				token, _ := svc.Login(context.Background(), "test@example.com", "pass")
				return token
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedID, err := s.ParseToken(tt.token())
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantID, parsedID)
		})
	}
}
