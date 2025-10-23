package dto

import "github.com/go-playground/validator/v10"

type SaveAnswerRequest struct {
	QuestionID          int64   `json:"question_id" validate:"required"`
	AnswerID            int64   `json:"answer_id" validate:"required"`
	AcceptableAnswerIDs []int64 `json:"acceptable_answer_ids"`
	Importance          string  `json:"importance" validate:"required"`
	//IsPrivate tells us whether the user wants us to allow their answers to be displayed on their profile or simply be used for algorithm
	IsPrivate bool `json:"is_private"`
}

func (sar SaveAnswerRequest) Validate() error {
	return validator.New().Struct(sar)
}
