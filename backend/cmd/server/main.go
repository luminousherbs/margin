package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"margin.at/internal/api"
	"margin.at/internal/db"
	"margin.at/internal/embeddings"
	"margin.at/internal/firehose"
	"margin.at/internal/logger"
	internalMiddleware "margin.at/internal/middleware"
	"margin.at/internal/oauth"
	"margin.at/internal/recommendations"
	"margin.at/internal/sync"
)

func main() {
	godotenv.Load("../.env", ".env")

	database, err := db.New(getEnv("DATABASE_URL", "margin.db"))
	if err != nil {
		logger.Fatal("Failed to connect to database: %v", err)
	}
	defer database.Close()

	if err := database.Migrate(); err != nil {
		logger.Fatal("Failed to run migrations: %v", err)
	}

	embeddingClient := embeddings.NewClient()
	if err := database.MigrateRecommendations(); err != nil {
		logger.Fatal("Failed to run recommendation migrations: %v", err)
	}
	recService := recommendations.NewService(database, embeddingClient)
	logger.Info("Recommendation engine initialized (embeddings enabled: %v)", embeddingClient.IsEnabled())

	syncSvc := sync.NewService(database)

	oauthHandler, err := oauth.NewHandler(database, syncSvc)
	if err != nil {
		logger.Fatal("Failed to initialize OAuth: %v", err)
	}

	ingester := firehose.NewIngester(database, syncSvc)
	firehose.RelayURL = getEnv("BLOCK_RELAY_URL", "wss://jetstream2.us-east.bsky.network/subscribe")
	logger.Info("Firehose URL: %s", firehose.RelayURL)

	if recService.IsEnabled() {
		ingester.SetOnAnnotation(recService.OnAnnotation)
		ingester.SetOnDocument(recService.OnDocument)

		go func() {
			logger.Info("Starting recommendation backfill...")
			if err := recService.BackfillDocumentEmbeddings(200); err != nil {
				logger.Error("Document embedding backfill error: %v", err)
			}
			annCount, err := recService.BackfillAnnotationEmbeddings(200)
			if err != nil {
				logger.Error("Annotation embedding backfill error: %v", err)
			}
			hlCount, err := recService.BackfillHighlightEmbeddings(200)
			if err != nil {
				logger.Error("Highlight embedding backfill error: %v", err)
			}
			profileCount, err := recService.RebuildAllProfiles()
			if err != nil {
				logger.Error("Profile rebuild error: %v", err)
			}
			logger.Info("Recommendation backfill complete (annotations: %d, highlights: %d, profiles: %d)", annCount, hlCount, profileCount)
		}()
	}

	go func() {
		if err := ingester.Start(context.Background()); err != nil {
			logger.Error("Firehose ingester error: %v", err)
		}
	}()

	r := chi.NewRouter()

	r.Use(internalMiddleware.PrivacyLogger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middleware.Throttle(100))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*", "chrome-extension://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Session-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	tokenRefresher := api.NewTokenRefresher(database, oauthHandler.GetPrivateKey())
	annotationSvc := api.NewAnnotationService(database, tokenRefresher)

	handler := api.NewHandler(database, annotationSvc, tokenRefresher, syncSvc, recService)
	handler.RegisterRoutes(r)

	r.Post("/api/annotations", annotationSvc.CreateAnnotation)
	r.Put("/api/annotations", annotationSvc.UpdateAnnotation)
	r.Delete("/api/annotations", annotationSvc.DeleteAnnotation)
	r.Post("/api/annotations/like", annotationSvc.LikeAnnotation)
	r.Delete("/api/annotations/like", annotationSvc.UnlikeAnnotation)
	r.Post("/api/annotations/reply", annotationSvc.CreateReply)
	r.Delete("/api/annotations/reply", annotationSvc.DeleteReply)
	r.Post("/api/highlights", annotationSvc.CreateHighlight)
	r.Put("/api/highlights", annotationSvc.UpdateHighlight)
	r.Delete("/api/highlights", annotationSvc.DeleteHighlight)
	r.Post("/api/bookmarks", annotationSvc.CreateBookmark)
	r.Put("/api/bookmarks", annotationSvc.UpdateBookmark)
	r.Delete("/api/bookmarks", annotationSvc.DeleteBookmark)

	r.Route("/auth", func(r chi.Router) {
		r.Use(middleware.Throttle(10))
		r.Get("/login", oauthHandler.HandleLogin)
		r.Post("/start", oauthHandler.HandleStart)
		r.Post("/signup", oauthHandler.HandleSignup)
		r.Get("/callback", oauthHandler.HandleCallback)
		r.Post("/logout", oauthHandler.HandleLogout)
		r.Get("/session", oauthHandler.HandleSession)
	})
	r.Get("/client-metadata.json", oauthHandler.HandleClientMetadata)
	r.Get("/jwks.json", oauthHandler.HandleJWKS)

	r.Get("/api/tags/trending", handler.HandleGetTrendingTags)
	r.Put("/api/profile", handler.UpdateProfile)
	r.Get("/api/profile/{did}", handler.GetProfile)
	r.Post("/api/profile/avatar", handler.UploadAvatar)

	port := getEnv("PORT", "8081")
	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		logger.Info("Margin API server running on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Infoln("Shutting down server...")
	ingester.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown: %v", err)
	}

	logger.Infoln("Server exited")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
