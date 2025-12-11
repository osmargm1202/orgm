# API Documentation - Calc API Management

Documentación completa de la API de gestión de proyectos y cálculos de ingeniería.

## Base URL

```
http://localhost:8000
```

## Endpoints de Salud

### GET /

Información general de la API.

**Ejemplo de solicitud:**
```bash
curl -X GET http://localhost:8000/
```

**Ejemplo de respuesta:**
```json
{
  "name": "Calc API Management",
  "version": "1.0.0",
  "description": "API de gestión de proyectos y cálculos de ingeniería",
  "status": "running"
}
```

**Códigos de estado:**
- `200 OK`: La API está funcionando correctamente

---

### GET /health

Verificación de salud de la API.

**Ejemplo de solicitud:**
```bash
curl -X GET http://localhost:8000/health
```

**Ejemplo de respuesta:**
```json
{
  "status": "healthy",
  "version": "1.0.0"
}
```

**Códigos de estado:**
- `200 OK`: La API está saludable

---

## Empresas

### POST /api/v1/empresas

Crear una nueva empresa.

**Ejemplo de solicitud:**
```bash
curl -X POST http://localhost:8000/api/v1/empresas \
  -H "Content-Type: application/json" \
  -d '{
    "nombre": "Empresa Ejemplo S.A.",
    "url_logo": "https://example.com/logo.png"
  }'
```

**Ejemplo de respuesta:**
```json
{
  "id": 1,
  "nombre": "Empresa Ejemplo S.A.",
  "url_logo": "https://example.com/logo.png",
  "created_at": "2024-01-15T10:30:00",
  "updated_at": "2024-01-15T10:30:00"
}
```

**Códigos de estado:**
- `201 Created`: Empresa creada exitosamente
- `422 Unprocessable Entity`: Error de validación en los datos

---

### GET /api/v1/empresas

Listar todas las empresas, ordenadas por ID.

**Parámetros de consulta:**
- `limit` (opcional): Número máximo de resultados a devolver (últimos N por ID)

**Ejemplo de solicitud:**
```bash
# Listar todas las empresas
curl -X GET http://localhost:8000/api/v1/empresas

# Listar las últimas 10 empresas
curl -X GET "http://localhost:8000/api/v1/empresas?limit=10"
```

**Ejemplo de respuesta:**
```json
[
  {
    "id": 1,
    "nombre": "Empresa Ejemplo S.A.",
    "url_logo": "https://example.com/logo.png",
    "created_at": "2024-01-15T10:30:00",
    "updated_at": "2024-01-15T10:30:00"
  },
  {
    "id": 2,
    "nombre": "Otra Empresa S.A.",
    "url_logo": null,
    "created_at": "2024-01-16T08:15:00",
    "updated_at": "2024-01-16T08:15:00"
  }
]
```

**Códigos de estado:**
- `200 OK`: Lista obtenida exitosamente

---

### GET /api/v1/empresas/search

Buscar empresas por nombre. Si no se proporciona query, lista todas ordenadas por ID.

**Parámetros de consulta:**
- `q` (opcional): Buscar por nombre (búsqueda parcial, case-insensitive).
- `limit` (opcional): Número máximo de resultados a devolver (últimos N por ID)

**Ejemplo de solicitud:**
```bash
# Buscar por nombre
curl -X GET "http://localhost:8000/api/v1/empresas/search?q=Ejemplo"

# Listar todas (últimas 20)
curl -X GET "http://localhost:8000/api/v1/empresas/search?limit=20"

# Listar todas sin límite
curl -X GET "http://localhost:8000/api/v1/empresas/search"
```

**Ejemplo de respuesta:**
```json
{
  "empresas": [
    {
      "id": 1,
      "nombre": "Empresa Ejemplo S.A.",
      "url_logo": "https://example.com/logo.png",
      "created_at": "2024-01-15T10:30:00",
      "updated_at": "2024-01-15T10:30:00"
    }
  ],
  "total": 1
}
```

**Códigos de estado:**
- `200 OK`: Búsqueda completada exitosamente

---

### GET /api/v1/empresas/{empresa_id}

Obtener una empresa por ID.

**Ejemplo de solicitud:**
```bash
curl -X GET http://localhost:8000/api/v1/empresas/1
```

**Ejemplo de respuesta:**
```json
{
  "id": 1,
  "nombre": "Empresa Ejemplo S.A.",
  "url_logo": "https://example.com/logo.png",
  "created_at": "2024-01-15T10:30:00",
  "updated_at": "2024-01-15T10:30:00"
}
```

**Códigos de estado:**
- `200 OK`: Empresa encontrada
- `404 Not Found`: Empresa no encontrada

---

### PUT /api/v1/empresas/{empresa_id}

Actualizar una empresa.

**Ejemplo de solicitud:**
```bash
curl -X PUT http://localhost:8000/api/v1/empresas/1 \
  -H "Content-Type: application/json" \
  -d '{
    "nombre": "Empresa Actualizada S.A.",
    "url_logo": "https://example.com/new-logo.png"
  }'
```

**Ejemplo de respuesta:**
```json
{
  "id": 1,
  "nombre": "Empresa Actualizada S.A.",
  "url_logo": "https://example.com/new-logo.png",
  "created_at": "2024-01-15T10:30:00",
  "updated_at": "2024-01-15T11:45:00"
}
```

**Códigos de estado:**
- `200 OK`: Empresa actualizada exitosamente
- `404 Not Found`: Empresa no encontrada
- `422 Unprocessable Entity`: Error de validación

**Nota:** Las empresas no pueden ser eliminadas por seguridad e integridad de datos.

---

## Ingenieros

### POST /api/v1/ingenieros

Crear un nuevo ingeniero.

**Ejemplo de solicitud:**
```bash
curl -X POST http://localhost:8000/api/v1/ingenieros \
  -H "Content-Type: application/json" \
  -d '{
    "nombre": "Juan Pérez",
    "profesion": "Ing. Eléctrico",
    "codia": "12345"
  }'
```

**Ejemplo de respuesta:**
```json
{
  "id": 1,
  "nombre": "Juan Pérez",
  "profesion": "Ing. Eléctrico",
  "codia": "12345",
  "created_at": "2024-01-15T10:30:00",
  "updated_at": "2024-01-15T10:30:00"
}
```

**Códigos de estado:**
- `201 Created`: Ingeniero creado exitosamente
- `422 Unprocessable Entity`: Error de validación

---

### GET /api/v1/ingenieros

Listar todos los ingenieros, ordenados por ID.

**Parámetros de consulta:**
- `limit` (opcional): Número máximo de resultados a devolver (últimos N por ID)

**Ejemplo de solicitud:**
```bash
# Listar todos los ingenieros
curl -X GET http://localhost:8000/api/v1/ingenieros

# Listar los últimos 10 ingenieros
curl -X GET "http://localhost:8000/api/v1/ingenieros?limit=10"
```

**Ejemplo de respuesta:**
```json
[
  {
    "id": 1,
    "nombre": "Juan Pérez",
    "profesion": "Ing. Eléctrico",
    "codia": "12345",
    "created_at": "2024-01-15T10:30:00",
    "updated_at": "2024-01-15T10:30:00"
  }
]
```

**Códigos de estado:**
- `200 OK`: Lista obtenida exitosamente

---

### GET /api/v1/ingenieros/search

Buscar ingenieros por nombre, profesión o codia. Si no se proporciona query, lista todos ordenados por ID.

**Parámetros de consulta:**
- `q` (opcional): Buscar por nombre, profesión o codia (búsqueda parcial, case-insensitive).
- `limit` (opcional): Número máximo de resultados a devolver (últimos N por ID)

**Ejemplo de solicitud:**
```bash
# Buscar por nombre
curl -X GET "http://localhost:8000/api/v1/ingenieros/search?q=Juan"

# Buscar por profesión
curl -X GET "http://localhost:8000/api/v1/ingenieros/search?q=Eléctrico"

# Listar todos (últimos 20)
curl -X GET "http://localhost:8000/api/v1/ingenieros/search?limit=20"

# Listar todos sin límite
curl -X GET "http://localhost:8000/api/v1/ingenieros/search"
```

**Ejemplo de respuesta:**
```json
{
  "ingenieros": [
    {
      "id": 1,
      "nombre": "Juan Pérez",
      "profesion": "Ing. Eléctrico",
      "codia": "12345",
      "created_at": "2024-01-15T10:30:00",
      "updated_at": "2024-01-15T10:30:00"
    }
  ],
  "total": 1
}
```

**Códigos de estado:**
- `200 OK`: Búsqueda completada exitosamente

---

### GET /api/v1/ingenieros/{ingeniero_id}

Obtener un ingeniero por ID.

**Ejemplo de solicitud:**
```bash
curl -X GET http://localhost:8000/api/v1/ingenieros/1
```

**Ejemplo de respuesta:**
```json
{
  "id": 1,
  "nombre": "Juan Pérez",
  "profesion": "Ing. Eléctrico",
  "codia": "12345",
  "created_at": "2024-01-15T10:30:00",
  "updated_at": "2024-01-15T10:30:00"
}
```

**Códigos de estado:**
- `200 OK`: Ingeniero encontrado
- `404 Not Found`: Ingeniero no encontrado

---

### PUT /api/v1/ingenieros/{ingeniero_id}

Actualizar un ingeniero.

**Ejemplo de solicitud:**
```bash
curl -X PUT http://localhost:8000/api/v1/ingenieros/1 \
  -H "Content-Type: application/json" \
  -d '{
    "nombre": "Juan Carlos Pérez",
    "profesion": "Ing. Eléctrico",
    "codia": "12345"
  }'
```

**Ejemplo de respuesta:**
```json
{
  "id": 1,
  "nombre": "Juan Carlos Pérez",
  "profesion": "Ing. Eléctrico",
  "codia": "12345",
  "created_at": "2024-01-15T10:30:00",
  "updated_at": "2024-01-15T11:45:00"
}
```

**Códigos de estado:**
- `200 OK`: Ingeniero actualizado exitosamente
- `404 Not Found`: Ingeniero no encontrado
- `422 Unprocessable Entity`: Error de validación

**Nota:** Los ingenieros no pueden ser eliminados por seguridad e integridad de datos.

---

## Proyectos

### POST /api/v1/proyectos

Crear un nuevo proyecto.

**Ejemplo de solicitud:**
```bash
curl -X POST http://localhost:8000/api/v1/proyectos \
  -H "Content-Type: application/json" \
  -d '{
    "nombre": "Proyecto Residencial",
    "cliente": "Cliente Ejemplo",
    "ubicacion": "Ciudad, País",
    "url_logo": "https://example.com/cliente-logo.png",
    "empresa_id": 1
  }'
```

**Ejemplo de respuesta:**
```json
{
  "id": 1,
  "nombre": "Proyecto Residencial",
  "cliente": "Cliente Ejemplo",
  "ubicacion": "Ciudad, País",
  "url_logo": "https://example.com/cliente-logo.png",
  "empresa_id": 1,
  "empresa": {
    "id": 1,
    "nombre": "Empresa Ejemplo S.A.",
    "url_logo": "https://example.com/logo.png",
    "created_at": "2024-01-15T10:30:00",
    "updated_at": "2024-01-15T10:30:00"
  },
  "created_at": "2024-01-15T10:30:00",
  "updated_at": "2024-01-15T10:30:00"
}
```

**Códigos de estado:**
- `201 Created`: Proyecto creado exitosamente
- `400 Bad Request`: Empresa no encontrada
- `422 Unprocessable Entity`: Error de validación

---

### GET /api/v1/proyectos

Listar todos los proyectos, ordenados por fecha de creación (más reciente primero).

**Parámetros de consulta:**
- `limit` (opcional): Número máximo de resultados a devolver (últimos N por fecha)

**Ejemplo de solicitud:**
```bash
# Listar todos los proyectos
curl -X GET http://localhost:8000/api/v1/proyectos

# Listar los últimos 10 proyectos
curl -X GET "http://localhost:8000/api/v1/proyectos?limit=10"
```

**Ejemplo de respuesta:**
```json
[
  {
    "id": 1,
    "nombre": "Proyecto Residencial",
    "cliente": "Cliente Ejemplo",
    "ubicacion": "Ciudad, País",
    "url_logo": "https://example.com/cliente-logo.png",
    "empresa_id": 1,
    "empresa": {
      "id": 1,
      "nombre": "Empresa Ejemplo S.A.",
      "url_logo": "https://example.com/logo.png",
      "created_at": "2024-01-15T10:30:00",
      "updated_at": "2024-01-15T10:30:00"
    },
    "created_at": "2024-01-15T10:30:00",
    "updated_at": "2024-01-15T10:30:00"
  }
]
```

**Códigos de estado:**
- `200 OK`: Lista obtenida exitosamente

---

### GET /api/v1/proyectos/{proyecto_id}

Obtener un proyecto por ID.

**Ejemplo de solicitud:**
```bash
curl -X GET http://localhost:8000/api/v1/proyectos/1
```

**Ejemplo de respuesta:**
```json
{
  "id": 1,
  "nombre": "Proyecto Residencial",
  "cliente": "Cliente Ejemplo",
  "ubicacion": "Ciudad, País",
  "url_logo": "https://example.com/cliente-logo.png",
  "empresa_id": 1,
  "empresa": {
    "id": 1,
    "nombre": "Empresa Ejemplo S.A.",
    "url_logo": "https://example.com/logo.png",
    "created_at": "2024-01-15T10:30:00",
    "updated_at": "2024-01-15T10:30:00"
  },
  "created_at": "2024-01-15T10:30:00",
  "updated_at": "2024-01-15T10:30:00"
}
```

**Códigos de estado:**
- `200 OK`: Proyecto encontrado
- `404 Not Found`: Proyecto no encontrado

---

### PUT /api/v1/proyectos/{proyecto_id}

Actualizar un proyecto.

**Ejemplo de solicitud:**
```bash
curl -X PUT http://localhost:8000/api/v1/proyectos/1 \
  -H "Content-Type: application/json" \
  -d '{
    "nombre": "Proyecto Residencial Actualizado",
    "cliente": "Cliente Ejemplo",
    "ubicacion": "Nueva Ciudad, País",
    "url_logo": "https://example.com/nuevo-logo.png"
  }'
```

**Ejemplo de respuesta:**
```json
{
  "id": 1,
  "nombre": "Proyecto Residencial Actualizado",
  "cliente": "Cliente Ejemplo",
  "ubicacion": "Nueva Ciudad, País",
  "url_logo": "https://example.com/nuevo-logo.png",
  "empresa_id": 1,
  "created_at": "2024-01-15T10:30:00",
  "updated_at": "2024-01-15T11:45:00"
}
```

**Códigos de estado:**
- `200 OK`: Proyecto actualizado exitosamente
- `400 Bad Request`: Empresa no encontrada (si se actualiza empresa_id)
- `404 Not Found`: Proyecto no encontrado
- `422 Unprocessable Entity`: Error de validación

---

### DELETE /api/v1/proyectos/{proyecto_id}

Eliminar un proyecto.

**Ejemplo de solicitud:**
```bash
curl -X DELETE http://localhost:8000/api/v1/proyectos/1
```

**Códigos de estado:**
- `204 No Content`: Proyecto eliminado exitosamente
- `404 Not Found`: Proyecto no encontrado

---

### GET /api/v1/proyectos/search

Buscar proyectos por nombre, cliente o ubicación. Si no se proporciona query, lista todos ordenados por fecha (más reciente primero).

**Parámetros de consulta:**
- `q` (opcional): Buscar por nombre, cliente o ubicación (búsqueda parcial, case-insensitive).
- `limit` (opcional): Número máximo de resultados a devolver (últimos N por fecha)

**Ejemplo de solicitud:**
```bash
# Buscar por nombre
curl -X GET "http://localhost:8000/api/v1/proyectos/search?q=Residencial"

# Buscar por cliente
curl -X GET "http://localhost:8000/api/v1/proyectos/search?q=Ejemplo"

# Listar todos (últimos 20)
curl -X GET "http://localhost:8000/api/v1/proyectos/search?limit=20"

# Listar todos sin límite
curl -X GET "http://localhost:8000/api/v1/proyectos/search"
```

**Ejemplo de respuesta:**
```json
{
  "proyectos": [
    {
      "id": 1,
      "nombre": "Proyecto Residencial",
      "cliente": "Cliente Ejemplo",
      "ubicacion": "Ciudad, País",
      "url_logo": "https://example.com/cliente-logo.png",
      "empresa_id": 1,
      "empresa": {
        "id": 1,
        "nombre": "Empresa Ejemplo S.A.",
        "url_logo": "https://example.com/logo.png",
        "created_at": "2024-01-15T10:30:00",
        "updated_at": "2024-01-15T10:30:00"
      },
      "created_at": "2024-01-15T10:30:00",
      "updated_at": "2024-01-15T10:30:00"
    }
  ],
  "total": 1
}
```

**Códigos de estado:**
- `200 OK`: Búsqueda completada exitosamente

---

## Tipos de Cálculo

Los tipos de cálculo son datos predefinidos que se inicializan automáticamente al iniciar la aplicación. Solo se pueden consultar mediante el endpoint GET.

### GET /api/v1/tipos-calculo

Listar todos los tipos de cálculo disponibles.

**Ejemplo de solicitud:**
```bash
curl -X GET http://localhost:8000/api/v1/tipos-calculo
```

**Ejemplo de respuesta:**
```json
[
  {
    "id": 1,
    "nombre": "Cálculo de baja tensión",
    "descripcion": "Cálculo de instalaciones eléctricas de baja tensión",
    "created_at": "2024-01-15T10:30:00"
  },
  {
    "id": 2,
    "nombre": "Cálculo de sistema de puesta a tierra",
    "descripcion": "Cálculo de sistemas de puesta a tierra para instalaciones eléctricas",
    "created_at": "2024-01-15T10:30:00"
  }
]
```

**Códigos de estado:**
- `200 OK`: Lista obtenida exitosamente

---
## Códigos de Estado HTTP

La API utiliza los siguientes códigos de estado HTTP:

- `200 OK`: Solicitud exitosa
- `201 Created`: Recurso creado exitosamente
- `204 No Content`: Solicitud exitosa sin contenido de respuesta
- `400 Bad Request`: Error en la solicitud (datos inválidos, relaciones no encontradas)
- `404 Not Found`: Recurso no encontrado
- `422 Unprocessable Entity`: Error de validación de datos

## Notas Importantes

1. **Propósito de la API**: Esta API está diseñada para gestionar únicamente los datos básicos de proyectos, empresas, ingenieros y tipos de cálculo.

2. **CRUD Completo**: La API proporciona operaciones CRUD completas para todos los recursos:
   - **Empresas**: CREATE, READ, UPDATE, SEARCH (NO DELETE - por seguridad e integridad de datos)
   - **Ingenieros**: CREATE, READ, UPDATE, SEARCH (NO DELETE - por seguridad e integridad de datos)
   - **Proyectos**: CREATE, READ, UPDATE, DELETE, SEARCH
   - **Tipos de Cálculo**: READ (solo lectura, datos predefinidos)

3. **Endpoints de Búsqueda**: Todos los recursos principales tienen endpoints de búsqueda flexibles:
   - **Búsqueda por nombre**: Los endpoints `/search` permiten buscar solo por nombre (búsqueda parcial, case-insensitive)
   - **Paginación opcional**: Todos los endpoints de listado y búsqueda aceptan el parámetro `limit` para limitar resultados
   - **Ordenamiento**: 
     - Proyectos: Ordenados por fecha de creación (más reciente primero)
     - Empresas e ingenieros: Ordenados por ID

4. **Tipos de Cálculo**: Los tipos de cálculo disponibles incluyen:
   - **Eléctricos**: Baja tensión, sistema de puesta a tierra, malla de puesta a tierra (SE), apantallamiento ángulos fijos, apantallamiento esfera rodante (SE)
   - **Hidrosanitarios**: Cisterna/séptico, pérdidas/bombeo, drenajes pluviales
   - **Civiles**: Zapata aislada de equipo, estructura metálica de equipos

5. **Relaciones**: 
   - Un proyecto pertenece a una empresa (incluye URL del logo del cliente y URL del logo de la empresa para uso global)
   - La ubicación del proyecto y las URLs de logos están disponibles para uso global en los programas cliente

6. **Restricciones de Eliminación**: 
   - Las empresas e ingenieros no pueden ser eliminados para mantener la integridad referencial de los datos históricos
   - Los proyectos pueden ser eliminados cuando sea necesario

7. **Tipos de Cálculo**:
   - Los tipos de cálculo se inicializan automáticamente al iniciar la aplicación con los tipos básicos predefinidos
   - Solo se pueden consultar mediante el endpoint GET `/api/v1/tipos-calculo`
   - No se pueden crear, modificar o eliminar mediante la API

