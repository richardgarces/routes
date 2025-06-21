# Variables
IMAGE_NAME = router-app
TAG = latest
DOCKER_USERNAME = rgarces # Reemplaza con tu nombre de usuario en Docker Hub
PORT = 8080

# Construir la imagen de Docker
build:
	docker build -t $(IMAGE_NAME):$(TAG) .

# Ejecutar la aplicación localmente
run:
	go run main.go

# Ejecutar la aplicación con Docker
run-docker:
	docker run --rm -p $(PORT):$(PORT) --env-file .env $(IMAGE_NAME):$(TAG)

# Subir la imagen a Docker Hub
push:
	docker tag $(IMAGE_NAME):$(TAG) $(DOCKER_USERNAME)/$(IMAGE_NAME):$(TAG)
	docker push $(DOCKER_USERNAME)/$(IMAGE_NAME):$(TAG)

# Limpiar imágenes y contenedores no utilizados
clean:
	docker system prune -f

# Ayuda
help:
	@echo "Comandos disponibles:"
	@echo "  build       - Construir la imagen de Docker"
	@echo "  run         - Ejecutar la aplicación localmente con Go"
	@echo "  run-docker  - Ejecutar la aplicación con Docker"
	@echo "  push        - Subir la imagen a Docker Hub"
	@echo "  clean       - Limpiar imágenes y contenedores no utilizados"