package communication

import (
	"github.com/twilio/twilio-go"
)

type Service interface {
	SendEmailOTP(toEmail, code string) error
	SendSMSOTP(toNumber, code string) error
}

type service struct {
	client     *twilio.RestClient
	fromNumber string
}

func NewService(
	TwilioAccountSID string,
	TwilioAuthToken string,
	TwilioFromNumber string,
) Service {
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: TwilioAccountSID,
		Password: TwilioAuthToken,
	})
	return &service{
		client:     client,
		fromNumber: TwilioFromNumber,
	}
}

func (s *service) SendEmailOTP(to, code string) error {
	// TODO: integrate SES/SendGrid/etc
	return nil
}
func (s *service) SendSMSOTP(toNumber, code string) error {
	/*
		params := &twilioApi.CreateMessageParams{}
		params.SetTo(toNumber)
		params.SetFrom(s.fromNumber)
		params.SetBody(fmt.Sprintf("Your Haerd code is %s", code))

		_, err := s.client.Api.CreateMessage(params)
		if err != nil {
			return err
		}
	*/

	return nil
}
