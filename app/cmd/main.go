package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"br.tec.suportes/portal/internal/config"
	apphttp "br.tec.suportes/portal/internal/http"
	"br.tec.suportes/portal/internal/repository"
	"br.tec.suportes/portal/internal/service"
	"br.tec.suportes/portal/internal/store"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	cfg := config.Load()

	if err := os.MkdirAll(cfg.Database.Path, 0755); err != nil {
		log.Fatalf("[gateway] falha ao criar diretório do banco: %v", err)
	}

	db, err := store.Open(cfg.Database.FilePath(), cfg.Database.Timeout())
	if err != nil {
		log.Fatalf("[gateway] falha ao abrir store: %v", err)
	}
	defer db.Close()

	dbRepo := repository.NewDatabaseRepo(db)
	cmdRepo := repository.NewCommandRepo(db)
	routeRepo := repository.NewRouteRepo(db)
	certRepo := repository.NewCertificateRepo(db)

	dbSvc := service.NewDatabaseService(dbRepo)
	cmdSvc := service.NewCommandService(cmdRepo)
	certSvc := service.NewCertificateService(certRepo)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(cfg.Timeout()))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders: []string{"*"},
		MaxAge:         300,
	}))

	routeSvc := service.NewRouteService(routeRepo, r, dbSvc, cmdSvc, certSvc)

	apphttp.NewDatabaseHandler(dbSvc).RegisterRoutes(r)
	apphttp.NewCommandHandler(cmdSvc).RegisterRoutes(r)
	apphttp.NewRouteHandler(routeSvc).RegisterRoutes(r)
	apphttp.NewCertificateHandler(certSvc).RegisterRoutes(r)
	apphttp.NewStatusHandler(dbSvc, cfg.Status).RegisterRoutes(r)

	srv := &http.Server{
		Addr:         cfg.Addr(),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: cfg.Timeout() + 5*time.Second,
		IdleTimeout:  60 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("[gateway] escutando em %s", cfg.Addr())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[gateway] servidor falhou: %v", err)
		}
	}()

	<-quit
	log.Println("[gateway] encerrando servidor...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[gateway] shutdown forçado: %v", err)
	}

	log.Println("[gateway] servidor encerrado")
}
