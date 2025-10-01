package lookup

// todo: look up
type Service interface{}

type service struct{}

func NewLookupService() Service {
	return &service{}
}
