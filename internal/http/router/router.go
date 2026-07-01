package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/adminrealtime"
	internaladminsession "github.com/Haerd-Limited/dating-api/internal/adminsession"
	apiadminrealtime "github.com/Haerd-Limited/dating-api/internal/api/adminrealtime"
	apiadminsession "github.com/Haerd-Limited/dating-api/internal/api/adminsession"
	apiauditlog "github.com/Haerd-Limited/dating-api/internal/api/auditlog"
	"github.com/Haerd-Limited/dating-api/internal/api/auth"
	apibroadcast "github.com/Haerd-Limited/dating-api/internal/api/broadcast"
	"github.com/Haerd-Limited/dating-api/internal/api/compatibility"
	apiconsent "github.com/Haerd-Limited/dating-api/internal/api/consent"
	"github.com/Haerd-Limited/dating-api/internal/api/conversation"
	"github.com/Haerd-Limited/dating-api/internal/api/discover"
	apifeedback "github.com/Haerd-Limited/dating-api/internal/api/feedback"
	apiinsights "github.com/Haerd-Limited/dating-api/internal/api/insights"
	"github.com/Haerd-Limited/dating-api/internal/api/interaction"
	"github.com/Haerd-Limited/dating-api/internal/api/lookup"
	"github.com/Haerd-Limited/dating-api/internal/api/media"
	apinotification "github.com/Haerd-Limited/dating-api/internal/api/notification"
	"github.com/Haerd-Limited/dating-api/internal/api/onboarding"
	"github.com/Haerd-Limited/dating-api/internal/api/profile"
	"github.com/Haerd-Limited/dating-api/internal/api/realtime"
	apisafety "github.com/Haerd-Limited/dating-api/internal/api/safety"
	"github.com/Haerd-Limited/dating-api/internal/api/verification"
	internalauditlog "github.com/Haerd-Limited/dating-api/internal/auditlog"
	internalauth "github.com/Haerd-Limited/dating-api/internal/auth"
	internalbroadcast "github.com/Haerd-Limited/dating-api/internal/broadcast"
	internalcompatibility "github.com/Haerd-Limited/dating-api/internal/compatibility"
	internalconsent "github.com/Haerd-Limited/dating-api/internal/consent"
	internalconversation "github.com/Haerd-Limited/dating-api/internal/conversation"
	internaldataexport "github.com/Haerd-Limited/dating-api/internal/dataexport"
	internaldiscover "github.com/Haerd-Limited/dating-api/internal/discover"
	internalfeedback "github.com/Haerd-Limited/dating-api/internal/feedback"
	internalinsights "github.com/Haerd-Limited/dating-api/internal/insights"
	internalinteraction "github.com/Haerd-Limited/dating-api/internal/interaction"
	internallookup "github.com/Haerd-Limited/dating-api/internal/lookup"
	internalmedia "github.com/Haerd-Limited/dating-api/internal/media"
	haerdmiddleware "github.com/Haerd-Limited/dating-api/internal/middleware"
	internalnotification "github.com/Haerd-Limited/dating-api/internal/notification"
	internalonboarding "github.com/Haerd-Limited/dating-api/internal/onboarding"
	internalpreference "github.com/Haerd-Limited/dating-api/internal/preference"
	internalprofile "github.com/Haerd-Limited/dating-api/internal/profile"
	internalrealtime "github.com/Haerd-Limited/dating-api/internal/realtime"
	internalsafety "github.com/Haerd-Limited/dating-api/internal/safety"
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
	compatibilityService internalcompatibility.Service,
	adminCompatibilityService internalcompatibility.AdminService,
	notificationService internalnotification.Service,
	safetyService internalsafety.Service,
	insightsService internalinsights.Service,
	userService internaluser.Service,
	feedbackService internalfeedback.Service,
	broadcastService internalbroadcast.Service,
	dataExportService internaldataexport.Service,
	preferenceService internalpreference.Service,
	consentService internalconsent.Service,
	enableConsentGate bool,
	adminAPIKey string,
	auditLogService internalauditlog.Service,
	adminSessionService internaladminsession.Service,
	adminHub *adminrealtime.Hub,
	adminPresence *adminrealtime.PresenceStore,
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
			"http://localhost:5173",
			"https://admin-dashboard-sit.up.railway.app",
			"https://admin-dashboard-prod.up.railway.app",
			"https://haerd-admin-dashboard.up.railway.app",
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Admin-Token", "X-Admin-Session"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// Add middleware.
	router.Use(middleware.Logger)    // logs every request
	router.Use(middleware.Recoverer) // recovers from panics

	authHandler := auth.NewAuthHandler(logger, authService)
	profileHandler := profile.NewProfileHandler(logger, profileService, userService, dataExportService, discoverService, preferenceService)
	consentHandler := apiconsent.NewConsentHandler(logger, consentService)
	onboardingHandler := onboarding.NewOnboardingHandler(logger, onboardingService)
	notificationHandler := apinotification.NewNotificationHandler(logger, notificationService)
	discoverHandler := discover.NewDiscoverHandler(logger, discoverService)
	interactionHandler := interaction.NewInteractionHandler(logger, interactionService)
	conversationHandler := conversation.NewConversationHandler(logger, conversationService)
	mediaHandler := media.NewMediaHandler(logger, mediaService)
	lookupHandler := lookup.NewLookupHandler(logger, lookupService)
	wsHandler := realtime.NewWsHandler(logger, hub, conversationService)
	verificationHandler := verification.NewVerificationHandler(logger, verificationService)
	compatibilityHandler := compatibility.NewCompatibilityHandler(logger, compatibilityService)
	adminCompatibilityHandler := compatibility.NewAdminCompatibilityHandler(logger, adminCompatibilityService)
	safetyHandler := apisafety.NewHandler(logger, safetyService)
	insightsHandler := apiinsights.NewHandler(logger, insightsService)
	feedbackHandler := apifeedback.NewFeedbackHandler(logger, feedbackService)
	broadcastHandler := apibroadcast.NewHandler(logger, broadcastService)
	adminSessionHandler := apiadminsession.NewHandler(logger, adminSessionService)
	adminRealtimeHandler := apiadminrealtime.NewHandler(logger, adminHub, adminPresence)
	auditLogHandler := apiauditlog.NewHandler(logger, auditLogService)

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
			r.Post("/media/transcribe-reel", mediaHandler.TranscribeInstagramReel())

			r.Route("/lookup", func(r chi.Router) {
				r.Get("/prompts", lookupHandler.GetPrompts())
				r.Get("/languages", lookupHandler.GetLanguages())
				r.Get("/religions", lookupHandler.GetReligions())
				r.Get("/political-beliefs", lookupHandler.GetPoliticalBeliefs())
				r.Get("/ethnicities", lookupHandler.GetEthnicities())
				r.Get("/genders", lookupHandler.GetGenders())
				r.Get("/sexualities", lookupHandler.GetSexualities())
				r.Get("/habits", lookupHandler.GetHabits())
				r.Get("/education-levels", lookupHandler.GetEducationLevels())
				r.Get("/family-plans", lookupHandler.GetFamilyPlans())
				r.Get("/family-status", lookupHandler.GetFamilyStatus())
				r.Get("/report-categories", lookupHandler.GetReportCategories())
			})

			// --- Protected (must be logged in)
			r.Group(func(r chi.Router) {
				r.Use(haerdmiddleware.AuthMiddleware([]byte(jwtSecret)))

				// Carve-outs that bypass the consent gate (Articles 15, 17 + consent endpoints).
				r.Route("/consents", func(r chi.Router) {
					r.Post("/", consentHandler.Record())
					r.Get("/", consentHandler.List())
					r.Delete("/{type}", consentHandler.Revoke())
				})
				r.Get("/users/me/data-export", profileHandler.GetDataExport())
				r.Delete("/users/me", profileHandler.DeleteAccount())
				r.Get("/users/me/account-status", safetyHandler.GetMyAccountStatus())
				r.Get("/users/me/warnings", safetyHandler.ListMyWarnings())
				r.Post("/users/me/warnings/{warningID}/acknowledge", safetyHandler.AcknowledgeWarning())
				r.Route("/notifications", func(r chi.Router) {
					r.Post("/device-tokens", notificationHandler.RegisterDeviceToken())
					r.Delete("/device-tokens", notificationHandler.RemoveDeviceToken())
				})

				r.Group(func(r chi.Router) {
					r.Use(haerdmiddleware.ConsentRequired(consentService, enableConsentGate, logger))
					r.Use(haerdmiddleware.AccountStatusRequired(userService, logger))
					r.Use(haerdmiddleware.AppOpenTracker())

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
							r.Post("/video-verification", onboardingHandler.VideoVerification())
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
						r.Post("/favourites", interactionHandler.AddFavourite())
						r.Delete("/favourites/{watchedUserID}", interactionHandler.RemoveFavourite())
					})

					r.Route("/insights", func(r chi.Router) {
						r.Get("/me/weekly", insightsHandler.GetMeWeekly())
						r.Get("/me/wrapped", insightsHandler.GetMyWrapped())
					})

					// router.go
					r.Route("/rt", func(r chi.Router) {
						r.Get("/ws", wsHandler.ServeWS())
					})

					r.Route("/conversations", func(r chi.Router) {
						r.Get("/", conversationHandler.GetConversations())
						r.Get("/{id}/messages", conversationHandler.GetConversationMessages())
						r.Post("/{id}/messages", conversationHandler.SendMessage())
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

					r.Route("/matching", func(r chi.Router) { // TODO: change to compatibility
						r.Get("/overview", compatibilityHandler.GetOverview())
						r.Get("/questions", compatibilityHandler.GetQuestions())
						r.Get("/compatibility", compatibilityHandler.GetCompatibility())
						r.Post("/answers", compatibilityHandler.SaveAnswer())
					})

					// Current user
					r.Route("/users/me", func(r chi.Router) {
						r.Get("/", profileHandler.GetMyProfile())
						r.Patch("/", profileHandler.UpdateMyProfile())
						r.Patch("/preferences/analytics-opt-out", profileHandler.SetAnalyticsOptOut())
						r.Patch("/verify", profileHandler.Verify()) // Simple version that doesnt use AWS. call when not in uat/prod.

						r.Route("/verification", func(r chi.Router) {
							r.Post("/photo/start", verificationHandler.Start()) // returns {session_id, region}
							r.Post("/photo/complete", verificationHandler.Complete())
							r.Post("/video/start", verificationHandler.StartVideo())   // returns {code, upload_url, upload_key}
							r.Post("/video/submit", verificationHandler.SubmitVideo()) // submits video for review
						})
					})

					// User profiles by ID
					r.Route("/users", func(r chi.Router) {
						r.Get("/{userID}", profileHandler.GetUserProfile())
						r.Get("/{userID}/voice-prompts/transcripts", profileHandler.GetUserPromptTranscripts())
					})
				})
			})

			r.Route("/admin", func(r chi.Router) {
				r.With(haerdmiddleware.AdminSessionFromQuery(adminSessionService)).Get("/stream", adminRealtimeHandler.Stream())

				r.Group(func(r chi.Router) {
					r.Use(haerdmiddleware.AdminMiddleware(adminAPIKey))

					r.Get("/roster", adminSessionHandler.GetRoster())
					r.Post("/session", adminSessionHandler.CreateSession())
				})

				r.Group(func(r chi.Router) {
					r.Use(haerdmiddleware.AdminMiddleware(adminAPIKey))
					r.Use(haerdmiddleware.AdminSession(adminSessionService))
					r.Use(haerdmiddleware.AdminAudit(auditLogService, logger))

					r.Delete("/session", adminSessionHandler.DeleteSession())
					r.Get("/events", auditLogHandler.ListEvents())
					r.Post("/presence", adminRealtimeHandler.SetPresence())
					r.Delete("/presence", adminRealtimeHandler.ClearPresence())

					r.Route("/reports", func(r chi.Router) {
						r.Get("/", safetyHandler.AdminListReports())
						r.Get("/{reportID}", safetyHandler.AdminGetReport())
						r.Post("/{reportID}/resolve", safetyHandler.AdminResolveReport())
					})

					r.Route("/analytics", func(r chi.Router) {
						r.Route("/retention", func(r chi.Router) {
							r.Get("/", insightsHandler.GetRetentionStats())
							r.Get("/cohorts", insightsHandler.GetRetentionCohorts())
							r.Get("/global", insightsHandler.GetGlobalRetentionStats())
							r.Get("/users/{userID}", insightsHandler.GetUserRetentionProfile())
						})
					})

					r.Route("/verification", func(r chi.Router) {
						r.Route("/video-attempts", func(r chi.Router) {
							r.Get("/", verificationHandler.AdminListVideoAttempts())
							r.Get("/{attemptID}", verificationHandler.AdminGetVideoAttempt())
							r.Get("/{attemptID}/video-url", verificationHandler.AdminGetVideoDownloadURL())
							r.Post("/{attemptID}/approve", verificationHandler.AdminApproveVideoAttempt())
							r.Post("/{attemptID}/reject", verificationHandler.AdminRejectVideoAttempt())
						})
					})

					r.Route("/waitlist", func(r chi.Router) {
						r.Get("/users", broadcastHandler.ListWaitlistUsers())
						r.Post("/broadcast", broadcastHandler.SendBroadcast())
					})

					r.Route("/question-packs", func(r chi.Router) {
						r.Get("/categories", adminCompatibilityHandler.ListCategories())
						r.Post("/categories", adminCompatibilityHandler.CreateCategory())
						r.Post("/categories/reorder", adminCompatibilityHandler.ReorderCategories())
						r.Patch("/categories/{categoryID}", adminCompatibilityHandler.UpdateCategory())
						r.Delete("/categories/{categoryID}", adminCompatibilityHandler.DeleteCategory())

						r.Get("/categories/{categoryID}/questions", adminCompatibilityHandler.ListQuestions())
						r.Post("/categories/{categoryID}/questions", adminCompatibilityHandler.CreateQuestion())
						r.Post("/categories/{categoryID}/questions/reorder", adminCompatibilityHandler.ReorderQuestions())
						r.Patch("/questions/{questionID}", adminCompatibilityHandler.UpdateQuestion())
						r.Delete("/questions/{questionID}", adminCompatibilityHandler.DeleteQuestion())

						r.Get("/questions/{questionID}/answers", adminCompatibilityHandler.ListAnswers())
						r.Post("/questions/{questionID}/answers", adminCompatibilityHandler.CreateAnswer())
						r.Post("/questions/{questionID}/answers/reorder", adminCompatibilityHandler.ReorderAnswers())
						r.Patch("/answers/{answerID}", adminCompatibilityHandler.UpdateAnswer())
						r.Delete("/answers/{answerID}", adminCompatibilityHandler.DeleteAnswer())
					})
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
