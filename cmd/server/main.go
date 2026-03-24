package main

import (
	authMiddleware "backend/internal/middleware"
	"fmt"
	"log"
	stdhttp "net/http"
	"os"
	"strings"
	"time"

	"backend/internal/adapters"
	"backend/internal/adapters/alphavantage"
	"backend/internal/adapters/postgres"
	"backend/internal/adapters/yahoofinance"
	"backend/internal/application"
	http "backend/internal/handler"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func main() {
	// Loading .env file
	godotenv.Load()

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbSSLMode := getEnv("DB_SSLMODE", "disable")

	db, err := postgres.NewPostgresDB(dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)
	if err != nil {
		log.Fatal("cannot connect to database: ", err)
	}
	defer db.Close()

	userRepo := postgres.NewUserRepository(db)
	if err := userRepo.Migrate(); err != nil {
		log.Fatal("cannot run user migration: ", err)
	}

	commodityRepo := postgres.NewCommodityRepository(db)
	if err := commodityRepo.Migrate(); err != nil {
		log.Fatal("cannot run commodity migration: ", err)
	}

	correlationRepo := postgres.NewCorrelationRepository(db)
	if err := correlationRepo.Migrate(); err != nil {
		log.Fatal("cannot run correlation migration: ", err)
	}

	userService := application.NewUserService(userRepo)
	
	// Composite provider to route different commodites to different backends
	alphaClient := alphavantage.NewClient()
	yahooClient := yahoofinance.NewClient()
	
	provider := adapters.NewCompositeProvider(alphaClient)
	provider.Register("copper", yahooClient)
	provider.Register("aluminum", yahooClient)
	provider.Register("aluminium", yahooClient)
	
	commodityService := application.NewCommodityService(provider, commodityRepo)
	userHandler := http.NewUserHandler(userService)
	commodityHandler := http.NewCommodityHandler(commodityService)

	correlationService := application.NewCorrelationService(correlationRepo, commodityRepo)
	correlationHandler := http.NewCorrelationHandler(correlationService)

	r := chi.NewRouter()
	allowedOrigins := getAllowedOrigins()

	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	r.Use(c.Handler)

	r.Use(middleware.Logger)
	r.Use(authMiddleware.CSRFMiddleware)

	r.Get("/health", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.Write([]byte("OK"))
	})

	r.Route("/api", func(r chi.Router) {
		// Public routes
		r.Post("/register", userHandler.RegisterUserHandler)
		r.Post("/login", userHandler.LoginUserHandler)
		r.Post("/logout", userHandler.LogoutUserHandler)
		r.Post("/refresh", userHandler.RefreshJWTokenHandler)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.JWTAuthMiddleware)
			r.Get("/me", userHandler.MeHandler)
			r.Post("/user/change-password", userHandler.ChangePasswordHandler)
			r.Get("/commodity", commodityHandler.GetCommodityHandler)
			r.Get("/commodity/{name}/history", commodityHandler.GetCommodityHistoryHandler)
			r.Get("/commodity/status", commodityHandler.GetCommodityStatusHandler)
			r.Get("/correlation", correlationHandler.GetCorrelationHandler)
			r.Get("/correlation/history", correlationHandler.GetCorrelationHistoryHandler)
		})

		// Admin routes
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.AdminRoleMiddleware)
		})
	})

	runUpdateCycle := func() {
		log.Printf("Starting scheduled commodity refresh")

		if err := commodityService.UpdatePreciousPrices(); err != nil {
			log.Printf("Error updating precious commodities: %v", err)
		}

		if err := commodityService.UpdateIndustrialPrices(); err != nil {
			log.Printf("Error updating industrial commodities: %v", err)
		}

		pairs := [][]string{
			{"gold", "silver"},
			{"copper", "aluminum"},
			{"gold", "copper"},
			{"gold", "brent"},
			{"copper", "brent"},
		}
		for _, p := range pairs {
			if err := correlationService.UpdateCorrelations(p[0], p[1]); err != nil {
				log.Printf("Error updating correlation %s-%s: %v", p[0], p[1], err)
			}
		}

		log.Printf("Finished scheduled commodity refresh")
	}

	// Prime the database immediately on startup so the UI has data before the first ticker tick.
	runUpdateCycle()

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			runUpdateCycle()
		}
	}()

	log.Println("server starting on :8080")
	if err := stdhttp.ListenAndServe(fmt.Sprintf("%s:%s", os.Getenv("HOST"), os.Getenv("PORT")), r); err != nil {
		log.Fatal(err)
	}
}

func getEnv(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}

func getAllowedOrigins() []string {
	configured := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if configured == "" {
		return []string{
			"http://localhost:3100",
			"http://127.0.0.1:3100",
			"https://primetrading-nine.vercel.app",
			"http://129.151.247.163:3100", // Your Oracle Server (Frontend port)
			"http://129.151.247.163",      // Your Oracle Server (Public IP)
		}
	}

	parts := strings.Split(configured, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		origin := strings.TrimSpace(p)
		if origin != "" {
			origins = append(origins, origin)
		}
	}

	if len(origins) == 0 {
		return []string{"http://localhost:3100", "http://127.0.0.1:3100"}
	}

	return origins
}
