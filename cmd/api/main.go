package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"event-booking/internal/config"
	"event-booking/internal/database"
	"event-booking/internal/handlers"
	"event-booking/internal/jobs"
	"event-booking/internal/repositories"
	"event-booking/internal/router"
	"event-booking/internal/services"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}

	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	queueClient := jobs.NewAsynqQueue(cfg.RedisAddress)
	defer queueClient.Close()

	userRepo := repositories.NewUserRepository(db)
	eventRepo := repositories.NewEventRepository(db)
	bookingRepo := repositories.NewBookingRepository(db)

	authService := services.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiry)
	eventService := services.NewEventService(db, eventRepo, bookingRepo, queueClient)
	bookingService := services.NewBookingService(db, eventRepo, bookingRepo, queueClient)

	authHandler := handlers.NewAuthHandler(authService)
	eventHandler := handlers.NewEventHandler(eventService)
	bookingHandler := handlers.NewBookingHandler(bookingService)
	healthHandler := handlers.NewHealthHandler()

	deps := router.Dependencies{
		AuthHandler:    authHandler,
		EventHandler:   eventHandler,
		BookingHandler: bookingHandler,
		HealthHandler:  healthHandler,
		AuthService:    authService,
	}

	mode := cfg.AppMode
	switch mode {
	case "", "api":
		runAPI(cfg.HTTPPort, deps)
	case "worker":
		runWorker(cfg.RedisAddress)
	default:
		log.Fatalf("unsupported APP_MODE %q", mode)
	}
}

func runAPI(port string, deps router.Dependencies) {
	engine := router.New(deps)
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("api listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("start api server: %v", err)
		}
	}()

	waitForShutdown(func(ctx context.Context) error {
		return srv.Shutdown(ctx)
	})
}

func runWorker(redisAddress string) {
	server := jobs.NewServer(redisAddress)

	go func() {
		log.Printf("worker listening on redis %s", redisAddress)
		if err := server.Run(jobs.NewMux()); err != nil {
			log.Fatalf("start worker: %v", err)
		}
	}()

	waitForShutdown(func(_ context.Context) error {
		server.Shutdown()
		return nil
	})
}

func waitForShutdown(stop func(context.Context) error) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := stop(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
