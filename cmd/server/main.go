package main

import (
	authMiddleware "backend/internal/middleware"
	"fmt"
	"log"
	stdhttp "net/http"
	"os"

	"backend/internal/adapters/postgres"
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

	db, err := postgres.NewPostgresDB(dbHost, dbPort, dbUser, dbPassword, dbName)
	if err != nil {
		log.Fatal("cannot connect to database: ", err)
	}
	defer db.Close()

	userRepo := postgres.NewUserRepository(db)
	_ = userRepo.Migrate()

	userService := application.NewUserService(userRepo)
	stockService := application.NewStockService()
	userHandler := http.NewUserHandler(userService)

	r := chi.NewRouter()

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3100", "http://127.0.0.1:3100", "https://primetrading-nine.vercel.app"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	r.Use(c.Handler)

	r.Use(middleware.Logger)

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
			r.Get("/stock", stockService.GetStocks)
		})

		// Admin routes
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.AdminRoleMiddleware)
		})
	})

	log.Println("server starting on :8080")
	if err := stdhttp.ListenAndServe(fmt.Sprintf("%s:%s", os.Getenv("HOST"), os.Getenv("PORT")), r); err != nil {
		log.Fatal(err)
	}
}
