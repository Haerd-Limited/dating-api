package user

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/user"
)

type Handler interface {
	MyProfile() http.HandlerFunc
}

type handler struct {
	logger      *zap.Logger
	userService user.Service
}

func NewUserHandler(
	logger *zap.Logger,
	userService user.Service,
) Handler {
	return &handler{
		logger:      logger,
		userService: userService,
	}
}

func (h *handler) MyProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// todo: implement
	}
}

/*
func mapErrorsToStatusCodeAndUserFriendlyMessages(err error) (int, string) {
	switch {
	case errors.Is(err, storage.ErrUserDoesNotExists):
		return http.StatusNotFound, "User does not exist"
	case errors.Is(err, user.ErrEmailAlreadyExists):
		return http.StatusConflict, messages.EmailAlreadyExistsMsg
	case errors.Is(err, user.ErrUserDetailsAlreadyExists):
		return http.StatusConflict, messages.UserDetailsAlreadyExistsMsg
	case errors.Is(err, commonErrors.ErrInvalidDob):
		return http.StatusBadRequest, messages.InvalidDobMsg
	case errors.Is(err, commonErrors.ErrInvalidGender):
		return http.StatusBadRequest, messages.InvalidGenderMsg
	case errors.Is(err, http.ErrNotMultipart):
		return http.StatusBadRequest, messages.InvalidUploadFormMsg
	default:
		return http.StatusInternalServerError, messages.InternalServerErrorMsg
	}
}
*/
