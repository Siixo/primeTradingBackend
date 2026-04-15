package main

import (
	"context"
	"fmt"
	"log"
	stdhttp "net/http"
	"os/signal"
	"syscall"
	"time"

	authMiddleware "backend/internal/middleware"

	"backend/internal/adapters/alphavantage"
	"backend/internal/adapters/postgres"
	"backend/internal/application"
	"backend/internal/config"
	http "backend/internal/handler"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("config: ", err)
	}

	db, err := postgres.NewPostgresDB(cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.SSLMode)
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

	// Graceful shutdown context
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Seed POC data
	application.RunCommoditySeeder(ctx, commodityRepo)

	// Services
	httpClient := &stdhttp.Client{Timeout: 30 * time.Second}
	alphaClient := alphavantage.NewClient(httpClient, cfg.Alpha.AlphaVantageKey, cfg.Alpha.GoldPricezKey)

	userService := application.NewUserService(userRepo, cfg.JWT.SigningKey)
	commodityService := application.NewCommodityService(alphaClient, commodityRepo)
	correlationService := application.NewCorrelationService(correlationRepo, commodityRepo)

	// Handlers
	userHandler := http.NewUserHandler(userService, cfg.Server.CookieSecure)
	commodityHandler := http.NewCommodityHandler(commodityService)
	correlationHandler := http.NewCorrelationHandler(correlationService)

	// Router
	r := chi.NewRouter()

	c := cors.New(cors.Options{
		AllowedOrigins:   cfg.Server.CORSOrigins,
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
			r.Use(authMiddleware.NewJWTAuthMiddleware(cfg.JWT.SigningKey))
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
			r.Use(authMiddleware.NewJWTAuthMiddleware(cfg.JWT.SigningKey))
			r.Use(authMiddleware.AdminRoleMiddleware)
		})
	})

	// Background commodity refresh
	runUpdateCycle := func() {
		log.Printf("Starting scheduled commodity refresh")

		if err := commodityService.UpdatePreciousPrices(ctx); err != nil {
			log.Printf("Error updating precious commodities: %v", err)
		}

		if err := commodityService.UpdateIndustrialPrices(ctx); err != nil {
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
			if err := correlationService.UpdateCorrelations(ctx, p[0], p[1]); err != nil {
				log.Printf("Error updating correlation %s-%s: %v", p[0], p[1], err)
			}
		}

		log.Printf("Finished scheduled commodity refresh")
	}

	runUpdateCycle()

	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				runUpdateCycle()
			}
		}
	}()

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	srv := &stdhttp.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		log.Printf("server starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != stdhttp.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("server shutdown error: ", err)
	}
	log.Println("server stopped")
}
