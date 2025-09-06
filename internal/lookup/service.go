package lookup

type Service interface{}

type service struct{}

func NewLookupService() Service {
	return &service{}
}
