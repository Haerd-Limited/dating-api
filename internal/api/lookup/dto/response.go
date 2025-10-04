package dto

type GetPromptsResponse struct {
	Prompts []Prompt `json:"prompts"`
}
type Prompt struct {
	ID       int16  `json:"id"`
	Key      string `json:"key"`
	Label    string `json:"label"`
	Category string `json:"category"`
}
