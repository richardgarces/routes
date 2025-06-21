package router

import (
	"log"
	"net/http"
	"router-app/config"
)

func AuthMiddleware(next http.Handler) http.Handler {
	requiredKey := config.APIKey // ...en vez de leer API_KEY directamente...
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" || apiKey != requiredKey {
			log.Printf("Intento fallido de autenticación desde IP %s, User-Agent: %s", r.RemoteAddr, r.UserAgent())
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
