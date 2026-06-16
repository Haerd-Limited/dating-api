package mapper

import (
	"github.com/aarondl/null/v8"

	"github.com/Haerd-Limited/dating-api/internal/broadcast/domain"
	"github.com/Haerd-Limited/dating-api/internal/entity"
)

func DomainToEntity(log domain.BroadcastLog) *entity.SMSBroadcast {
	e := &entity.SMSBroadcast{
		UserID:  log.UserID,
		Phone:   log.Phone,
		Message: log.Message,
		Status:  log.Status,
	}

	if log.Error != nil {
		e.Error = null.StringFrom(*log.Error)
	}

	return e
}
