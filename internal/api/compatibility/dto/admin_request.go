package dto

import "github.com/go-playground/validator/v10"

type AdminCreateCategoryRequest struct {
	Key  string `json:"key" validate:"required"`
	Name string `json:"name" validate:"required"`
}

func (r AdminCreateCategoryRequest) Validate() error {
	return validator.New().Struct(r)
}

type AdminUpdateCategoryRequest struct {
	Name string `json:"name" validate:"required"`
}

func (r AdminUpdateCategoryRequest) Validate() error {
	return validator.New().Struct(r)
}

type AdminCreateQuestionRequest struct {
	Text     string `json:"text" validate:"required"`
	IsActive *bool  `json:"is_active"`
}

func (r AdminCreateQuestionRequest) Validate() error {
	return validator.New().Struct(r)
}

type AdminUpdateQuestionRequest struct {
	Text     string `json:"text" validate:"required"`
	IsActive *bool  `json:"is_active"`
}

func (r AdminUpdateQuestionRequest) Validate() error {
	return validator.New().Struct(r)
}

type AdminCreateAnswerRequest struct {
	Label string `json:"label" validate:"required"`
}

func (r AdminCreateAnswerRequest) Validate() error {
	return validator.New().Struct(r)
}

type AdminUpdateAnswerRequest struct {
	Label string `json:"label" validate:"required"`
}

func (r AdminUpdateAnswerRequest) Validate() error {
	return validator.New().Struct(r)
}

type AdminReorderRequest struct {
	OrderedIDs []int64 `json:"ordered_ids" validate:"required,min=1"`
}

func (r AdminReorderRequest) Validate() error {
	return validator.New().Struct(r)
}
