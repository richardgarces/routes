package router

import (
	"log"
	"sync"
)

type Service interface {
	GetBalancedRoute(key, tipo string) (string, error)
	AddDestino(key, tipo, destino string) error
	RefreshRoutes()
}

type service struct {
	repo      Repository
	rr        map[string]int
	mu        sync.Mutex
	routes    map[string][]string
	refreshMu sync.RWMutex
}

func NewService(repo Repository) *service {
	s := &service{
		repo:   repo,
		rr:     make(map[string]int),
		routes: make(map[string][]string),
	}
	s.RefreshRoutes()
	return s
}

func routeMapKey(key, tipo string) string {
	return key + "|" + tipo
}

func (s *service) RefreshRoutes() {
	log.Println("Refrescando rutas desde la base de datos...")
	routes, err := s.repo.GetAllRoutes()
	if err != nil {
		log.Printf("Error al refrescar rutas: %v", err)
		return
	}
	s.refreshMu.Lock()
	defer s.refreshMu.Unlock()
	s.routes = make(map[string][]string)
	for _, route := range routes {
		s.routes[routeMapKey(route.Key, route.Tipo)] = route.Destinos
	}
	log.Printf("Rutas cargadas en memoria: %d", len(s.routes))
}

func (s *service) GetBalancedRoute(key, tipo string) (string, error) {
	mapKey := routeMapKey(key, tipo)
	s.refreshMu.RLock()
	destinos, ok := s.routes[mapKey]
	s.refreshMu.RUnlock()
	log.Printf("[Service] Buscando en memoria mapKey='%s'. Encontrado: %v, destinos: %#v", mapKey, ok, destinos)
	if !ok || len(destinos) == 0 {
		log.Printf("[Service] No se encontró la ruta en memoria para key='%s', tipo='%s'. Consultando MongoDB...", key, tipo)
		route, err := s.repo.GetRoute(key, tipo)
		if err != nil {
			log.Printf("[Service] Error consultando MongoDB: %v", err)
			return "", err
		}
		if len(route.Destinos) == 0 {
			log.Printf("[Service] Documento encontrado pero sin destinos: %+v", route)
			return "", nil
		}
		destinos = route.Destinos
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	idx := s.rr[mapKey] % len(destinos)
	s.rr[mapKey] = (s.rr[mapKey] + 1) % len(destinos)
	log.Printf("[Service] Retornando destino: %s para key='%s', tipo='%s'", destinos[idx], key, tipo)
	return destinos[idx], nil
}

func (s *service) AddDestino(key, tipo, destino string) error {
	log.Printf("Agregando destino %s a la key %s, tipo %s", destino, key, tipo)
	err := s.repo.SaveRoute(key, tipo, destino)
	if err != nil {
		log.Printf("Error agregando destino %s a la key %s, tipo %s: %v", destino, key, tipo, err)
		return err
	}
	s.refreshMu.Lock()
	mapKey := routeMapKey(key, tipo)
	s.routes[mapKey] = append(s.routes[mapKey], destino)
	s.refreshMu.Unlock()
	return nil
}
