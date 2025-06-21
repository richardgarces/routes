package config

import (
	"context"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	// Validaciones de parámetros
	MaxKeyLength     = getEnvInt("MAX_KEY_LENGTH", 64)
	MaxTipoLength    = getEnvInt("MAX_TIPO_LENGTH", 32)
	MaxDestinoLength = getEnvInt("MAX_DESTINO_LENGTH", 256)
	MaxBodySize      = getEnvInt("MAX_BODY_SIZE", 1024) // bytes

	// Rate Limiting
	RateLimitRequests = getEnvInt("RATE_LIMIT_REQUESTS", 100)
	RateLimitWindow   = getEnvDuration("RATE_LIMIT_WINDOW", time.Minute)

	// Circuit Breaker
	CircuitBreakerMaxFailures = getEnvInt("CB_MAX_FAILURES", 5)
	CircuitBreakerOpenSeconds = getEnvInt("CB_OPEN_SECONDS", 30)

	// MongoDB
	MongoURI                    = getEnvStr("MONGO_URI", "mongodb://localhost:27017")
	MongoMaxPoolSize            = uint64(getEnvInt("MONGO_MAX_POOL_SIZE", 20))
	MongoConnectTimeout         = getEnvDuration("MONGO_CONNECT_TIMEOUT", 5*time.Second)
	MongoServerSelectionTimeout = getEnvDuration("MONGO_SERVER_SELECTION_TIMEOUT", 5*time.Second)

	// Servidor HTTP
	ServerPort         = getEnvStr("PORT", "8080")
	ServerReadTimeout  = getEnvDuration("SERVER_READ_TIMEOUT", 5*time.Second)
	ServerWriteTimeout = getEnvDuration("SERVER_WRITE_TIMEOUT", 10*time.Second)
	ServerIdleTimeout  = getEnvDuration("SERVER_IDLE_TIMEOUT", 30*time.Second)

	// Refresco de rutas
	RoutesRefreshSeconds = getEnvInt("ROUTES_REFRESH_SECONDS", 30)

	// Seguridad
	APIKey = getEnvStr("API_KEY", "MIApi1MIAMIApi12345pi123452MIApi12345345")
)

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getEnvStr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return time.Duration(n) * time.Second
		}
	}
	return def
}

// ConnectMongo establece la conexión con la base de datos MongoDB
func ConnectMongo() (*mongo.Client, error) {
	clientOpts := options.Client().
		ApplyURI(MongoURI).
		SetMaxPoolSize(MongoMaxPoolSize).
		SetConnectTimeout(MongoConnectTimeout).
		SetServerSelectionTimeout(MongoServerSelectionTimeout)

	client, err := mongo.Connect(context.TODO(), clientOpts)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func DisconnectMongo(client *mongo.Client) error {
	if client == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), MongoConnectTimeout)
	defer cancel()
	return client.Disconnect(ctx)
}

// IsValidKey verifica si la clave cumple con las restricciones de longitud
func IsValidKey(key string) bool {
	if len(key) == 0 || len(key) > MaxKeyLength {
		return false
	}
	return true
}

// IsValidTipo verifica si el tipo cumple con las restricciones de longitud
func IsValidTipo(tipo string) bool {
	if len(tipo) == 0 || len(tipo) > MaxTipoLength {
		return false
	}
	return true
}

// IsValidDestino verifica si el destino cumple con las restricciones de longitud
func IsValidDestino(destino string) bool {
	if len(destino) == 0 || len(destino) > MaxDestinoLength {
		return false
	}
	return true
}

// IsValidBodySize verifica si el tamaño del cuerpo de la solicitud es válido
func IsValidBodySize(size int64) bool {
	if size <= 0 || size > int64(MaxBodySize) {
		return false
	}
	return true
}

// IsValidAPIKey verifica si la clave API es válida
func IsValidAPIKey(apiKey string) bool {
	if len(apiKey) == 0 {
		return false
	}
	return apiKey == APIKey
}

// IsValidRateLimit verifica si la solicitud cumple con las restricciones de rate limiting
func IsValidRateLimit(requests int) bool {
	if requests <= 0 || requests > RateLimitRequests {
		return false
	}
	return true
}

// IsValidCircuitBreaker verifica si la solicitud cumple con las restricciones del circuito
func IsValidCircuitBreaker(failures int) bool {
	if failures < 0 || failures > CircuitBreakerMaxFailures {
		return false
	}
	return true
}

// IsValidMongoURI verifica si la URI de MongoDB es válida
func IsValidMongoURI(uri string) bool {
	if len(uri) == 0 {
		return false
	}
	// Aquí podrías agregar más validaciones específicas de la URI si es necesario
	return true
}

// IsValidMongoMaxPoolSize verifica si el tamaño máximo del pool de conexiones es válido
func IsValidMongoMaxPoolSize(size uint64) bool {
	if size == 0 || size > 1000 { // Definir un límite razonable
		return false
	}
	return true
}

// IsValidMongoConnectTimeout verifica si el tiempo de espera de conexión es válido
func IsValidMongoConnectTimeout(timeout time.Duration) bool {
	if timeout <= 0 || timeout > 30*time.Second { // Definir un límite razonable
		return false
	}
	return true
}

// IsValidMongoServerSelectionTimeout verifica si el tiempo de espera de selección de servidor es válido
func IsValidMongoServerSelectionTimeout(timeout time.Duration) bool {
	if timeout <= 0 || timeout > 30*time.Second { // Definir un límite razonable
		return false
	}
	return true
}

// IsValidServerPort verifica si el puerto del servidor es válido
func IsValidServerPort(port string) bool {
	if len(port) == 0 {
		return false
	}
	if _, err := strconv.Atoi(port); err != nil {
		return false
	}
	// Aquí podrías agregar más validaciones específicas del puerto si es necesario
	return true
}

// IsValidServerReadTimeout verifica si el tiempo de espera de lectura del servidor es válido
func IsValidServerReadTimeout(timeout time.Duration) bool {
	if timeout <= 0 || timeout > 30*time.Second { // Definir un límite razonable
		return false
	}
	return true
}

// IsValidServerWriteTimeout verifica si el tiempo de espera de escritura del servidor es válido
func IsValidServerWriteTimeout(timeout time.Duration) bool {
	if timeout <= 0 || timeout > 30*time.Second { // Definir un límite razonable
		return false
	}
	return true
}

// IsValidServerIdleTimeout verifica si el tiempo de espera inactivo del servidor es válido
func IsValidServerIdleTimeout(timeout time.Duration) bool {
	if timeout <= 0 || timeout > 30*time.Second { // Definir un límite razonable
		return false
	}
	return true
}

// IsValidRoutesRefreshSeconds verifica si el intervalo de refresco de rutas es válido
func IsValidRoutesRefreshSeconds(seconds int) bool {
	if seconds <= 0 || seconds > 3600 { // Definir un límite razonable (1 hora)
		return false
	}
	return true
}

// IsValidAPIKey verifica si la clave API es válida
