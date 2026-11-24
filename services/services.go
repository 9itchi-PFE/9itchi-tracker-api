package services

import (
	"9itchi-tracker-api/core"
	"9itchi-tracker-api/pkg/email"
	"9itchi-tracker-api/repositories"
	"9itchi-tracker-api/utils"
)

// Services holds service implementations and the peer clients they call.
type Services struct {
	Tracking TrackingService
	Clients  *Clients
}

// Clients holds one HTTP client per peer service (singletons, 50s timeout).
type Clients struct {
	Email *email.Client
}

func InitServices(config utils.AppConfig, repos *repositories.Repositories) *Services {
	return &Services{
		Tracking: NewTrackingService(repos.Tracking, core.NewPublisher(config.RabbitMQURL), config.SecretKey),
		Clients: &Clients{
			Email: email.New(config.EmailAPIURL, config.InternalAPIKey),
		},
	}
}
