package domain

type Prompt struct {
	ID       int16
	Key      string
	Label    string
	Category string
}

type Language struct {
	ID    int16
	Label string
}

type Religion struct {
	ID    int16
	Label string
}
