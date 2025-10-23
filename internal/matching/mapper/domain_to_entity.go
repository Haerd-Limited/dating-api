package mapper

import (
	"time"

	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/internal/matching/domain"
)

func MapSaveAnswerCommandToUserAnswerEntity(d domain.SaveAnswerCommand) entity.UserAnswer {
	return entity.UserAnswer{
		UserID:              d.UserID,
		QuestionID:          d.QuestionID,
		AnswerID:            d.AnswerID,
		AcceptableAnswerIds: d.AcceptableAnswerIDs,
		Importance:          d.Importance,
		IsPrivate:           d.IsPrivate,
		UpdatedAt:           time.Now().UTC(),
	}
}
