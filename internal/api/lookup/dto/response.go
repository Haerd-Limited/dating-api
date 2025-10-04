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

type GetLanguagesResponse struct {
	Languages []Language `json:"languages"`
}

type Language struct {
	ID    int16  `json:"id"`
	Label string `json:"label"`
}

type GetReligionsResponse struct {
	Religions []Religion `json:"religions"`
}

type Religion struct {
	ID    int16  `json:"id"`
	Label string `json:"label"`
}

type GetPoliticalBeliefsResponse struct {
	PoliticalBeliefs []PoliticalBelief `json:"political_beliefs"`
}

type PoliticalBelief struct {
	ID    int16  `json:"id"`
	Label string `json:"label"`
}
