package domain

type RegisterDeviceTokenInput struct {
	UserID string
	Token  string
}

type SendNotificationInput struct {
	UserID string
	Title  string
	Body   string
	Data   map[string]string // optional key-value data payload
}
