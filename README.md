# CRUD con Property-Based Testing en Go

Sistema CRUD de usuarios con pruebas basadas en propiedades usando `pgregory.net/rapid`.

**Coverage: 97.8%** ✅

---

## 📋 Requisitos

- **Go 1.21+**
- **Git** (opcional, para clonar)

---

## 🚀 Instalación y Configuración

### 1. Clonar el Repositorio

```bash
git clone https://github.com/Claudio-Espinoza/CRUD__Property-based-Testing.git
cd CRUD__Property-based-Testing
```

### 2. Instalar Dependencias

```bash
go mod download
```

### 3. Verificar Instalación

```bash
# Compilar el proyecto
go build -o bin/app cmd/main.go

# Ejecutar el ejemplo
./bin/app  # Linux/Mac
.\bin\app.exe  # Windows
```

**Salida esperada:**
```
Created user: &{ID:uuid-123 Name:John Doe Email:john@example.com Age:30 ...}
Created user: &{ID:uuid-456 Name:Jane Smith Email:jane@example.com Age:25 ...}
All users:
  - John Doe (john@example.com)
  - Jane Smith (jane@example.com)
...
```

---

## 🧪 Ejecutar Tests

### Tests Básicos

```bash
# Todos los tests (2,100 casos)
go test ./test/features/user/... -v

# Con más casos (10,500 casos)
go test ./test/features/user/... -v -rapid.checks=500
```

### Tests por Categoría

```bash
go test ./test/features/user/... -v -run Create   # Solo CREATE
go test ./test/features/user/... -v -run Read     # Solo READ
go test ./test/features/user/... -v -run Update   # Solo UPDATE
go test ./test/features/user/... -v -run Delete   # Solo DELETE
```

### Test Específico

```bash
go test ./test/features/user/... -v -run TestProperty_UserCreate_ValidData
```

---

## 📊 Coverage

### Generar Reporte HTML

```bash
# Generar coverage
go test ./test/features/user/... \
  -coverprofile=coverage.out \
  -coverpkg=./internal/...

# Abrir en navegador
go tool cover -html=coverage.out
```

### Coverage por Función

```bash
go tool cover -func=coverage.out
```

**Resultado esperado:**
```
internal/domain/user.go:25:        Validate    100.0%
internal/service/user_service.go:15: CreateUser 100.0%
total:                             (statements) 97.8%
```

---

## 🔧 Comandos Útiles

### Limpiar Cache

```bash
go clean -testcache
rm -rf test/features/user/testdata/
```


### Reproducir un Fallo

```bash
# Cuando un test falla, usa el archivo .fail generado
go test ./test/features/user/... -v \
  -rapid.failfile="testdata/rapid/TestName/TestName-timestamp.fail"

# O usa la semilla directamente
go test ./test/features/user/... -v -rapid.seed=12345
```

---

## 📁 Estructura del Proyecto

```
├── cmd/main.go                      # Ejemplo de uso
├── internal/
│   ├── domain/
│   │   ├── user.go                 # Entidad User + validaciones
│   │   └── error.go                # Errores de dominio
│   ├── repository/
│   │   └── user_repository.go      # Persistencia en memoria
│   └── service/
│       └── user_service.go         # Lógica de negocio CRUD
├── test/
│   ├── features/user/              # Tests property-based
│   │   ├── create_test.go          # 4 tests CREATE
│   │   ├── read_test.go            # 5 tests READ
│   │   ├── update_test.go          # 5 tests UPDATE
│   │   └── delete_test.go          # 7 tests DELETE
│   ├── generators/
│   │   └── user_generators.go      # Generadores de datos
│   └── helpers/
│       └── test_helpers.go         # Utilidades de test
└── README.md
```

---

## ✅ Propiedades Probadas

### CREATE (4 tests)
- ✅ Datos válidos → usuario creado y recuperable
- ✅ Datos inválidos → error sin persistir
- ✅ Email duplicado → `ErrAlreadyExists`
- ✅ Creaciones concurrentes → IDs únicos

### READ (5 tests)
- ✅ Usuario existente → datos correctos
- ✅ Usuario inexistente → `ErrNotFound`
- ✅ Búsqueda por email → usuario correcto
- ✅ GetAll → todos los usuarios
- ✅ Lecturas concurrentes → consistencia

### UPDATE (5 tests)
- ✅ Datos válidos → actualización persistida
- ✅ Datos inválidos → error sin modificar
- ✅ Usuario inexistente → `ErrNotFound`
- ✅ Email duplicado → `ErrAlreadyExists`
- ✅ Updates secuenciales → cada uno persiste

### DELETE (7 tests)
- ✅ Usuario existente → eliminado
- ✅ Usuario inexistente → `ErrNotFound`
- ✅ Segunda eliminación → falla (no idempotente)
- ✅ Email liberado → permite reutilización
- ✅ Eliminación selectiva → solo el especificado
- ✅ Deletes concurrentes → uno sucede, otros fallan
- ✅ Eliminar todos → sistema vacío

---

## 🎯 Reglas de Negocio

| Campo | Validación |
|-------|------------|
| **Name** | 2-50 caracteres, solo letras y espacios |
| **Email** | Formato válido, único en el sistema |
| **Age** | 1-150 años |

---

## 🐛 Troubleshooting

### Error: `package not found`
```bash
go mod tidy
go mod download
```

### Tests fallan con `entity already exists`
```bash
# Limpiar cache
go clean -testcache
```

### Coverage sale vacío
```bash
# Verificar que existan archivos en internal/
ls -la internal/domain/
ls -la internal/service/
ls -la internal/repository/

# Ejecutar con -coverpkg explícito
go test ./test/features/user/... \
  -coverprofile=coverage.out \
  -coverpkg=./internal/domain,./internal/service,./internal/repository
```

---

## 📦 Dependencias

```go
require (
    github.com/google/uuid v1.6.0      // Generación de IDs
    pgregory.net/rapid v1.2.0          // Property-based testing
)
```

---

## 📚 Referencias

- [Property-Based Testing Guide](https://www.thesoftwarelounge.com/the-beginners-guide-to-property-based-testing/)
- [pgregory.net/rapid Documentation](https://pkg.go.dev/pgregory.net/rapid)
- [QuickCheck Paper](https://www.cs.tufts.edu/~nr/cs257/archive/john-hughes/quick.pdf)
