package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/auth"
	"github.com/Haerd-Limited/dating-api/internal/api/onboarding"
	"github.com/Haerd-Limited/dating-api/internal/api/user"
	internalauth "github.com/Haerd-Limited/dating-api/internal/auth"
	haerdmiddleware "github.com/Haerd-Limited/dating-api/internal/middleware"
	internalonboarding "github.com/Haerd-Limited/dating-api/internal/onboarding"
	internaluser "github.com/Haerd-Limited/dating-api/internal/user"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
)

func New(
	logger *zap.Logger,
	jwtSecret string,
	authService internalauth.Service,
	userService internaluser.Service,
	onboardingService internalonboarding.Service,
) http.Handler {
	// Create a new Chi router.
	router := chi.NewRouter()

	// Add middleware.
	router.Use(middleware.Logger)    // logs every request
	router.Use(middleware.Recoverer) // recovers from panics

	authHandler := auth.NewAuthHandler(logger, authService)
	userHandler := user.NewUserHandler(logger, userService)
	onboardingHandler := onboarding.NewOnboardingHandler(logger, onboardingService)
	// notificationsHandler := notification.NewNotificationHandler(logger, notificationService)

	// Define the /alive endpoint.
	registerAliveEndpoint(router)
	router.Route(
		"/api/v1", func(r chi.Router) {
			r.Route(
				"/auth", func(r chi.Router) {
					// todo: make login return the user's step. frontend will check if step != COMPLETE. if so call GetStep
					r.Post("/login", authHandler.Login())
					r.Post("/refresh", authHandler.Refresh())
					r.Post("/logout", authHandler.Logout())
				},
			)

			// --- Protected (must be logged in)
			r.Group(func(r chi.Router) {
				r.Route(
					"/onboarding", func(r chi.Router) {
						r.Post("/register", onboardingHandler.Register())
						// After the register endpoint, the user has an account and therefore must be authroised to complete the other steps
						r.Group(func(r chi.Router) {
							r.Use(haerdmiddleware.AuthMiddleware([]byte(jwtSecret)))
							r.Get("/step", onboardingHandler.GetStep())
							r.Patch("/basics", onboardingHandler.Basics())
							r.Patch("/location", onboardingHandler.Location())
							r.Patch("/lifestyle", onboardingHandler.Lifestyle())
							r.Patch("/beliefs", onboardingHandler.Beliefs())
							r.Patch("/background", onboardingHandler.Background())
							r.Patch("/word-and-education", onboardingHandler.WorkAndEducation())
							r.Post("/languages", onboardingHandler.Languages())
							r.Post("/photos", onboardingHandler.Photos())
							r.Post("/prompts", onboardingHandler.Prompts())
						})
					},
				)

				// Media (used during onboarding & later)
				/*
					r.Post("/media/photos/presign", mediaHandler.PresignPhoto()) // returns URL/fields
					r.Post("/media/photos", mediaHandler.AttachPhoto())          // save URL, position, is_primary
					r.Patch("/media/photos/{id}", mediaHandler.UpdatePhoto())    // reorder / set primary
					r.Delete("/media/photos/{id}", mediaHandler.DeletePhoto())

					r.Post("/media/voice/presign", mediaHandler.PresignVoice())
					r.Post("/media/voice", mediaHandler.AttachVoice())
					r.Delete("/media/voice/{id}", mediaHandler.DeleteVoice())
				*/

				// Current user
				r.Route("/users/me", func(r chi.Router) {
					r.Get("/", userHandler.MyProfile()) // full profile (with visibility flags)
				})
			})
			/*r.Route(
				"/notifications", func(r chi.Router) {
					r.Use(haerdmiddleware.AuthMiddleware([]byte(jwtSecret))) // Protected endpoints: wrap these with auth middleware.
					r.Post("/device-token", notificationsHandler.RegisterDeviceToken())
					r.Post("/push-test", notificationsHandler.TestPushNotification()) // for testing in postman
				},
			)*/
		},
	)

	return router
}

func registerAliveEndpoint(router *chi.Mux) {
	router.Get("/alive", func(w http.ResponseWriter, r *http.Request) {
		// Return a simple status message.
		render.Json(w, http.StatusOK, "API is alive!")
	})
}
