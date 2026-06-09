package interaction

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	"github.com/Haerd-Limited/dating-api/internal/interaction/storage"
)

func TestAddFavouriteSelfWatch(t *testing.T) {
	svc := newServiceWithRepo(t, storage.NewMockInteractionRepository(gomock.NewController(t)))

	err := svc.AddFavourite(context.Background(), "user-1", "user-1")

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSelfWatch)
}

func TestAddFavouriteNotIncomingLiker(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := storage.NewMockInteractionRepository(ctrl)

	repo.EXPECT().HasIncomingLike(gomock.Any(), "watcher-1", "other-1").Return(false, nil)

	svc := newServiceWithRepo(t, repo)

	err := svc.AddFavourite(context.Background(), "watcher-1", "other-1")

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNotAnIncomingLiker)
}

func TestAddFavouriteSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := storage.NewMockInteractionRepository(ctrl)

	repo.EXPECT().HasIncomingLike(gomock.Any(), "watcher-1", "liker-1").Return(true, nil)
	repo.EXPECT().InsertWatch(gomock.Any(), "watcher-1", "liker-1").Return(nil)

	svc := newServiceWithRepo(t, repo)

	err := svc.AddFavourite(context.Background(), "watcher-1", "liker-1")

	require.NoError(t, err)
}

func TestRemoveFavouriteSelfWatch(t *testing.T) {
	svc := &service{logger: zaptest.NewLogger(t)}

	err := svc.RemoveFavourite(context.Background(), "user-1", "user-1")

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSelfWatch)
}
