package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/auth"
	"github.com/Haerd-Limited/dating-api/internal/api/user"
	auth2 "github.com/Haerd-Limited/dating-api/internal/auth"
	haerdMiddleware "github.com/Haerd-Limited/dating-api/internal/middleware"
	user2 "github.com/Haerd-Limited/dating-api/internal/user"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
)

func New(
	logger *zap.Logger,
	jwtSecret string,
	authService auth2.AuthService,
	userService user2.Service,
) http.Handler {
	// Create a new Chi router.
	router := chi.NewRouter()

	// Add middleware.
	router.Use(middleware.Logger)    // logs every request
	router.Use(middleware.Recoverer) // recovers from panics

	authHandler := auth.NewAuthHandler(logger, authService)
	userHandler := user.NewUserHandler(logger, userService)
	//notificationsHandler := notification.NewNotificationHandler(logger, notificationService)

	// Define the /alive endpoint.
	registerAliveEndpoint(router)
	router.Route(
		"/api/v1", func(r chi.Router) {
			r.Route(
				"/auth", func(r chi.Router) {
					r.Post("/register", authHandler.Register())
					r.Post("/login", authHandler.Login())
					r.Post("/refresh", authHandler.Refresh())
					r.Post("/logout", authHandler.Logout())
				},
			)
			/*r.Route(
				"/notifications", func(r chi.Router) {
					r.Use(haerdMiddleware.AuthMiddleware([]byte(jwtSecret))) // Protected endpoints: wrap these with auth middleware.
					r.Post("/device-token", notificationsHandler.RegisterDeviceToken())
					r.Post("/push-test", notificationsHandler.TestPushNotification()) // for testing in postman
				},
			)*/

			r.Route(
				"/users", func(r chi.Router) {
					r.Use(haerdMiddleware.AuthMiddleware([]byte(jwtSecret))) // Protected endpoints: wrap these with auth middleware.
					r.Get("/{userID}/profile", userHandler.ViewProfile())    // anyone's profile

					r.Route("/me", func(r chi.Router) {
						r.Get("/", userHandler.MyProfile())       // logged-in user's own profile
						r.Patch("/", userHandler.UpdateProfile()) // PUT or PATCH depending on your style
					})
				},
			)

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
