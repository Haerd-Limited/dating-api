package compatibility

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	"github.com/Haerd-Limited/dating-api/internal/compatibility/domain"
	"github.com/Haerd-Limited/dating-api/internal/compatibility/storage"
	"github.com/Haerd-Limited/dating-api/internal/entity"
)

func newAdminTestService(t *testing.T) (*adminService, *storage.MockAdminCompatibilityRepository) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := storage.NewMockAdminCompatibilityRepository(ctrl)
	svc := &adminService{logger: zaptest.NewLogger(t), adminRepo: repo}

	return svc, repo
}

func TestAdminDeleteQuestion(t *testing.T) {
	ctx := context.Background()

	t.Run("blocks when question has user answers", func(t *testing.T) {
		svc, repo := newAdminTestService(t)
		repo.EXPECT().CountUserAnswersForQuestion(ctx, int64(7)).Return(3, nil)

		err := svc.DeleteQuestion(ctx, 7)
		if !errors.Is(err, ErrQuestionHasUserAnswers) {
			t.Fatalf("expected ErrQuestionHasUserAnswers, got %v", err)
		}
	})

	t.Run("deletes when zero user answers", func(t *testing.T) {
		svc, repo := newAdminTestService(t)
		repo.EXPECT().CountUserAnswersForQuestion(ctx, int64(7)).Return(0, nil)
		repo.EXPECT().DeleteQuestion(ctx, int64(7)).Return(nil)

		if err := svc.DeleteQuestion(ctx, 7); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})
}

func TestAdminDeleteAnswer(t *testing.T) {
	ctx := context.Background()

	t.Run("blocks when answer in use", func(t *testing.T) {
		svc, repo := newAdminTestService(t)
		repo.EXPECT().CountUserAnswersForAnswer(ctx, int64(9)).Return(1, nil)

		err := svc.DeleteAnswer(ctx, 9)
		if !errors.Is(err, ErrAnswerInUse) {
			t.Fatalf("expected ErrAnswerInUse, got %v", err)
		}
	})

	t.Run("maps FK violation to in-use", func(t *testing.T) {
		svc, repo := newAdminTestService(t)
		repo.EXPECT().CountUserAnswersForAnswer(ctx, int64(9)).Return(0, nil)
		repo.EXPECT().DeleteAnswer(ctx, int64(9)).Return(storage.ErrForeignKeyViolation)

		err := svc.DeleteAnswer(ctx, 9)
		if !errors.Is(err, ErrAnswerInUse) {
			t.Fatalf("expected ErrAnswerInUse, got %v", err)
		}
	})
}

func TestAdminDeleteCategory(t *testing.T) {
	ctx := context.Background()

	t.Run("blocks when category has questions", func(t *testing.T) {
		svc, repo := newAdminTestService(t)
		repo.EXPECT().CountQuestionsInCategory(ctx, int64(2)).Return(5, nil)

		err := svc.DeleteCategory(ctx, 2)
		if !errors.Is(err, ErrCategoryNotEmpty) {
			t.Fatalf("expected ErrCategoryNotEmpty, got %v", err)
		}
	})

	t.Run("deletes empty category", func(t *testing.T) {
		svc, repo := newAdminTestService(t)
		repo.EXPECT().CountQuestionsInCategory(ctx, int64(2)).Return(0, nil)
		repo.EXPECT().DeleteCategory(ctx, int64(2)).Return(nil)

		if err := svc.DeleteCategory(ctx, 2); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})
}

func TestAdminCreateCategoryValidation(t *testing.T) {
	ctx := context.Background()

	t.Run("rejects empty name", func(t *testing.T) {
		svc, _ := newAdminTestService(t)

		_, err := svc.CreateCategory(ctx, domain.CategoryInput{Key: "faith_values", Name: "  "})
		if !errors.Is(err, ErrValidation) {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("rejects invalid key pattern", func(t *testing.T) {
		svc, _ := newAdminTestService(t)

		_, err := svc.CreateCategory(ctx, domain.CategoryInput{Key: "Faith Values", Name: "Faith"})
		if !errors.Is(err, ErrValidation) {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("maps unique violation to duplicate key", func(t *testing.T) {
		svc, repo := newAdminTestService(t)
		repo.EXPECT().NextCategorySortOrder(ctx).Return(3, nil)
		repo.EXPECT().CreateCategory(ctx, "faith_values", "Faith", 4).Return(nil, storage.ErrUniqueViolation)

		_, err := svc.CreateCategory(ctx, domain.CategoryInput{Key: "faith_values", Name: "Faith"})
		if !errors.Is(err, ErrDuplicateCategoryKey) {
			t.Fatalf("expected ErrDuplicateCategoryKey, got %v", err)
		}
	})

	t.Run("assigns next sort order on success", func(t *testing.T) {
		svc, repo := newAdminTestService(t)
		repo.EXPECT().NextCategorySortOrder(ctx).Return(3, nil)
		repo.EXPECT().CreateCategory(ctx, "faith_values", "Faith", 4).
			Return(&entity.QuestionCategory{ID: 10, Key: "faith_values", Name: "Faith", SortOrder: 4}, nil)

		row, err := svc.CreateCategory(ctx, domain.CategoryInput{Key: "faith_values", Name: "Faith"})
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if row.ID != 10 || row.SortOrder != 4 {
			t.Fatalf("unexpected row: %+v", row)
		}
	})
}

func TestAdminCreateQuestionDuplicate(t *testing.T) {
	ctx := context.Background()
	svc, repo := newAdminTestService(t)

	repo.EXPECT().GetCategory(ctx, int64(1)).Return(&entity.QuestionCategory{ID: 1}, nil)
	repo.EXPECT().NextQuestionSortOrder(ctx, int64(1)).Return(2, nil)
	repo.EXPECT().CreateQuestion(ctx, int64(1), "Do you want children?", true, 3).Return(nil, storage.ErrUniqueViolation)

	_, err := svc.CreateQuestion(ctx, domain.QuestionInput{CategoryID: 1, Text: "Do you want children?", IsActive: true})
	if !errors.Is(err, ErrDuplicateQuestionText) {
		t.Fatalf("expected ErrDuplicateQuestionText, got %v", err)
	}
}

func TestAdminReorderValidation(t *testing.T) {
	ctx := context.Background()

	t.Run("rejects empty set", func(t *testing.T) {
		svc, _ := newAdminTestService(t)

		err := svc.ReorderCategories(ctx, domain.ReorderCommand{OrderedIDs: nil})
		if !errors.Is(err, ErrInvalidReorder) {
			t.Fatalf("expected ErrInvalidReorder, got %v", err)
		}
	})

	t.Run("rejects duplicate ids", func(t *testing.T) {
		svc, _ := newAdminTestService(t)

		err := svc.ReorderCategories(ctx, domain.ReorderCommand{OrderedIDs: []int64{1, 2, 2}})
		if !errors.Is(err, ErrInvalidReorder) {
			t.Fatalf("expected ErrInvalidReorder, got %v", err)
		}
	})

	t.Run("maps repo not-found to invalid reorder", func(t *testing.T) {
		svc, repo := newAdminTestService(t)
		repo.EXPECT().ReorderCategories(ctx, []int64{3, 1, 2}).Return(storage.ErrNotFound)

		err := svc.ReorderCategories(ctx, domain.ReorderCommand{OrderedIDs: []int64{3, 1, 2}})
		if !errors.Is(err, ErrInvalidReorder) {
			t.Fatalf("expected ErrInvalidReorder, got %v", err)
		}
	})
}
