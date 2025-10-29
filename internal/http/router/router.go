package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/auth"
	"github.com/Haerd-Limited/dating-api/internal/api/conversation"
	"github.com/Haerd-Limited/dating-api/internal/api/discover"
	"github.com/Haerd-Limited/dating-api/internal/api/interaction"
	"github.com/Haerd-Limited/dating-api/internal/api/lookup"
	"github.com/Haerd-Limited/dating-api/internal/api/matching"
	"github.com/Haerd-Limited/dating-api/internal/api/media"
	"github.com/Haerd-Limited/dating-api/internal/api/onboarding"
	"github.com/Haerd-Limited/dating-api/internal/api/profile"
	"github.com/Haerd-Limited/dating-api/internal/api/realtime"
	"github.com/Haerd-Limited/dating-api/internal/api/verification"
	internalauth "github.com/Haerd-Limited/dating-api/internal/auth"
	internalconversation "github.com/Haerd-Limited/dating-api/internal/conversation"
	internaldiscover "github.com/Haerd-Limited/dating-api/internal/discover"
	internalinteraction "github.com/Haerd-Limited/dating-api/internal/interaction"
	internallookup "github.com/Haerd-Limited/dating-api/internal/lookup"
	internalmatching "github.com/Haerd-Limited/dating-api/internal/matching"
	internalmedia "github.com/Haerd-Limited/dating-api/internal/media"
	haerdmiddleware "github.com/Haerd-Limited/dating-api/internal/middleware"
	internalonboarding "github.com/Haerd-Limited/dating-api/internal/onboarding"
	internalprofile "github.com/Haerd-Limited/dating-api/internal/profile"
	internalrealtime "github.com/Haerd-Limited/dating-api/internal/realtime"
	internalverification "github.com/Haerd-Limited/dating-api/internal/verification"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
)

func New(
	logger *zap.Logger,
	jwtSecret string,
	authService internalauth.Service,
	onboardingService internalonboarding.Service,
	profileService internalprofile.Service,
	discoverService internaldiscover.Service,
	interactionService internalinteraction.Service,
	conversationService internalconversation.Service,
	mediaService internalmedia.Service,
	lookupService internallookup.Service,
	hub *internalrealtime.Hub,
	verificationService internalverification.Service,
	matchingService internalmatching.Service,
) http.Handler {
	// Create a new Chi router.
	router := chi.NewRouter()

	// Add middleware.
	router.Use(middleware.Logger)    // logs every request
	router.Use(middleware.Recoverer) // recovers from panics

	authHandler := auth.NewAuthHandler(logger, authService)
	profileHandler := profile.NewProfileHandler(logger, profileService)
	onboardingHandler := onboarding.NewOnboardingHandler(logger, onboardingService)
	discoverHandler := discover.NewDiscoverHandler(logger, discoverService)
	interactionHandler := interaction.NewInteractionHandler(logger, interactionService)
	conversationHandler := conversation.NewConversationHandler(logger, conversationService)
	mediaHandler := media.NewMediaHandler(logger, mediaService)
	lookupHandler := lookup.NewLookupHandler(logger, lookupService)
	wsHandler := realtime.NewWsHandler(logger, hub, conversationService)
	verificationHandler := verification.NewVerificationHandler(logger, verificationService)
	matchingHandler := matching.NewMatchingHandler(logger, matchingService)
	// notificationsHandler := notification.NewNotificationHandler(logger, notificationService)

	// Define the /alive endpoint.
	registerAliveEndpoint(router)
	router.Route(
		"/api/v1", func(r chi.Router) {
			r.Route(
				"/auth", func(r chi.Router) {
					r.Post("/request-code", authHandler.RequestCode())
					r.Post("/verify-code", authHandler.VerifyCode())
					r.Post("/refresh", authHandler.Refresh())
					r.Post("/logout", authHandler.Logout())
				},
			)

			// --- Protected (must be logged in)
			r.Group(func(r chi.Router) {
				r.Use(haerdmiddleware.AuthMiddleware([]byte(jwtSecret)))

				r.Route("/lookup", func(r chi.Router) {
					r.Get("/prompts", lookupHandler.GetPrompts())
					r.Get("/languages", lookupHandler.GetLanguages())
					r.Get("/religions", lookupHandler.GetReligions())
					r.Get("/political-beliefs", lookupHandler.GetPoliticalBeliefs())
					r.Get("/ethnicities", lookupHandler.GetEthnicities())
					r.Get("/genders", lookupHandler.GetGenders())
					r.Get("/habits", lookupHandler.GetHabits())
					r.Get("/dating-intentions", lookupHandler.GetDatingIntentions())
					r.Get("/education-levels", lookupHandler.GetEducationLevels())
					r.Get("/family-plans", lookupHandler.GetFamilyPlans())
					r.Get("/family-status", lookupHandler.GetFamilyStatus())
				})

				r.Route(
					"/onboarding", func(r chi.Router) {
						r.Get("/step", onboardingHandler.GetStep())
						r.Post("/intro", onboardingHandler.Intro())
						r.Patch("/basics", onboardingHandler.Basics())
						r.Patch("/location", onboardingHandler.Location())
						r.Patch("/lifestyle", onboardingHandler.Lifestyle())
						r.Patch("/beliefs", onboardingHandler.Beliefs())
						r.Patch("/background", onboardingHandler.Background())
						r.Patch("/work-and-education", onboardingHandler.WorkAndEducation())
						r.Post("/languages", onboardingHandler.Languages())
						r.Post("/photos", onboardingHandler.Photos())
						r.Post("/prompts", onboardingHandler.Prompts())
						r.Post("/profile", onboardingHandler.Profile())
					},
				)

				r.Route("/discover", func(r chi.Router) {
					r.Get("/", discoverHandler.GetDiscover())
				})
				r.Route("/vwh", func(r chi.Router) {
					r.Get("/", discoverHandler.GetVoiceWorthHearing())
				})
				r.Route("/swipes", func(r chi.Router) {
					r.Post("/", interactionHandler.Create())
				})

				r.Route("/likes", func(r chi.Router) {
					r.Get("/", interactionHandler.GetLikes())
				})

				// router.go
				r.Route("/rt", func(r chi.Router) {
					r.Get("/ws", wsHandler.ServeWS())
				})

				r.Route("/conversations", func(r chi.Router) {
					r.Get("/", conversationHandler.GetConversations())
					r.Get("/{id}/score", conversationHandler.GetConversationScore())
					r.Get("/{id}/messages", conversationHandler.GetConversationMessages())
					r.Post("/{id}/messages", conversationHandler.SendMessage())
					r.Post("/{id}/reveal/initiate", conversationHandler.InitiateReveal())
					r.Post("/{id}/reveal/confirm", conversationHandler.ConfirmReveal())
					r.Post("/{id}/reveal/decision", conversationHandler.MakeRevealDecision())
					r.Get("/{id}/photos", conversationHandler.GetMatchPhotos())
				})

				r.Route("/media", func(r chi.Router) {
					r.Get("/photos/presign", mediaHandler.GeneratePhotoUploadUrl()) // returns URL/fields
					// r.Post("/media/photos", mediaHandler.AttachPhoto())          // save URL, position, is_primary
					// r.Patch("/media/photos/{id}", mediaHandler.UpdatePhoto())    // reorder / set primary
					// r.Delete("/media/photos/{id}", mediaHandler.DeletePhoto())

					r.Get("/voice/presign", mediaHandler.GenerateVoiceNoteUploadUrl())
					// r.Delete("/media/voice/{id}", mediaHandler.DeleteVoice())
				})

				r.Route("/matching", func(r chi.Router) {
					r.Get("/overview", matchingHandler.GetOverview())
					r.Get("/questions", matchingHandler.GetQuestions())
					r.Post("/answers", matchingHandler.SaveAnswer())
				})

				// Current user
				r.Route("/users/me", func(r chi.Router) {
					r.Get("/", profileHandler.GetMyProfile())
					r.Patch("/", profileHandler.UpdateMyProfile())
					r.Patch("/verify", profileHandler.Verify()) // Simple version that doesnt use AWS. call when not in uat/prod.

					r.Route("/verification", func(r chi.Router) {
						r.Post("/photo/start", verificationHandler.Start()) // returns {session_id, region}
						r.Post("/photo/complete", verificationHandler.Complete())
					})
					// TODO(high-priority): create delete account endpoint that deletes all user data from DB and S3 bucket
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
