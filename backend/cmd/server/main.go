package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"margin.at/internal/api"
	"margin.at/internal/db"
	"margin.at/internal/firehose"
	internalMiddleware "margin.at/internal/middleware"
	"margin.at/internal/oauth"
	"margin.at/internal/sync"
)

func main() {
	godotenv.Load("../.env", ".env")

	database, err := db.New(getEnv("DATABASE_URL", "margin.db"))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	if err := database.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	syncSvc := sync.NewService(database)

	oauthHandler, err := oauth.NewHandler(database, syncSvc)
	if err != nil {
		log.Fatalf("Failed to initialize OAuth: %v", err)
	}

	ingester := firehose.NewIngester(database, syncSvc)
	firehose.RelayURL = getEnv("BLOCK_RELAY_URL", "wss://jetstream2.us-east.bsky.network/subscribe")
	log.Printf("Firehose URL: %s", firehose.RelayURL)

	go func() {
		if err := ingester.Start(context.Background()); err != nil {
			log.Printf("Firehose ingester error: %v", err)
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

	handler := api.NewHandler(database, annotationSvc, tokenRefresher, syncSvc)
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

	staticDir := getEnv("STATIC_DIR", "../web/dist")
	serveStatic(r, staticDir)

	port := getEnv("PORT", "8080")
	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	baseURL := getEnv("BASE_URL", "http://localhost:"+port)
	go func() {
		log.Printf("🚀 Margin server running on %s", baseURL)
		log.Printf("📝 App: %s", baseURL)
		log.Printf("🔗 API: %s/api/annotations", baseURL)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ingester.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func serveStatic(r chi.Router, staticDir string) {
	absPath, err := filepath.Abs(staticDir)
	if err != nil {
		log.Printf("Warning: Could not resolve static directory: %v", err)
		return
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		log.Printf("Warning: Static directory does not exist: %s", absPath)
		log.Printf("Run 'npm run build' in the web directory first")
		return
	}

	log.Printf("📂 Serving static files from: %s", absPath)

	fileServer := http.FileServer(http.Dir(absPath))

	r.Get("/*", func(w http.ResponseWriter, req *http.Request) {
		path := req.URL.Path

		if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/auth/") {
			http.NotFound(w, req)
			return
		}

		filePath := filepath.Join(absPath, path)
		if _, err := os.Stat(filePath); err == nil {
			fileServer.ServeHTTP(w, req)
			return
		}

		if strings.HasPrefix(path, "/.well-known/") {
			http.NotFound(w, req)
			return
		}

		lastSlash := strings.LastIndex(path, "/")
		lastSegment := path
		if lastSlash >= 0 {
			lastSegment = path[lastSlash+1:]
		}

		staticExts := []string{".js", ".css", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".woff", ".woff2", ".ttf", ".eot", ".map"}
		for _, ext := range staticExts {
			if strings.HasSuffix(lastSegment, ext) {
				http.NotFound(w, req)
				return
			}
		}

		http.ServeFile(w, req, filepath.Join(absPath, "index.html"))
	})
}
