package compatibility

//go:generate mockgen -source=admin_service.go -destination=admin_service_mock.go -package=compatibility

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/compatibility/domain"
	"github.com/Haerd-Limited/dating-api/internal/compatibility/storage"
)

var (
	ErrCategoryNotFound = errors.New("category not found")
	ErrAnswerNotFound   = errors.New("answer not found")
	// ErrQuestionNotFound is declared in service.go and reused here.

	ErrDuplicateCategoryKey  = errors.New("category key already exists")
	ErrDuplicateQuestionText = errors.New("question text already exists in this category")
	ErrDuplicateAnswerLabel  = errors.New("answer label already exists for this question")

	ErrCategoryNotEmpty       = errors.New("category still has questions")
	ErrQuestionHasUserAnswers = errors.New("question has user answers")
	ErrAnswerInUse            = errors.New("answer option is in use")

	ErrValidation     = errors.New("validation failed")
	ErrInvalidReorder = errors.New("invalid reorder set")
)

var categoryKeyPattern = regexp.MustCompile(`^[a-z0-9_]+$`)

// AdminService manages compatibility question-pack content for the admin dashboard.
type AdminService interface {
	// categories
	ListCategories(ctx context.Context) ([]domain.CategoryListRow, error)
	CreateCategory(ctx context.Context, in domain.CategoryInput) (domain.CategoryListRow, error)
	UpdateCategory(ctx context.Context, id int64, name string) error
	DeleteCategory(ctx context.Context, id int64) error
	ReorderCategories(ctx context.Context, cmd domain.ReorderCommand) error

	// questions
	ListQuestions(ctx context.Context, categoryID int64) ([]domain.QuestionAdminRow, error)
	CreateQuestion(ctx context.Context, in domain.QuestionInput) (domain.QuestionAdminRow, error)
	UpdateQuestion(ctx context.Context, id int64, text string, isActive bool) error
	DeleteQuestion(ctx context.Context, id int64) error
	ReorderQuestions(ctx context.Context, categoryID int64, cmd domain.ReorderCommand) error

	// answers
	ListAnswers(ctx context.Context, questionID int64) ([]domain.AnswerAdminRow, error)
	CreateAnswer(ctx context.Context, in domain.AnswerInput) (domain.AnswerAdminRow, error)
	UpdateAnswer(ctx context.Context, id int64, label string) error
	DeleteAnswer(ctx context.Context, id int64) error
	ReorderAnswers(ctx context.Context, questionID int64, cmd domain.ReorderCommand) error
}

type adminService struct {
	logger    *zap.Logger
	adminRepo storage.AdminCompatibilityRepository
}

func NewAdminService(
	logger *zap.Logger,
	adminRepo storage.AdminCompatibilityRepository,
) AdminService {
	return &adminService{
		logger:    logger,
		adminRepo: adminRepo,
	}
}

// ---- categories ----

func (s *adminService) ListCategories(ctx context.Context) ([]domain.CategoryListRow, error) {
	rows, err := s.adminRepo.ListCategoriesWithCounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}

	return rows, nil
}

func (s *adminService) CreateCategory(ctx context.Context, in domain.CategoryInput) (domain.CategoryListRow, error) {
	key := strings.TrimSpace(in.Key)
	name := strings.TrimSpace(in.Name)

	if key == "" || name == "" {
		return domain.CategoryListRow{}, fmt.Errorf("%w: key and name are required", ErrValidation)
	}

	if !categoryKeyPattern.MatchString(key) {
		return domain.CategoryListRow{}, fmt.Errorf("%w: key must match ^[a-z0-9_]+$", ErrValidation)
	}

	next, err := s.adminRepo.NextCategorySortOrder(ctx)
	if err != nil {
		return domain.CategoryListRow{}, fmt.Errorf("next category sort order: %w", err)
	}

	created, err := s.adminRepo.CreateCategory(ctx, key, name, next+1)
	if err != nil {
		if errors.Is(err, storage.ErrUniqueViolation) {
			return domain.CategoryListRow{}, ErrDuplicateCategoryKey
		}

		return domain.CategoryListRow{}, fmt.Errorf("create category: %w", err)
	}

	return domain.CategoryListRow{
		ID:        created.ID,
		Key:       created.Key,
		Name:      created.Name,
		SortOrder: created.SortOrder,
	}, nil
}

func (s *adminService) UpdateCategory(ctx context.Context, id int64, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("%w: name is required", ErrValidation)
	}

	if err := s.adminRepo.UpdateCategoryName(ctx, id, name); err != nil {
		switch {
		case errors.Is(err, storage.ErrNotFound):
			return ErrCategoryNotFound
		case errors.Is(err, storage.ErrUniqueViolation):
			return ErrDuplicateCategoryKey
		default:
			return fmt.Errorf("update category: %w", err)
		}
	}

	return nil
}

func (s *adminService) DeleteCategory(ctx context.Context, id int64) error {
	count, err := s.adminRepo.CountQuestionsInCategory(ctx, id)
	if err != nil {
		return fmt.Errorf("count questions in category: %w", err)
	}

	if count > 0 {
		return ErrCategoryNotEmpty
	}

	if err := s.adminRepo.DeleteCategory(ctx, id); err != nil {
		switch {
		case errors.Is(err, storage.ErrNotFound):
			return ErrCategoryNotFound
		case errors.Is(err, storage.ErrForeignKeyViolation):
			return ErrCategoryNotEmpty
		default:
			return fmt.Errorf("delete category: %w", err)
		}
	}

	return nil
}

func (s *adminService) ReorderCategories(ctx context.Context, cmd domain.ReorderCommand) error {
	if err := validateReorder(cmd); err != nil {
		return err
	}

	if err := s.adminRepo.ReorderCategories(ctx, cmd.OrderedIDs); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return ErrInvalidReorder
		}

		return fmt.Errorf("reorder categories: %w", err)
	}

	return nil
}

// ---- questions ----

func (s *adminService) ListQuestions(ctx context.Context, categoryID int64) ([]domain.QuestionAdminRow, error) {
	if _, err := s.adminRepo.GetCategory(ctx, categoryID); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrCategoryNotFound
		}

		return nil, fmt.Errorf("get category: %w", err)
	}

	rows, err := s.adminRepo.ListQuestionsByCategoryAdmin(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("list questions: %w", err)
	}

	return rows, nil
}

func (s *adminService) CreateQuestion(ctx context.Context, in domain.QuestionInput) (domain.QuestionAdminRow, error) {
	text := strings.TrimSpace(in.Text)
	if text == "" {
		return domain.QuestionAdminRow{}, fmt.Errorf("%w: text is required", ErrValidation)
	}

	if _, err := s.adminRepo.GetCategory(ctx, in.CategoryID); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return domain.QuestionAdminRow{}, ErrCategoryNotFound
		}

		return domain.QuestionAdminRow{}, fmt.Errorf("get category: %w", err)
	}

	next, err := s.adminRepo.NextQuestionSortOrder(ctx, in.CategoryID)
	if err != nil {
		return domain.QuestionAdminRow{}, fmt.Errorf("next question sort order: %w", err)
	}

	created, err := s.adminRepo.CreateQuestion(ctx, in.CategoryID, text, in.IsActive, next+1)
	if err != nil {
		if errors.Is(err, storage.ErrUniqueViolation) {
			return domain.QuestionAdminRow{}, ErrDuplicateQuestionText
		}

		return domain.QuestionAdminRow{}, fmt.Errorf("create question: %w", err)
	}

	return domain.QuestionAdminRow{
		ID:         created.ID,
		CategoryID: created.CategoryID,
		Text:       created.Text,
		IsActive:   created.IsActive,
		SortOrder:  created.SortOrder,
	}, nil
}

func (s *adminService) UpdateQuestion(ctx context.Context, id int64, text string, isActive bool) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return fmt.Errorf("%w: text is required", ErrValidation)
	}

	if err := s.adminRepo.UpdateQuestion(ctx, id, text, isActive); err != nil {
		switch {
		case errors.Is(err, storage.ErrNotFound):
			return ErrQuestionNotFound
		case errors.Is(err, storage.ErrUniqueViolation):
			return ErrDuplicateQuestionText
		default:
			return fmt.Errorf("update question: %w", err)
		}
	}

	return nil
}

func (s *adminService) DeleteQuestion(ctx context.Context, id int64) error {
	count, err := s.adminRepo.CountUserAnswersForQuestion(ctx, id)
	if err != nil {
		return fmt.Errorf("count user answers for question: %w", err)
	}

	if count > 0 {
		return ErrQuestionHasUserAnswers
	}

	if err := s.adminRepo.DeleteQuestion(ctx, id); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return ErrQuestionNotFound
		}

		return fmt.Errorf("delete question: %w", err)
	}

	return nil
}

func (s *adminService) ReorderQuestions(ctx context.Context, categoryID int64, cmd domain.ReorderCommand) error {
	if err := validateReorder(cmd); err != nil {
		return err
	}

	if err := s.adminRepo.ReorderQuestions(ctx, categoryID, cmd.OrderedIDs); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return ErrInvalidReorder
		}

		return fmt.Errorf("reorder questions: %w", err)
	}

	return nil
}

// ---- answers ----

func (s *adminService) ListAnswers(ctx context.Context, questionID int64) ([]domain.AnswerAdminRow, error) {
	if _, err := s.adminRepo.GetQuestionAdmin(ctx, questionID); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrQuestionNotFound
		}

		return nil, fmt.Errorf("get question: %w", err)
	}

	rows, err := s.adminRepo.ListAnswersByQuestionAdmin(ctx, questionID)
	if err != nil {
		return nil, fmt.Errorf("list answers: %w", err)
	}

	return rows, nil
}

func (s *adminService) CreateAnswer(ctx context.Context, in domain.AnswerInput) (domain.AnswerAdminRow, error) {
	label := strings.TrimSpace(in.Label)
	if label == "" {
		return domain.AnswerAdminRow{}, fmt.Errorf("%w: label is required", ErrValidation)
	}

	if _, err := s.adminRepo.GetQuestionAdmin(ctx, in.QuestionID); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return domain.AnswerAdminRow{}, ErrQuestionNotFound
		}

		return domain.AnswerAdminRow{}, fmt.Errorf("get question: %w", err)
	}

	next, err := s.adminRepo.NextAnswerSort(ctx, in.QuestionID)
	if err != nil {
		return domain.AnswerAdminRow{}, fmt.Errorf("next answer sort: %w", err)
	}

	created, err := s.adminRepo.CreateAnswer(ctx, in.QuestionID, label, next+1)
	if err != nil {
		if errors.Is(err, storage.ErrUniqueViolation) {
			return domain.AnswerAdminRow{}, ErrDuplicateAnswerLabel
		}

		return domain.AnswerAdminRow{}, fmt.Errorf("create answer: %w", err)
	}

	return domain.AnswerAdminRow{
		ID:         created.ID,
		QuestionID: created.QuestionID,
		Label:      created.Label,
		Sort:       created.Sort,
	}, nil
}

func (s *adminService) UpdateAnswer(ctx context.Context, id int64, label string) error {
	label = strings.TrimSpace(label)
	if label == "" {
		return fmt.Errorf("%w: label is required", ErrValidation)
	}

	if err := s.adminRepo.UpdateAnswerLabel(ctx, id, label); err != nil {
		switch {
		case errors.Is(err, storage.ErrNotFound):
			return ErrAnswerNotFound
		case errors.Is(err, storage.ErrUniqueViolation):
			return ErrDuplicateAnswerLabel
		default:
			return fmt.Errorf("update answer: %w", err)
		}
	}

	return nil
}

func (s *adminService) DeleteAnswer(ctx context.Context, id int64) error {
	count, err := s.adminRepo.CountUserAnswersForAnswer(ctx, id)
	if err != nil {
		return fmt.Errorf("count user answers for answer: %w", err)
	}

	if count > 0 {
		return ErrAnswerInUse
	}

	if err := s.adminRepo.DeleteAnswer(ctx, id); err != nil {
		switch {
		case errors.Is(err, storage.ErrNotFound):
			return ErrAnswerNotFound
		case errors.Is(err, storage.ErrForeignKeyViolation):
			return ErrAnswerInUse
		default:
			return fmt.Errorf("delete answer: %w", err)
		}
	}

	return nil
}

func (s *adminService) ReorderAnswers(ctx context.Context, questionID int64, cmd domain.ReorderCommand) error {
	if err := validateReorder(cmd); err != nil {
		return err
	}

	if err := s.adminRepo.ReorderAnswers(ctx, questionID, cmd.OrderedIDs); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return ErrInvalidReorder
		}

		return fmt.Errorf("reorder answers: %w", err)
	}

	return nil
}

// validateReorder ensures the ID set is non-empty and free of duplicates.
func validateReorder(cmd domain.ReorderCommand) error {
	if len(cmd.OrderedIDs) == 0 {
		return fmt.Errorf("%w: ordered_ids is required", ErrInvalidReorder)
	}

	seen := make(map[int64]struct{}, len(cmd.OrderedIDs))
	for _, id := range cmd.OrderedIDs {
		if _, dup := seen[id]; dup {
			return fmt.Errorf("%w: duplicate id %d", ErrInvalidReorder, id)
		}

		seen[id] = struct{}{}
	}

	return nil
}
