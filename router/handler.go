package router

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"router-app/config"
)

var (
	validKey  = regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)
	validTipo = regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)
	validURL  = regexp.MustCompile(`^https?://[^\s]+$`)
)

func validateParam(param string, maxLen int, re *regexp.Regexp) bool {
	return len(param) > 0 && len(param) <= maxLen && re.MatchString(param)
}

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	log.Println("Registering routes...")
	mux.HandleFunc("/route/", h.RouteRequest)
	mux.HandleFunc("/add-destino/", h.AddDestino)
}

func (h *Handler) RouteRequest(w http.ResponseWriter, r *http.Request) {
	log.Println("[RouteRequest] Solicitud recibida:", r.Method, r.URL.Path)
	if r.Method != http.MethodGet {
		log.Printf("[RouteRequest] Método inválido: %s\n", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/route/"), "/")
	log.Printf("[RouteRequest] URL parts extraídos: %#v", parts)
	if len(parts) != 2 {
		log.Printf("[RouteRequest] Formato de ruta inválido: %s", r.URL.Path)
		http.Error(w, "Formato de ruta inválido. Usa /route/{tipo}/{key}", http.StatusBadRequest)
		return
	}
	tipo, key := parts[0], parts[1]
	log.Printf("[RouteRequest] tipo='%s', key='%s'", tipo, key)

	if !validateParam(tipo, config.MaxTipoLength, validTipo) {
		log.Printf("[RouteRequest] Validación fallida para tipo: '%s'", tipo)
		http.Error(w, "Parámetro 'tipo' inválido", http.StatusBadRequest)
		return
	}
	if !validateParam(key, config.MaxKeyLength, validKey) {
		log.Printf("[RouteRequest] Validación fallida para key: '%s'", key)
		http.Error(w, "Parámetro 'key' inválido", http.StatusBadRequest)
		return
	}

	log.Printf("[RouteRequest] Buscando destino para tipo='%s', key='%s'", tipo, key)
	destino, err := h.svc.GetBalancedRoute(key, tipo)
	if err != nil {
		log.Printf("[RouteRequest] Error al obtener destino: %v", err)
		http.Error(w, "No route found", http.StatusNotFound)
		return
	}
	if destino == "" {
		log.Printf("[RouteRequest] No se encontró destino para tipo='%s', key='%s'", tipo, key)
		http.Error(w, "No route found", http.StatusNotFound)
		return
	}

	log.Printf("[RouteRequest] Destino encontrado: %s", destino)
	response := map[string]string{"destino": destino}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) AddDestino(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request for AddDestino")
	if r.Method != http.MethodPost {
		log.Printf("Invalid method: %s\n", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Espera: /add-destino/{tipo}/{key}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/add-destino/"), "/")
	if len(parts) != 2 {
		http.Error(w, "Formato de ruta inválido. Usa /add-destino/{tipo}/{key}", http.StatusBadRequest)
		return
	}
	tipo, key := parts[0], parts[1]

	// Usa los valores de config.go
	if !validateParam(tipo, config.MaxTipoLength, validTipo) || !validateParam(key, config.MaxKeyLength, validKey) {
		http.Error(w, "Parámetros inválidos", http.StatusBadRequest)
		return
	}

	// Limitar el tamaño del cuerpo de la solicitud usando config.MaxBodySize
	r.Body = http.MaxBytesReader(w, r.Body, int64(config.MaxBodySize))
	var req struct {
		Destino string `json:"destino"`
	}
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		if err == io.EOF {
			http.Error(w, "Cuerpo vacío", http.StatusBadRequest)
		} else {
			http.Error(w, "Payload inválido", http.StatusBadRequest)
		}
		return
	}

	if !validateParam(req.Destino, config.MaxDestinoLength, validURL) {
		http.Error(w, "Destino inválido", http.StatusBadRequest)
		return
	}

	log.Printf("Decoded destino: %s\n", req.Destino)
	err := h.svc.AddDestino(key, tipo, req.Destino)
	if err != nil {
		log.Printf("Error saving destino for key %s, tipo %s: %v\n", key, tipo, err)
		http.Error(w, "Could not save", http.StatusInternalServerError)
		return
	}

	log.Printf("Destino added successfully for key %s, tipo %s\n", key, tipo)
	response := map[string]string{"status": "added"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

var cb = NewCircuitBreaker(
	config.CircuitBreakerMaxFailures,
	time.Duration(config.CircuitBreakerOpenSeconds)*time.Second,
)

func llamadaAUnServicioInterno() error {
	// Simulación de llamada a un servicio interno
	return nil
}

func (h *Handler) SomeInternalCallHandler(w http.ResponseWriter, r *http.Request) {
	if !cb.Allow() {
		http.Error(w, "Servicio temporalmente no disponible", http.StatusServiceUnavailable)
		return
	}
	err := llamadaAUnServicioInterno()
	if err != nil {
		cb.Failure()
		log.Printf("Error en llamada a servicio interno: %v\n", err)
		http.Error(w, "Error interno", http.StatusInternalServerError)
		return
	}
	cb.Success()
	w.WriteHeader(http.StatusNoContent)
}
