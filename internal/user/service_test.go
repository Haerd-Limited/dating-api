package user

import (
	"context"
	"errors"
	"testing"

	"github.com/aarondl/null/v8"
	"github.com/coocood/freecache"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	authstorage "github.com/Haerd-Limited/dating-api/internal/auth/storage"
	"github.com/Haerd-Limited/dating-api/internal/aws"
	"github.com/Haerd-Limited/dating-api/internal/entity"
	userstorage "github.com/Haerd-Limited/dating-api/internal/user/storage"
)

func TestDeleteAccountPurgesVerificationCodes(t *testing.T) {
	ctrl := gomock.NewController(t)

	userRepo := userstorage.NewMockUserRepository(ctrl)
	authRepo := authstorage.NewMockAuthRepository(ctrl)
	awsSvc := aws.NewMockService(ctrl)

	svc := NewUserService(
		zaptest.NewLogger(t),
		userRepo,
		authRepo,
		awsSvc,
		freecache.NewCache(1024),
		nil,
		nil,
		nil,
	)

	phone := "+441234567890"
	email := "user@example.com"
	userEntity := &entity.User{
		ID:    "user-1",
		Phone: null.StringFrom(phone),
		Email: null.StringFrom(email),
	}

	userRepo.EXPECT().GetUserByID(gomock.Any(), "user-1").Return(userEntity, nil)
	awsSvc.EXPECT().DeleteAllUserFiles(gomock.Any(), "user-1").Return(nil)
	authRepo.EXPECT().DeleteVerificationCodesForUser(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, p, e *string) error {
			require.NotNil(t, p)
			require.NotNil(t, e)
			require.Equal(t, phone, *p)
			require.Equal(t, email, *e)

			return nil
		})
	userRepo.EXPECT().DeleteUser(gomock.Any(), "user-1").Return(nil)

	err := svc.DeleteAccount(context.Background(), "user-1")
	require.NoError(t, err)
}

func TestDeleteAccountContinuesWhenVerificationPurgeFails(t *testing.T) {
	ctrl := gomock.NewController(t)

	userRepo := userstorage.NewMockUserRepository(ctrl)
	authRepo := authstorage.NewMockAuthRepository(ctrl)
	awsSvc := aws.NewMockService(ctrl)

	svc := NewUserService(
		zaptest.NewLogger(t),
		userRepo,
		authRepo,
		awsSvc,
		freecache.NewCache(1024),
		nil,
		nil,
		nil,
	)

	userEntity := &entity.User{
		ID:    "user-1",
		Phone: null.StringFrom("+441234567890"),
	}

	userRepo.EXPECT().GetUserByID(gomock.Any(), "user-1").Return(userEntity, nil)
	awsSvc.EXPECT().DeleteAllUserFiles(gomock.Any(), "user-1").Return(nil)
	authRepo.EXPECT().DeleteVerificationCodesForUser(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New("purge failed"))
	userRepo.EXPECT().DeleteUser(gomock.Any(), "user-1").Return(nil)

	err := svc.DeleteAccount(context.Background(), "user-1")
	require.NoError(t, err)
}
