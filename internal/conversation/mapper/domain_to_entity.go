package mapper

import (
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/ericlagergren/decimal"

	"github.com/Haerd-Limited/dating-api/internal/conversation/domain"
	"github.com/Haerd-Limited/dating-api/internal/entity"
)

func MapMessageDomainToEntity(msg domain.Message) entity.Message {
	return entity.Message{
		ID:             msg.ID,
		ConversationID: msg.ConversationID,
		SenderID:       msg.SenderID,
		Type:           string(msg.Type),
		TextBody:       null.StringFromPtr(msg.TextBody),
		MediaKey:       null.StringFromPtr(msg.MediaUrl),
		MediaSeconds:   FloatPtrToNullDecimal(msg.MediaSeconds),
		ClientMSGID:    null.StringFrom(msg.ClientMsgID),
	}
}

// *float64 -> NullDecimal
func FloatPtrToNullDecimal(f *float64) types.NullDecimal {
	if f == nil {
		return types.NullDecimal{} // Big=nil => NULL
	}

	d := new(decimal.Big).SetFloat64(*f)

	return types.NullDecimal{Big: d}
}
