# Data Ingestion Microservice - Rust Implementation (Improved)

Esta es la implementación mejorada en Rust del microservicio de ingesta de datos para el Sistema Distribuido de Seguimiento de Rutas GPS, siguiendo las mejores prácticas de Rust.

## 🚀 Características Principales

- **Arquitectura Modular**: Código organizado en módulos especializados
- **Manejo de Errores Robusto**: Tipos de error personalizados con `thiserror`
- **Configuración Flexible**: Variables de entorno con valores por defecto
- **Métricas Integradas**: Monitoreo de rendimiento en tiempo real
- **Health Checks**: Verificaciones automáticas de salud del servicio
- **Graceful Shutdown**: Cierre ordenado del servicio
- **Tests Comprehensivos**: Tests unitarios, de integración y benchmarks
- **Concurrencia Optimizada**: Procesamiento paralelo de mensajes MQTT

## 🏗️ Estructura del Proyecto

```
data_ingestion_microservice/
├── src/
│   ├── main.rs                 # Punto de entrada principal
│   ├── config.rs              # Configuración y variables de entorno
│   ├── types.rs               # Tipos de datos y errores personalizados
│   ├── route_simplification.rs # Algoritmos de simplificación de rutas
│   └── service.rs             # Servicio principal
├── Cargo.toml                 # Dependencias y configuración del proyecto
├── Makefile                   # Comandos de desarrollo y construcción
├── env.example               # Variables de entorno de ejemplo
└── README.md                 # Esta documentación
```

## 📦 Dependencias Principales

- **tokio**: Runtime asíncrono para concurrencia
- **rumqttc**: Cliente MQTT con soporte TLS
- **redis**: Cliente Redis asíncrono
- **mongodb**: Driver oficial de MongoDB
- **geo**: Algoritmos geoespaciales
- **serde**: Serialización/deserialización
- **thiserror**: Manejo de errores ergonómico
- **log + pretty_env_logger**: Sistema de logging

## 🛠️ Instalación y Configuración

### Prerrequisitos

1. **Rust** (versión 1.70 o superior):

```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source ~/.cargo/env
```

2. **Servicios de infraestructura**:

```bash
# Desde el directorio raíz del proyecto
docker-compose up -d
```

### Configuración de Variables de Entorno

Copia el archivo de ejemplo y personalízalo:

```bash
cp env.example .env
```

Variables disponibles:

- `RUST_LOG`: Nivel de logging (debug, info, warn, error)
- `MQTT_BROKER`: Dirección del broker MQTT
- `MQTT_PORT`: Puerto del broker MQTT
- `REDIS_URL`: URL de conexión a Redis
- `MONGODB_URI`: URI de conexión a MongoDB
- `ROUTE_TOLERANCE`: Tolerancia para simplificación de rutas

## 🔧 Comandos de Desarrollo

### Configuración inicial

```bash
# Instalar herramientas de desarrollo
make install-tools

# Configurar entorno de desarrollo
make dev-setup
```

### Desarrollo

```bash
# Compilar en modo debug
make build

# Ejecutar la aplicación
make run

# Ejecutar con recarga automática
make watch

# Formatear código
make format

# Ejecutar lints
make lint

# Ejecutar tests
make test

# Ejecutar tests con salida detallada
make test-verbose
```

### Calidad de Código

```bash
# Ejecutar todas las verificaciones
make check-all

# Auditar dependencias por vulnerabilidades
make audit

# Verificar dependencias desactualizadas
make outdated

# Generar documentación
make doc
```

### Producción

```bash
# Compilar para producción
make build-release

# Ejecutar benchmarks
make bench

# Crear paquete de distribución
make package
```

### Docker

```bash
# Construir imagen Docker
make docker-build

# Ejecutar en contenedor
make docker-run
```

## 📊 Métricas y Monitoreo

El servicio incluye métricas integradas que se reportan automáticamente:

- **Mensajes procesados**: Total de mensajes MQTT recibidos
- **Rutas en progreso**: Rutas actualmente siendo rastreadas
- **Rutas completadas**: Total de rutas finalizadas y guardadas
- **Puntos procesados/simplificados**: Estadísticas de compresión
- **Ratio de compresión**: Eficiencia del algoritmo de simplificación
- **Errores**: Contador de errores del servicio

## 🔍 Health Checks

El servicio incluye verificaciones automáticas de salud:

- **Tasa de errores**: Alerta si supera el 10%
- **Rutas pendientes**: Alerta si hay más de 100 rutas en progreso
- **Uso de memoria**: Monitoreo básico de memoria (Linux)

## 🌐 API de Mensajes MQTT

### Mensaje "in_route"

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

### Mensaje "finished"

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

## 🧮 Algoritmo de Simplificación

Implementa el algoritmo **Ramer-Douglas-Peucker** con las siguientes características:

- **Tolerancia configurable**: Ajustable vía variable de entorno
- **Dos implementaciones**: Usando la librería `geo` y implementación personalizada
- **Métricas detalladas**: Estadísticas de compresión y rendimiento
- **Validación**: Verificación de entrada y manejo de casos edge

## 🔧 Configuración Avanzada

### Features de Cargo

- `full` (default): Incluye todas las características
- `metrics`: Habilita reportes de métricas
- `health-checks`: Habilita verificaciones de salud

### Perfiles de Compilación

- **Development**: Optimizado para velocidad de compilación
- **Release**: Optimizado para rendimiento con LTO
- **Test**: Balance entre velocidad y debugging

## 🧪 Testing

```bash
# Tests unitarios
make test

# Tests de integración (requiere servicios)
make test-integration

# Benchmarks de rendimiento
make bench

# Tests con coverage
cargo tarpaulin --all-features
```

## 📈 Benchmarking y Profiling

```bash
# Ejecutar benchmarks
make bench

# Profiling con perf (Linux)
make profile

# Generar flamegraph
make flamegraph
```

## 🚀 Deployment

### Compilación para Producción

```bash
make build-release
```

### Contenedor Docker

```bash
make docker-build
make docker-run
```

### Variables de Entorno para Producción

```bash
RUST_LOG=info
MQTT_BROKER=production-mqtt.example.com
REDIS_URL=redis://production-redis.example.com:6379
MONGODB_URI=mongodb://user:pass@production-mongo.example.com:27017
```

## 🔒 Seguridad

- **TLS/SSL**: Soporte para conexiones seguras MQTT
- **Validación de entrada**: Sanitización de datos MQTT
- **Auditoría de dependencias**: Verificación automática de vulnerabilidades
- **Logs seguros**: Sin exposición de datos sensibles

## 🐛 Troubleshooting

### Problemas Comunes

1. **Error de conexión MQTT**:

   - Verificar que el broker esté ejecutándose
   - Comprobar configuración de red y firewall

2. **Error de conexión Redis**:

   - Verificar servicio Redis: `docker ps`
   - Comprobar URL de conexión

3. **Error de conexión MongoDB**:
   - Verificar credenciales en URI
   - Comprobar que la base de datos esté accesible

### Logs de Debug

```bash
RUST_LOG=debug make run
```

## 🤝 Contribución

1. Fork el proyecto
2. Crea una rama para tu feature (`git checkout -b feature/amazing-feature`)
3. Ejecuta las verificaciones: `make check-all`
4. Commit tus cambios (`git commit -m 'Add amazing feature'`)
5. Push a la rama (`git push origin feature/amazing-feature`)
6. Abre un Pull Request

## 📄 Licencia

Este proyecto está bajo la licencia MIT. Ver `LICENSE` para más detalles.

## 📊 Comparación con Go

### Ventajas de esta implementación Rust:

- **Seguridad de memoria**: Sin garbage collection, cero-cost abstractions
- **Rendimiento**: Compilación nativa optimizada
- **Type safety**: Sistema de tipos avanzado previene muchos errores
- **Concurrencia segura**: Ownership model previene data races
- **Ecosistema robusto**: Crates.io con librerías de alta calidad

### Métricas esperadas vs Go:

- **Memoria**: ~30-50% menos uso de memoria
- **CPU**: ~10-20% mejor rendimiento
- **Latencia**: Menor latencia en percentil 99
- **Throughput**: Mayor capacidad de procesamiento de mensajes
