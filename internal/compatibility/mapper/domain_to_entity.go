package mapper

import (
	"time"

	"github.com/Haerd-Limited/dating-api/internal/compatibility/domain"
	"github.com/Haerd-Limited/dating-api/internal/entity"
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
