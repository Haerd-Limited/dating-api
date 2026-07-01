package compatibility

import (
	"net/http"
	"strconv"

	"github.com/friendsofgo/errors"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/compatibility/dto"
	"github.com/Haerd-Limited/dating-api/internal/api/compatibility/dto/mapper"
	"github.com/Haerd-Limited/dating-api/internal/compatibility"
	"github.com/Haerd-Limited/dating-api/internal/compatibility/domain"
	commonMappers "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/mappers"
	commonMessages "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/messages"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/request"
)

type AdminHandler interface {
	ListCategories() http.HandlerFunc
	CreateCategory() http.HandlerFunc
	UpdateCategory() http.HandlerFunc
	DeleteCategory() http.HandlerFunc
	ReorderCategories() http.HandlerFunc

	ListQuestions() http.HandlerFunc
	CreateQuestion() http.HandlerFunc
	UpdateQuestion() http.HandlerFunc
	DeleteQuestion() http.HandlerFunc
	ReorderQuestions() http.HandlerFunc

	ListAnswers() http.HandlerFunc
	CreateAnswer() http.HandlerFunc
	UpdateAnswer() http.HandlerFunc
	DeleteAnswer() http.HandlerFunc
	ReorderAnswers() http.HandlerFunc
}

type adminHandler struct {
	logger       *zap.Logger
	adminService compatibility.AdminService
}

func NewAdminCompatibilityHandler(
	logger *zap.Logger,
	adminService compatibility.AdminService,
) AdminHandler {
	return &adminHandler{
		logger:       logger,
		adminService: adminService,
	}
}

// ---- categories ----

func (h *adminHandler) ListCategories() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result, err := h.adminService.ListCategories(r.Context())
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "ListCategories", err, mapAdminErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, mapper.MapCategoryRowsToResponse(result))
	}
}

func (h *adminHandler) CreateCategory() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.AdminCreateCategoryRequest
		if err := request.DecodeAndValidate(r.Body, &req); err != nil {
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("invalid request payload"))
			return
		}

		created, err := h.adminService.CreateCategory(r.Context(), domain.CategoryInput{Key: req.Key, Name: req.Name})
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "CreateCategory", err, mapAdminErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusCreated, mapper.MapCategoryRowToResponse(created))
	}
}

func (h *adminHandler) UpdateCategory() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseIDParam(w, r, "categoryID")
		if !ok {
			return
		}

		var req dto.AdminUpdateCategoryRequest
		if err := request.DecodeAndValidate(r.Body, &req); err != nil {
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("invalid request payload"))
			return
		}

		if err := h.adminService.UpdateCategory(r.Context(), id, req.Name); err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "UpdateCategory", err, mapAdminErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, commonMappers.ToSimpleMessageResponse("Category updated"))
	}
}

func (h *adminHandler) DeleteCategory() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseIDParam(w, r, "categoryID")
		if !ok {
			return
		}

		if err := h.adminService.DeleteCategory(r.Context(), id); err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "DeleteCategory", err, mapAdminErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, commonMappers.ToSimpleMessageResponse("Category deleted"))
	}
}

func (h *adminHandler) ReorderCategories() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.AdminReorderRequest
		if err := request.DecodeAndValidate(r.Body, &req); err != nil {
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("invalid request payload"))
			return
		}

		if err := h.adminService.ReorderCategories(r.Context(), domain.ReorderCommand{OrderedIDs: req.OrderedIDs}); err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "ReorderCategories", err, mapAdminErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, commonMappers.ToSimpleMessageResponse("Categories reordered"))
	}
}

// ---- questions ----

func (h *adminHandler) ListQuestions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		categoryID, ok := parseIDParam(w, r, "categoryID")
		if !ok {
			return
		}

		result, err := h.adminService.ListQuestions(r.Context(), categoryID)
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "ListQuestions", err, mapAdminErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, mapper.MapQuestionRowsToResponse(result))
	}
}

func (h *adminHandler) CreateQuestion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		categoryID, ok := parseIDParam(w, r, "categoryID")
		if !ok {
			return
		}

		var req dto.AdminCreateQuestionRequest
		if err := request.DecodeAndValidate(r.Body, &req); err != nil {
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("invalid request payload"))
			return
		}

		isActive := true
		if req.IsActive != nil {
			isActive = *req.IsActive
		}

		created, err := h.adminService.CreateQuestion(r.Context(), domain.QuestionInput{
			CategoryID: categoryID,
			Text:       req.Text,
			IsActive:   isActive,
		})
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "CreateQuestion", err, mapAdminErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusCreated, mapper.MapQuestionRowToResponse(created))
	}
}

func (h *adminHandler) UpdateQuestion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseIDParam(w, r, "questionID")
		if !ok {
			return
		}

		var req dto.AdminUpdateQuestionRequest
		if err := request.DecodeAndValidate(r.Body, &req); err != nil {
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("invalid request payload"))
			return
		}

		isActive := true
		if req.IsActive != nil {
			isActive = *req.IsActive
		}

		if err := h.adminService.UpdateQuestion(r.Context(), id, req.Text, isActive); err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "UpdateQuestion", err, mapAdminErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, commonMappers.ToSimpleMessageResponse("Question updated"))
	}
}

func (h *adminHandler) DeleteQuestion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseIDParam(w, r, "questionID")
		if !ok {
			return
		}

		if err := h.adminService.DeleteQuestion(r.Context(), id); err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "DeleteQuestion", err, mapAdminErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, commonMappers.ToSimpleMessageResponse("Question deleted"))
	}
}

func (h *adminHandler) ReorderQuestions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		categoryID, ok := parseIDParam(w, r, "categoryID")
		if !ok {
			return
		}

		var req dto.AdminReorderRequest
		if err := request.DecodeAndValidate(r.Body, &req); err != nil {
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("invalid request payload"))
			return
		}

		if err := h.adminService.ReorderQuestions(r.Context(), categoryID, domain.ReorderCommand{OrderedIDs: req.OrderedIDs}); err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "ReorderQuestions", err, mapAdminErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, commonMappers.ToSimpleMessageResponse("Questions reordered"))
	}
}

// ---- answers ----

func (h *adminHandler) ListAnswers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		questionID, ok := parseIDParam(w, r, "questionID")
		if !ok {
			return
		}

		result, err := h.adminService.ListAnswers(r.Context(), questionID)
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "ListAnswers", err, mapAdminErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, mapper.MapAnswerRowsToResponse(result))
	}
}

func (h *adminHandler) CreateAnswer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		questionID, ok := parseIDParam(w, r, "questionID")
		if !ok {
			return
		}

		var req dto.AdminCreateAnswerRequest
		if err := request.DecodeAndValidate(r.Body, &req); err != nil {
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("invalid request payload"))
			return
		}

		created, err := h.adminService.CreateAnswer(r.Context(), domain.AnswerInput{
			QuestionID: questionID,
			Label:      req.Label,
		})
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "CreateAnswer", err, mapAdminErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusCreated, mapper.MapAnswerRowToResponse(created))
	}
}

func (h *adminHandler) UpdateAnswer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseIDParam(w, r, "answerID")
		if !ok {
			return
		}

		var req dto.AdminUpdateAnswerRequest
		if err := request.DecodeAndValidate(r.Body, &req); err != nil {
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("invalid request payload"))
			return
		}

		if err := h.adminService.UpdateAnswer(r.Context(), id, req.Label); err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "UpdateAnswer", err, mapAdminErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, commonMappers.ToSimpleMessageResponse("Answer updated"))
	}
}

func (h *adminHandler) DeleteAnswer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseIDParam(w, r, "answerID")
		if !ok {
			return
		}

		if err := h.adminService.DeleteAnswer(r.Context(), id); err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "DeleteAnswer", err, mapAdminErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, commonMappers.ToSimpleMessageResponse("Answer deleted"))
	}
}

func (h *adminHandler) ReorderAnswers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		questionID, ok := parseIDParam(w, r, "questionID")
		if !ok {
			return
		}

		var req dto.AdminReorderRequest
		if err := request.DecodeAndValidate(r.Body, &req); err != nil {
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("invalid request payload"))
			return
		}

		if err := h.adminService.ReorderAnswers(r.Context(), questionID, domain.ReorderCommand{OrderedIDs: req.OrderedIDs}); err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "ReorderAnswers", err, mapAdminErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, commonMappers.ToSimpleMessageResponse("Answers reordered"))
	}
}

func parseIDParam(w http.ResponseWriter, r *http.Request, name string) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, name), 10, 64)
	if err != nil || id < 1 {
		render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("invalid "+name))
		return 0, false
	}

	return id, true
}

func mapAdminErrorsToStatusCodeAndUserFriendlyMessages(err error) (int, string) {
	switch {
	case errors.Is(err, compatibility.ErrValidation), errors.Is(err, compatibility.ErrInvalidReorder):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, compatibility.ErrCategoryNotFound):
		return http.StatusNotFound, "Category not found"
	case errors.Is(err, compatibility.ErrQuestionNotFound):
		return http.StatusNotFound, "Question not found"
	case errors.Is(err, compatibility.ErrAnswerNotFound):
		return http.StatusNotFound, "Answer not found"
	case errors.Is(err, compatibility.ErrDuplicateCategoryKey):
		return http.StatusConflict, "A category with that key already exists"
	case errors.Is(err, compatibility.ErrDuplicateQuestionText):
		return http.StatusConflict, "A question with that text already exists in this pack"
	case errors.Is(err, compatibility.ErrDuplicateAnswerLabel):
		return http.StatusConflict, "An answer with that label already exists for this question"
	case errors.Is(err, compatibility.ErrCategoryNotEmpty):
		return http.StatusConflict, "Remove or move the questions in this pack before deleting it"
	case errors.Is(err, compatibility.ErrQuestionHasUserAnswers):
		return http.StatusConflict, "This question has user answers; retire it instead of deleting"
	case errors.Is(err, compatibility.ErrAnswerInUse):
		return http.StatusConflict, "This answer option is in use; edit its label instead of deleting"
	default:
		return http.StatusInternalServerError, commonMessages.InternalServerErrorMsg
	}
}
