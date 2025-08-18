package domain

type (
	User struct {
		ID          string
		Email       string
		PhoneNumber string
		FirstName   string
		LastName    *string
	}
)
