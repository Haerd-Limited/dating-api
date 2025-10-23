package dto

import "github.com/go-playground/validator/v10"

type SaveAnswerRequest struct {
	QuestionID          int64   `json:"question_id" validate:"required"`
	AnswerID            int64   `json:"answer_id" validate:"required"`
	AcceptableAnswerIDs []int64 `json:"acceptable_answer_ids"`
	Importance          string  `json:"importance" validate:"required"`
	IsPrivate           bool    `json:"is_private"`
}

func (sar SaveAnswerRequest) Validate() error {
	return validator.New().Struct(sar)
}
