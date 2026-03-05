package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/api/compatibility/dto"
	"github.com/Haerd-Limited/dating-api/internal/compatibility/domain"
)

func MapSaveAnswerRequestToDomain(req dto.SaveAnswerRequest, userID string) domain.SaveAnswerCommand {
	return domain.SaveAnswerCommand{
		UserID:              userID,
		QuestionID:          req.QuestionID,
		AnswerID:            req.AnswerID,
		AcceptableAnswerIDs: req.AcceptableAnswerIDs,
		Importance:          req.Importance,
		IsPrivate:           req.IsPrivate,
	}
}
