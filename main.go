package main

import (
	"log"
	"net/http"
	"router-app/config"
	"router-app/router"
	"time"
)

func main() {
	port := config.ServerPort

	log.Printf("Intervalo de refresco de rutas: %d segundos", config.RoutesRefreshSeconds)

	db, err := config.ConnectMongo()
	if err != nil {
		log.Fatalf("Error al conectar a MongoDB: %v", err)
	}
	defer func() {
		if cerr := config.DisconnectMongo(db); cerr != nil {
			log.Printf("Error al desconectar MongoDB: %v", cerr)
		}
	}()

	database := db.Database("routingdb")
	repo := router.NewRepository(database)
	svc := router.NewService(repo)
	h := router.NewHandler(svc)

	// Refrescar rutas periódicamente en background usando el valor de config
	go func() {
		ticker := time.NewTicker(time.Duration(config.RoutesRefreshSeconds) * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			svc.RefreshRoutes()
		}
	}()

	// Inicializa el rate limiter usando los parámetros de config.go
	rl := router.NewRateLimiter(config.RateLimitRequests, config.RateLimitWindow)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Aplica el middleware de rate limiting
	handler := router.RateLimitMiddleware(rl, mux)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  config.ServerReadTimeout,
		WriteTimeout: config.ServerWriteTimeout,
		IdleTimeout:  config.ServerIdleTimeout,
	}

	log.Printf("Servidor corriendo en :%s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}

// DatabaseName is the name of the MongoDB database to use
const DatabaseName = "routerdb"
