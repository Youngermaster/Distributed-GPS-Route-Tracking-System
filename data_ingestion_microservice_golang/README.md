# Data Ingestion Microservice - Go Implementation

Esta es la implementación en Go del microservicio de ingesta de datos para el Sistema Distribuido de Seguimiento de Rutas GPS.

## Funcionalidades

- **Conexión MQTT**: Se suscribe al tópico `drivers_location/#` para recibir datos de ubicación de los conductores
- **Almacenamiento temporal en Redis**: Guarda las ubicaciones en tiempo real durante el trayecto
- **Simplificación de rutas**: Implementa el algoritmo Ramer-Douglas-Peucker para optimizar las rutas
- **Persistencia en MongoDB**: Almacena las rutas simplificadas una vez completadas
- **Manejo de errores robusto**: Incluye logging detallado y manejo graceful de desconexiones
- **Concurrencia**: Procesa mensajes MQTT de forma concurrente usando gorrutinas

## Prerrequisitos

### Instalar Go

#### macOS (usando Homebrew)

```bash
brew install go
```

#### Ubuntu/Debian

```bash
sudo apt update
sudo apt install golang-go
```

#### Windows

Descarga e instala desde [https://golang.org/dl/](https://golang.org/dl/)

### Verificar instalación

```bash
go version
```

### Servicios requeridos

Asegúrate de que los siguientes servicios estén ejecutándose:

- **MQTT Broker** (EMQX): Puerto 1883
- **Redis**: Puerto 6379
- **MongoDB**: Puerto 27017

Puedes usar Docker Compose desde el directorio raíz del proyecto:

```bash
docker-compose up -d
```

## Instalación y Configuración

1. **Clonar y navegar al directorio**:

```bash
cd data_ingestion_microservice_golang
```

2. **Instalar dependencias**:

```bash
make install-deps
```

O manualmente:

```bash
go mod download
go mod tidy
```

## Uso

### Ejecutar el microservicio

```bash
make run
```

O directamente:

```bash
go run main.go
```

### Compilar para producción

```bash
make build-prod
```

### Ejecutar todas las verificaciones

```bash
make all
```

## Estructura del Proyecto

```
data_ingestion_microservice_golang/
├── main.go           # Archivo principal con toda la lógica
├── go.mod           # Definición del módulo y dependencias
├── go.sum           # Checksums de las dependencias
├── Makefile         # Comandos de construcción y desarrollo
└── README.md        # Esta documentación
```

## Mensajes MQTT Esperados

### Formato de mensaje "in_route"

```json
{
  "driverId": "driver_123",
  "driverLocation": {
    "latitude": 40.7128,
    "longitude": -74.006
  },
  "timestamp": 1634567890,
  "currentRouteId": "route_456",
  "status": "in_route"
}
```

### Formato de mensaje "finished"

```json
{
  "driverId": "driver_123",
  "driverLocation": {
    "latitude": 40.7829,
    "longitude": -73.9654
  },
  "timestamp": 1634571490,
  "currentRouteId": "route_456",
  "status": "finished"
}
```

## Configuración

El microservicio utiliza las siguientes configuraciones por defecto:

- **MQTT Broker**: `localhost:1883`
- **Redis**: `127.0.0.1:6379`
- **MongoDB**: `mongodb://root:examplepassword@127.0.0.1:27017`
- **Base de datos**: `distributed_gps_route_tracking_system`
- **Colección**: `trips`

## Algoritmo de Simplificación

Implementa el algoritmo **Ramer-Douglas-Peucker** para simplificar rutas GPS:

- **Tolerancia**: 0.0001 (aproximadamente 11 metros)
- **Propósito**: Reducir el número de puntos GPS manteniendo la forma general de la ruta
- **Beneficio**: Optimiza el almacenamiento y mejora el rendimiento de consultas

## Dependencias

- `github.com/eclipse/paho.mqtt.golang`: Cliente MQTT
- `github.com/redis/go-redis/v9`: Cliente Redis
- `go.mongodb.org/mongo-driver`: Driver oficial de MongoDB
- Bibliotecas estándar de Go para JSON, logging, context, etc.

## Comandos de Desarrollo

```bash
# Formatear código
make format

# Verificar estilo (requiere golangci-lint)
make lint

# Verificar compilación
make check

# Ejecutar tests
make test

# Compilar aplicación
make build

# Limpiar artefactos
make clean
```

## Comparación con la Versión Rust

### Ventajas de Go:

- **Simplicidad**: Sintaxis más sencilla y familiar
- **Concurrencia nativa**: Gorrutinas y channels integrados
- **Ecosistema maduro**: Gran cantidad de librerías bien mantenidas
- **Tooling excelente**: `go fmt`, `go mod`, `go test` integrados
- **Compilación rápida**: Tiempos de compilación más cortos

### Diferencias de implementación:

- **Manejo de errores**: Go utiliza el patrón `if err != nil`
- **Concurrencia**: Gorrutinas en lugar de async/await
- **Gestión de memoria**: Garbage collection automático
- **Tipos**: Sistema de tipos más simple que Rust

## Logging

El microservicio registra:

- Conexiones y desconexiones de servicios
- Procesamiento de mensajes MQTT
- Almacenamiento en Redis y MongoDB
- Errores y excepciones
- Estadísticas de simplificación de rutas

## Manejo de Errores

- **Reconexión automática**: MQTT y bases de datos
- **Logging detallado**: Para debugging y monitoreo
- **Graceful shutdown**: Limpieza ordenada de recursos
- **Validación de datos**: Verificación de formato JSON

## Contribución

1. Fork el proyecto
2. Crea una rama para tu feature (`git checkout -b feature/AmazingFeature`)
3. Commit tus cambios (`git commit -m 'Add some AmazingFeature'`)
4. Push a la rama (`git push origin feature/AmazingFeature`)
5. Abre un Pull Request
