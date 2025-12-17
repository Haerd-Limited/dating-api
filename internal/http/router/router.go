package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/auth"
	"github.com/Haerd-Limited/dating-api/internal/api/conversation"
	"github.com/Haerd-Limited/dating-api/internal/api/discover"
	apifeedback "github.com/Haerd-Limited/dating-api/internal/api/feedback"
	apiinsights "github.com/Haerd-Limited/dating-api/internal/api/insights"
	"github.com/Haerd-Limited/dating-api/internal/api/interaction"
	"github.com/Haerd-Limited/dating-api/internal/api/lookup"
	"github.com/Haerd-Limited/dating-api/internal/api/matching"
	"github.com/Haerd-Limited/dating-api/internal/api/media"
	apinotification "github.com/Haerd-Limited/dating-api/internal/api/notification"
	"github.com/Haerd-Limited/dating-api/internal/api/onboarding"
	"github.com/Haerd-Limited/dating-api/internal/api/profile"
	"github.com/Haerd-Limited/dating-api/internal/api/realtime"
	apisafety "github.com/Haerd-Limited/dating-api/internal/api/safety"
	apisession "github.com/Haerd-Limited/dating-api/internal/api/session"
	"github.com/Haerd-Limited/dating-api/internal/api/verification"
	internalauth "github.com/Haerd-Limited/dating-api/internal/auth"
	internalconversation "github.com/Haerd-Limited/dating-api/internal/conversation"
	internaldiscover "github.com/Haerd-Limited/dating-api/internal/discover"
	internalfeedback "github.com/Haerd-Limited/dating-api/internal/feedback"
	internalinsights "github.com/Haerd-Limited/dating-api/internal/insights"
	internalinteraction "github.com/Haerd-Limited/dating-api/internal/interaction"
	internallookup "github.com/Haerd-Limited/dating-api/internal/lookup"
	internalmatching "github.com/Haerd-Limited/dating-api/internal/matching"
	internalmedia "github.com/Haerd-Limited/dating-api/internal/media"
	haerdmiddleware "github.com/Haerd-Limited/dating-api/internal/middleware"
	internalnotification "github.com/Haerd-Limited/dating-api/internal/notification"
	internalonboarding "github.com/Haerd-Limited/dating-api/internal/onboarding"
	internalprofile "github.com/Haerd-Limited/dating-api/internal/profile"
	internalrealtime "github.com/Haerd-Limited/dating-api/internal/realtime"
	internalsafety "github.com/Haerd-Limited/dating-api/internal/safety"
	internalsession "github.com/Haerd-Limited/dating-api/internal/session"
	internaluser "github.com/Haerd-Limited/dating-api/internal/user"
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
	notificationService internalnotification.Service,
	safetyService internalsafety.Service,
	insightsService internalinsights.Service,
	userService internaluser.Service,
	feedbackService internalfeedback.Service,
	sessionService internalsession.Service,
	adminAPIKey string,
) http.Handler {
	// Create a new Chi router.
	router := chi.NewRouter()

	// Add CORS middleware - must be before other middleware
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"http://haerd.com",
			"https://haerd.com",
			"http://localhost:3000",
			"http://localhost:3001",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:3001",
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Admin-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// Add middleware.
	router.Use(middleware.Logger)    // logs every request
	router.Use(middleware.Recoverer) // recovers from panics

	authHandler := auth.NewAuthHandler(logger, authService)
	profileHandler := profile.NewProfileHandler(logger, profileService, userService)
	onboardingHandler := onboarding.NewOnboardingHandler(logger, onboardingService)
	notificationHandler := apinotification.NewNotificationHandler(logger, notificationService)
	discoverHandler := discover.NewDiscoverHandler(logger, discoverService)
	interactionHandler := interaction.NewInteractionHandler(logger, interactionService)
	conversationHandler := conversation.NewConversationHandler(logger, conversationService)
	mediaHandler := media.NewMediaHandler(logger, mediaService)
	lookupHandler := lookup.NewLookupHandler(logger, lookupService)
	wsHandler := realtime.NewWsHandler(logger, hub, conversationService)
	verificationHandler := verification.NewVerificationHandler(logger, verificationService)
	matchingHandler := matching.NewMatchingHandler(logger, matchingService)
	safetyHandler := apisafety.NewHandler(logger, safetyService)
	insightsHandler := apiinsights.NewHandler(logger, insightsService)
	feedbackHandler := apifeedback.NewFeedbackHandler(logger, feedbackService)
	sessionHandler := apisession.NewSessionHandler(logger, sessionService)

	// Define the /alive endpoint.
	registerAliveEndpoint(router)
	router.Route(
		"/api/v1", func(r chi.Router) {
			// Public endpoints
			r.Get("/landing/stats", onboardingHandler.Stats())

			r.Get("/insights/public/weekly", insightsHandler.GetPublicWeekly())

			r.Route(
				"/auth", func(r chi.Router) {
					r.Post("/request-code", authHandler.RequestCode())
					r.Post("/verify-code", authHandler.VerifyCode())
					r.Post("/refresh", authHandler.Refresh())
					r.Post("/logout", authHandler.Logout())
				},
			)

			r.Route("/lookup", func(r chi.Router) {
				r.Get("/prompts", lookupHandler.GetPrompts())
				r.Get("/languages", lookupHandler.GetLanguages())
				r.Get("/religions", lookupHandler.GetReligions())
				r.Get("/political-beliefs", lookupHandler.GetPoliticalBeliefs())
				r.Get("/ethnicities", lookupHandler.GetEthnicities())
				r.Get("/genders", lookupHandler.GetGenders())
				r.Get("/sexualities", lookupHandler.GetSexualities())
				r.Get("/habits", lookupHandler.GetHabits())
				r.Get("/dating-intentions", lookupHandler.GetDatingIntentions())
				r.Get("/education-levels", lookupHandler.GetEducationLevels())
				r.Get("/family-plans", lookupHandler.GetFamilyPlans())
				r.Get("/family-status", lookupHandler.GetFamilyStatus())
				r.Get("/report-categories", lookupHandler.GetReportCategories())
			})

			// --- Protected (must be logged in)
			r.Group(func(r chi.Router) {
				r.Use(haerdmiddleware.AuthMiddleware([]byte(jwtSecret)))

				r.Route("/safety", func(r chi.Router) {
					r.Post("/block", safetyHandler.Block())
					r.Post("/report", safetyHandler.Report())
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
					r.Post("/filters", discoverHandler.GetDiscoverWithFilters())
					r.Get("/preferences", discoverHandler.GetUserPreferences())
				})
				r.Get("/voice-prompts/{id}/transcript", profileHandler.GetVoicePromptTranscript())
				r.Route("/vwh", func(r chi.Router) {
					r.Get("/", discoverHandler.GetVoiceWorthHearing())
				})
				r.Route("/swipes", func(r chi.Router) {
					r.Post("/", interactionHandler.Create())
				})

				r.Route("/likes", func(r chi.Router) {
					r.Get("/", interactionHandler.GetLikes())
				})

				r.Route("/insights", func(r chi.Router) {
					r.Get("/me/weekly", insightsHandler.GetMeWeekly())
					r.Get("/me/wrapped", insightsHandler.GetMyWrapped())
				})

				r.Route("/notifications", func(r chi.Router) {
					r.Post("/device-tokens", notificationHandler.RegisterDeviceToken())
					r.Delete("/device-tokens", notificationHandler.RemoveDeviceToken())
				})

				// router.go
				r.Route("/rt", func(r chi.Router) {
					r.Get("/ws", wsHandler.ServeWS())
				})

				r.Route("/conversations", func(r chi.Router) {
					r.Get("/", conversationHandler.GetConversations())
					r.Get("/{id}/score", conversationHandler.GetChemistryScore())
					r.Get("/{id}/messages", conversationHandler.GetConversationMessages())
					r.Post("/{id}/messages", conversationHandler.SendMessage())
					r.Post("/{id}/reveal/initiate", conversationHandler.InitiateReveal())
					r.Post("/{id}/reveal/confirm", conversationHandler.ConfirmReveal())
					r.Post("/{id}/reveal/decision", conversationHandler.MakeRevealDecision())
					r.Get("/{id}/photos", conversationHandler.GetMatchPhotos())
					r.Post("/{id}/unmatch", conversationHandler.Unmatch())
				})

				r.Route("/media", func(r chi.Router) {
					r.Get("/photos/presign", mediaHandler.GeneratePhotoUploadUrl()) // returns URL/fields
					// r.Post("/media/photos", mediaHandler.AttachPhoto())          // save URL, position, is_primary
					// r.Patch("/media/photos/{id}", mediaHandler.UpdatePhoto())    // reorder / set primary
					// r.Delete("/media/photos/{id}", mediaHandler.DeletePhoto())

					r.Get("/voice/presign", mediaHandler.GenerateVoiceNoteUploadUrl())
					// r.Delete("/media/voice/{id}", mediaHandler.DeleteVoice())

					r.Get("/feedback/presign", mediaHandler.GenerateFeedbackAttachmentUploadUrl())
				})

				r.Route("/feedback", func(r chi.Router) {
					r.Post("/", feedbackHandler.CreateFeedback())
				})

				r.Route("/session", func(r chi.Router) {
					r.Post("/track-open", sessionHandler.TrackAppOpen())
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
					r.Patch("/verify", profileHandler.Verify())   // Simple version that doesnt use AWS. call when not in uat/prod.
					r.Delete("/", profileHandler.DeleteAccount()) // Delete account and all user data

					r.Route("/verification", func(r chi.Router) {
						r.Post("/photo/start", verificationHandler.Start()) // returns {session_id, region}
						r.Post("/photo/complete", verificationHandler.Complete())
					})
				})

				// User profiles by ID
				r.Route("/users", func(r chi.Router) {
					r.Get("/{userID}", profileHandler.GetUserProfile())
				})
			})

			r.Route("/admin", func(r chi.Router) {
				r.Use(haerdmiddleware.AdminMiddleware(adminAPIKey))

				r.Route("/reports", func(r chi.Router) {
					r.Get("/", safetyHandler.AdminListReports())
					r.Get("/{reportID}", safetyHandler.AdminGetReport())
					r.Post("/{reportID}/resolve", safetyHandler.AdminResolveReport())
				})
			})
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
