package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/psds-microservice/user-service/internal/auth"
	"github.com/psds-microservice/user-service/internal/config"
	"github.com/psds-microservice/user-service/internal/database"
	grpcserver "github.com/psds-microservice/user-service/internal/grpc"
	"github.com/psds-microservice/user-service/internal/handler"
	"github.com/psds-microservice/user-service/internal/repository"
	"github.com/psds-microservice/user-service/internal/service"
	"github.com/psds-microservice/user-service/pkg/gen/user_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// load .env from current dir or project root (when run via make from bin/)
	if err := godotenv.Load(".env"); err != nil {
		_ = godotenv.Load("../.env")
	}
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// Subcommand: migrate up (e.g. make migrate)
	if len(os.Args) >= 3 && os.Args[1] == "migrate" && os.Args[2] == "up" {
		if err := database.MigrateUp(cfg.DatabaseURL()); err != nil {
			log.Fatalf("migrate: %v", err)
		}
		return
	}

	// Subcommand: seed (e.g. make seed)
	if len(os.Args) >= 2 && os.Args[1] == "seed" {
		if err := database.MigrateUp(cfg.DatabaseURL()); err != nil {
			log.Fatalf("migrate: %v", err)
		}
		db, err := database.Open(cfg.DSN())
		if err != nil {
			log.Fatalf("db: %v", err)
		}
		if err := database.RunSeeds(db); err != nil {
			log.Fatalf("seed: %v", err)
		}
		return
	}

	// Run migrations on app startup
	if err := database.MigrateUp(cfg.DatabaseURL()); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	db, err := database.Open(cfg.DSN())
	if err != nil {
		log.Fatalf("db: %v", err)
	}

	// Init Layers
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewUserSessionRepository(db)
	userSvc := service.NewUserService(userRepo, sessionRepo)
	userHandler := handler.NewUserHandler(userSvc)

	jwtCfg, err := auth.NewConfig(cfg.JWTSecret, cfg.JWTAccess, cfg.JWTRefresh)
	if err != nil {
		log.Printf("jwt config: %v, using defaults", err)
	}
	blacklist := auth.NewBlacklist()

	// HTTP server (health + REST API)
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handler.Health)
	mux.HandleFunc("/ready", handler.Ready)
	// Legacy routes (backward compat)
	mux.HandleFunc("/users/create", userHandler.CreateUser)
	mux.HandleFunc("/users/", userHandler.GetUser)
	mux.HandleFunc("/login", userHandler.Login)
	// API v1 (promt.txt): auth, users/me, operators, sessions
	mux.Handle("/api/v1/", handler.APIv1(userSvc, jwtCfg, blacklist, userHandler))

	httpAddr := cfg.AppHost + ":" + cfg.HTTPPort
	httpSrv := &http.Server{Addr: httpAddr, Handler: mux}
	go func() {
		log.Printf("HTTP server listening on %s", httpAddr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("http: %v", err)
		}
	}()

	// gRPC server
	grpcAddr := cfg.AppHost + ":" + cfg.GRPCPort
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("grpc listen: %v", err)
	}
	grpcSrv := grpc.NewServer()
	user_service.RegisterUserServiceServer(grpcSrv, grpcserver.NewServer(userSvc))
	reflection.Register(grpcSrv)
	go func() {
		log.Printf("gRPC server listening on %s", grpcAddr)
		if err := grpcSrv.Serve(lis); err != nil {
			log.Printf("grpc: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	if err := httpSrv.Shutdown(context.Background()); err != nil {
		log.Printf("http shutdown: %v", err)
	}
	grpcSrv.GracefulStop()
	fmt.Println("bye")
}
