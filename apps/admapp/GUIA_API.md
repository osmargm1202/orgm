# Ejemplos de Uso - API v2.0

Ejemplos prácticos de cómo usar la API del sistema de cotizaciones y facturas.

## 🔑 Autenticación

Todos los endpoints requieren el header `X-Tenant-Id`:

```bash
X-Tenant-Id: 1
```

## 📑 Índice de Contenidos

1. [Clientes](#1-clientes) - 8 endpoints
2. [Proyectos](#2-proyectos) - 7 endpoints
3. [Cotizaciones](#3-cotizaciones) - 7 endpoints
4. [Presupuestos](#4-presupuestos) - 3 endpoints
5. [Notas](#5-notas) - 3 endpoints
6. [Facturas](#6-facturas) - 8 endpoints
7. [Pagos](#7-pagos) - **11 endpoints** (incluye comprobantes)
8. [Comprobantes (NC)](#8-comprobantes-nc) - **5 endpoints** (incluye bulk)
9. [Tenants](#9-tenants) - 8 endpoints
10. [Flujo Completo](#-flujo-completo-de-ejemplo)
11. [Testing Python](#-testing-con-python)
12. [Testing JavaScript](#-testing-con-javascriptfetch)
13. [Filtros](#-filtros-disponibles)

**Total: 75+ endpoints**

## 📋 Ejemplos por Categoría

### 1. Clientes

#### Crear un cliente
```bash
curl -X POST http://localhost:8000/api/clientes \
  -H "X-Tenant-Id: 1" \
  -H "Content-Type: application/json" \
  -d '{
    "nombre": "Empresa ABC S.A.",
    "nombre_comercial": "ABC Corp",
    "numero": "RNC-123456789",
    "correo": "contacto@abc.com",
    "direccion": "Av. Principal 123",
    "ciudad": "Santo Domingo",
    "provincia": "Distrito Nacional",
    "telefono": "809-555-1234",
    "representante": "Juan Pérez",
    "correo_representante": "juan@abc.com",
    "tipo_factura": "NCFC"
  }'
```

#### Listar clientes
```bash
curl -X GET http://localhost:8000/api/clientes \
  -H "X-Tenant-Id: 1"
```

#### Subir logo de cliente
```bash
curl -X POST http://localhost:8000/api/clientes/1/logo \
  -H "X-Tenant-Id: 1" \
  -F "file=@/path/to/logo.png"
```

---

### 2. Proyectos

#### Crear un proyecto
```bash
curl -X POST http://localhost:8000/api/proyectos \
  -H "X-Tenant-Id: 1" \
  -H "Content-Type: application/json" \
  -d '{
    "id_cliente": 1,
    "ubicacion": "Santo Domingo, DN",
    "nombre_proyecto": "Construcción Edificio Torre ABC",
    "descripcion": "Proyecto de construcción de edificio de oficinas de 10 plantas"
  }'
```

#### Listar proyectos de un cliente
```bash
curl -X GET http://localhost:8000/api/clientes/1/proyectos \
  -H "X-Tenant-Id: 1"
```

---

### 3. Cotizaciones

#### Crear una cotización
```bash
curl -X POST http://localhost:8000/api/cotizaciones \
  -H "X-Tenant-Id: 1" \
  -H "Content-Type: application/json" \
  -d '{
    "id_cliente": 1,
    "id_proyecto": 1,
    "id_servicio": 1,
    "moneda": "RD$",
    "fecha": "2025-10-08",
    "tasa_moneda": 1.0,
    "tiempo_entrega": "30",
    "avance": "60",
    "validez": 30,
    "estado": "GENERADA",
    "idioma": "ES",
    "descripcion": "Cotización para construcción edificio Torre ABC",
    "retencion": "NINGUNA",
    "descuentop": 5.0,
    "retencionp": 0.0,
    "itbisp": 18.0
  }'
```

#### Obtener cotización completa (con totales calculados)
```bash
curl -X GET http://localhost:8000/api/cotizaciones/1/full \
  -H "X-Tenant-Id: 1"
```

**Respuesta:**
```json
{
  "cotizacion": {
    "id": 1,
    "id_tenant": 1,
    "id_cliente": 1,
    "id_proyecto": 1,
    "moneda": "RD$",
    "descuentop": 5.0,
    "retencionp": 0.0,
    "itbisp": 18.0,
    ...
  },
  "presupuesto": {
    "categorias": [...]
  },
  "notas": {...},
  "totales": {
    "subtotal": 100000.00,
    "descuentom": 5000.00,
    "retencionm": 0.00,
    "itbism": 17100.00,
    "total_sin_itbis": 95000.00,
    "total": 112100.00
  }
}
```

---

### 4. Presupuestos

#### Guardar presupuesto de una cotización
```bash
curl -X PUT http://localhost:8000/api/cotizaciones/1/presupuesto \
  -H "X-Tenant-Id: 1" \
  -H "Content-Type: application/json" \
  -d '{
    "presupuesto": [
      {
        "id": "cat1",
        "item": "I-1",
        "descripcion": "Trabajos Preliminares",
        "categoria": "preliminares",
        "children": [
          {
            "id": "part1",
            "item": "P-1.1",
            "descripcion": "Limpieza y preparación de terreno",
            "cantidad": 1000,
            "unidad": "m²",
            "moneda": "RD$",
            "precio": 50.00,
            "total": 50000.00
          }
        ]
      },
      {
        "id": "cat2",
        "item": "I-2",
        "descripcion": "Estructura",
        "categoria": "estructura",
        "children": [
          {
            "id": "part2",
            "item": "P-2.1",
            "descripcion": "Concreto armado",
            "cantidad": 200,
            "unidad": "m³",
            "moneda": "RD$",
            "precio": 8000.00,
            "total": 1600000.00
          }
        ]
      }
    ]
  }'
```

#### Calcular totales sin guardar
```bash
curl -X GET http://localhost:8000/api/cotizaciones/1/presupuesto/calc \
  -H "X-Tenant-Id: 1"
```

---

### 5. Notas

#### Guardar notas de una cotización
```bash
curl -X POST http://localhost:8000/api/cotizaciones/1/notas \
  -H "X-Tenant-Id: 1" \
  -H "Content-Type: application/json" \
  -d '{
    "notas": {
      "entrega": "Tiempo de entrega: 30 días hábiles",
      "pago": "50% adelanto, 50% contra entrega",
      "validez": "Cotización válida por 30 días",
      "cuenta": "Banco Popular - Cuenta: 123456789",
      "observaciones": "Incluye materiales y mano de obra"
    }
  }'
```

---

### 6. Facturas

#### Crear factura desde cotización
```bash
curl -X POST http://localhost:8000/api/facturas \
  -H "X-Tenant-Id: 1" \
  -H "Content-Type: application/json" \
  -d '{
    "id_cotizacion": 1,
    "id_cliente": 1,
    "id_proyecto": 1,
    "moneda": "RD$",
    "tipo_factura": "NCFC",
    "fecha": "2025-10-08",
    "tasa_moneda": 1.0,
    "original": "VENDEDOR",
    "estado": "GENERADA",
    "idioma": "ES",
    "descuentop": 5.0,
    "retencionp": 0.0,
    "itbisp": 18.0
  }'
```

#### Asignar comprobante fiscal a factura
```bash
curl -X POST http://localhost:8000/api/facturas/1/comprobante \
  -H "X-Tenant-Id: 1" \
  -H "Content-Type: application/json" \
  -d '{
    "tipo": "NCFC",
    "numero": "B0100000001",
    "fecha": "2025-10-08"
  }'
```

---

### 7. Pagos

#### Registrar un pago
```bash
curl -X POST http://localhost:8000/api/pagos \
  -H "X-Tenant-Id: 1" \
  -H "Content-Type: application/json" \
  -d '{
    "id_cliente": 1,
    "moneda": "RD$",
    "monto": 50000.00,
    "fecha": "2025-10-08",
    "comprobante": "TRF-12345"
  }'
```

#### Asignar pago a cotización
```bash
curl -X POST http://localhost:8000/api/pagos/1/asignar \
  -H "X-Tenant-Id: 1" \
  -H "Content-Type: application/json" \
  -d '{
    "id_cotizacion": 1,
    "id_pago": 1,
    "monto": 50000.00
  }'
```

#### Ver pagos de una cotización
```bash
curl -X GET http://localhost:8000/api/cotizaciones/1/pagos \
  -H "X-Tenant-Id: 1"
```

#### Subir comprobante de pago (PDF o imagen)
```bash
curl -X POST http://localhost:8000/api/pagos/1/comprobante \
  -H "X-Tenant-Id: 1" \
  -F "file=@/path/to/comprobante.pdf"
```

O con una imagen:
```bash
curl -X POST http://localhost:8000/api/pagos/1/comprobante \
  -H "X-Tenant-Id: 1" \
  -F "file=@/path/to/comprobante.jpg"
```

**Extensiones permitidas:** `pdf`, `jpg`, `jpeg`, `png`

#### Obtener URL del comprobante de pago
```bash
curl -X GET http://localhost:8000/api/pagos/1/comprobante \
  -H "X-Tenant-Id: 1"
```

**Respuesta:**
```json
{
  "path": "tenant/1/pagos/1/comprobante.pdf"
}
```

#### Descargar comprobante de pago
```bash
curl -X GET http://localhost:8000/api/pagos/1/comprobante/download \
  -H "X-Tenant-Id: 1" \
  --output comprobante.pdf
```

#### Eliminar comprobante de pago
```bash
curl -X DELETE http://localhost:8000/api/pagos/1/comprobante \
  -H "X-Tenant-Id: 1"
```

---

### 8. Comprobantes (NC)

#### Crear comprobante
```bash
curl -X POST http://localhost:8000/api/nc \
  -H "X-Tenant-Id: 1" \
  -H "Content-Type: application/json" \
  -d '{
    "tipo": "NCFC",
    "numero": "B0100000001",
    "fecha": "2025-10-08"
  }'
```

#### Obtener siguiente número disponible
```bash
curl -X GET http://localhost:8000/api/nc/NCFC/siguiente \
  -H "X-Tenant-Id: 1"
```

**Respuesta:**
```json
{
  "tipo": "NCFC",
  "siguiente": "B0100000002",
  "ultimo": "B0100000001"
}
```

#### Crear comprobantes en bulk ⭐
```bash
curl -X POST http://localhost:8000/api/nc/bulk \
  -H "X-Tenant-Id: 1" \
  -H "Content-Type: application/json" \
  -d '{
    "tipo": "NCFC",
    "numero_final": "00000100",
    "fecha_validez": "2026-12-31"
  }'
```

**Ejemplo:**
- Último número en BD: `00000010`
- `numero_final`: `00000100`
- **Se crearán:** 90 comprobantes (del `00000011` al `00000100`)
- Todos con fecha de validez: `2026-12-31`

**Respuesta:**
```json
{
  "tipo": "NCFC",
  "cantidad_creada": 90,
  "numero_inicial": "00000011",
  "numero_final": "00000100",
  "fecha_validez": "2026-12-31",
  "mensaje": "Se crearon 90 comprobantes exitosamente"
}
```

**Notas:**
- El sistema obtiene automáticamente el último número usado
- Crea todos los números consecutivos hasta el `numero_final`
- Límite máximo: 10,000 comprobantes por request
- Todos los comprobantes tienen la misma fecha de validez

#### Listar tipos de comprobantes
```bash
curl -X GET http://localhost:8000/api/nc/tipos
```

**Respuesta:**
```json
{
  "tipos": ["NCF", "NCFC", "NCG", "NCRE", "NDC", "NDD"],
  "descripcion": {
    "NCF": "Número de Comprobante Fiscal",
    "NCFC": "Número de Comprobante Fiscal de Crédito",
    "NCG": "Número de Comprobante Gubernamental",
    "NCRE": "Nota de Crédito Electrónica",
    "NDC": "Nota de Débito de Crédito",
    "NDD": "Nota de Débito de Débito"
  }
}
```

---

### 9. Tenants

#### Crear tenant
```bash
curl -X POST http://localhost:8000/api/tenants \
  -H "Content-Type: application/json" \
  -d '{
    "numero": "RNC-987654321",
    "nombre": "Mi Empresa Constructora",
    "correo": "info@miempresa.com",
    "telefono": "809-555-5678",
    "direccion": "Calle Principal 456",
    "ciudad": "Santo Domingo",
    "provincia": "DN",
    "pais": "República Dominicana",
    "descripcion": "Empresa de construcción",
    "password": "mi_password_seguro"
  }'
```

#### Subir logo de tenant
```bash
curl -X POST http://localhost:8000/api/tenants/1/logo \
  -F "file=@/path/to/logo.png"
```

---

## 🔄 Flujo Completo de Ejemplo

### Paso 1: Crear Cliente
```bash
CLIENTE_ID=$(curl -X POST http://localhost:8000/api/clientes \
  -H "X-Tenant-Id: 1" \
  -H "Content-Type: application/json" \
  -d '{
    "nombre": "Empresa XYZ",
    "nombre_comercial": "XYZ Corp",
    "tipo_factura": "NCFC"
  }' | jq -r '.id')
```

### Paso 2: Crear Proyecto
```bash
PROYECTO_ID=$(curl -X POST http://localhost:8000/api/proyectos \
  -H "X-Tenant-Id: 1" \
  -H "Content-Type: application/json" \
  -d "{
    \"id_cliente\": $CLIENTE_ID,
    \"nombre_proyecto\": \"Proyecto ABC\",
    \"descripcion\": \"Descripción del proyecto\"
  }" | jq -r '.id')
```

### Paso 3: Crear Cotización
```bash
COT_ID=$(curl -X POST http://localhost:8000/api/cotizaciones \
  -H "X-Tenant-Id: 1" \
  -H "Content-Type: application/json" \
  -d "{
    \"id_cliente\": $CLIENTE_ID,
    \"id_proyecto\": $PROYECTO_ID,
    \"id_servicio\": 1,
    \"fecha\": \"2025-10-08\",
    \"descuentop\": 5.0,
    \"itbisp\": 18.0
  }" | jq -r '.id')
```

### Paso 4: Guardar Presupuesto
```bash
curl -X PUT http://localhost:8000/api/cotizaciones/$COT_ID/presupuesto \
  -H "X-Tenant-Id: 1" \
  -H "Content-Type: application/json" \
  -d '{
    "presupuesto": [
      {
        "id": "cat1",
        "item": "I-1",
        "descripcion": "Trabajos",
        "children": [
          {
            "id": "part1",
            "item": "P-1",
            "descripcion": "Trabajo 1",
            "cantidad": 100,
            "unidad": "m²",
            "precio": 1000,
            "total": 100000
          }
        ]
      }
    ]
  }'
```

### Paso 5: Ver Cotización Completa con Totales
```bash
curl -X GET http://localhost:8000/api/cotizaciones/$COT_ID/full \
  -H "X-Tenant-Id: 1" | jq
```

### Paso 6: Crear Factura
```bash
FACTURA_ID=$(curl -X POST http://localhost:8000/api/facturas \
  -H "X-Tenant-Id: 1" \
  -H "Content-Type: application/json" \
  -d "{
    \"id_cotizacion\": $COT_ID,
    \"id_cliente\": $CLIENTE_ID,
    \"id_proyecto\": $PROYECTO_ID,
    \"fecha\": \"2025-10-08\",
    \"tipo_factura\": \"NCFC\"
  }" | jq -r '.id')
```

### Paso 7: Asignar Comprobante
```bash
curl -X POST http://localhost:8000/api/facturas/$FACTURA_ID/comprobante \
  -H "X-Tenant-Id: 1" \
  -H "Content-Type: application/json" \
  -d '{
    "tipo": "NCFC",
    "numero": "B0100000001",
    "fecha": "2025-10-08"
  }'
```

---

## 🧪 Testing con Python

```python
import requests

BASE_URL = "http://localhost:8000"
HEADERS = {
    "X-Tenant-Id": "1",
    "Content-Type": "application/json"
}

# Crear cliente
response = requests.post(
    f"{BASE_URL}/api/clientes",
    headers=HEADERS,
    json={
        "nombre": "Cliente Test",
        "nombre_comercial": "Test Inc",
        "tipo_factura": "NCFC"
    }
)
cliente = response.json()
print(f"Cliente creado: {cliente['id']}")

# Listar clientes
response = requests.get(
    f"{BASE_URL}/api/clientes",
    headers={"X-Tenant-Id": "1"}
)
clientes = response.json()
print(f"Total clientes: {len(clientes)}")

# Obtener cotización completa
response = requests.get(
    f"{BASE_URL}/api/cotizaciones/1/full",
    headers={"X-Tenant-Id": "1"}
)
cotizacion = response.json()
print(f"Total: {cotizacion['totales']['total']}")

# Subir comprobante de pago
with open("comprobante.pdf", "rb") as file:
    response = requests.post(
        f"{BASE_URL}/api/pagos/1/comprobante",
        headers={"X-Tenant-Id": "1"},
        files={"file": file}
    )
    result = response.json()
    print(f"Comprobante subido: {result['path']}")

# Descargar comprobante de pago
response = requests.get(
    f"{BASE_URL}/api/pagos/1/comprobante/download",
    headers={"X-Tenant-Id": "1"}
)
with open("comprobante_descargado.pdf", "wb") as file:
    file.write(response.content)
print("Comprobante descargado")

# Crear comprobantes NC en bulk
response = requests.post(
    f"{BASE_URL}/api/nc/bulk",
    headers=HEADERS,
    json={
        "tipo": "NCFC",
        "numero_final": "00000500",
        "fecha_validez": "2026-12-31"
    }
)
result = response.json()
print(f"Comprobantes creados: {result['cantidad_creada']}")
print(f"Desde: {result['numero_inicial']} hasta: {result['numero_final']}")
```

---

## 📱 Testing con JavaScript/Fetch

```javascript
const BASE_URL = "http://localhost:8000";
const HEADERS = {
  "X-Tenant-Id": "1",
  "Content-Type": "application/json"
};

// Crear cliente
async function crearCliente() {
  const response = await fetch(`${BASE_URL}/api/clientes`, {
    method: "POST",
    headers: HEADERS,
    body: JSON.stringify({
      nombre: "Cliente Test",
      nombre_comercial: "Test Inc",
      tipo_factura: "NCFC"
    })
  });
  
  const cliente = await response.json();
  console.log("Cliente creado:", cliente.id);
  return cliente;
}

// Obtener cotización completa
async function obtenerCotizacion(id) {
  const response = await fetch(`${BASE_URL}/api/cotizaciones/${id}/full`, {
    headers: { "X-Tenant-Id": "1" }
  });
  
  const data = await response.json();
  console.log("Total:", data.totales.total);
  return data;
}

// Subir comprobante de pago
async function subirComprobante(pagoId, file) {
  const formData = new FormData();
  formData.append("file", file);
  
  const response = await fetch(`${BASE_URL}/api/pagos/${pagoId}/comprobante`, {
    method: "POST",
    headers: { "X-Tenant-Id": "1" },
    body: formData
  });
  
  const result = await response.json();
  console.log("Comprobante subido:", result.path);
  return result;
}

// Descargar comprobante de pago
async function descargarComprobante(pagoId) {
  const response = await fetch(`${BASE_URL}/api/pagos/${pagoId}/comprobante/download`, {
    headers: { "X-Tenant-Id": "1" }
  });
  
  const blob = await response.blob();
  
  // Crear enlace de descarga
  const url = window.URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = `comprobante_${pagoId}.pdf`;
  a.click();
  
  console.log("Comprobante descargado");
}

// Crear comprobantes NC en bulk
async function crearComprobantesBulk(tipo, numeroFinal, fechaValidez) {
  const response = await fetch(`${BASE_URL}/api/nc/bulk`, {
    method: "POST",
    headers: HEADERS,
    body: JSON.stringify({
      tipo: tipo,
      numero_final: numeroFinal,
      fecha_validez: fechaValidez
    })
  });
  
  const result = await response.json();
  console.log(`Comprobantes creados: ${result.cantidad_creada}`);
  console.log(`Desde: ${result.numero_inicial} hasta: ${result.numero_final}`);
  return result;
}

// Ejemplo de uso con input file
// <input type="file" id="comprobante" accept=".pdf,.jpg,.jpeg,.png">
const input = document.getElementById("comprobante");
input.addEventListener("change", async (e) => {
  const file = e.target.files[0];
  await subirComprobante(1, file);
});

// Ejecutar
crearCliente().then(cliente => {
  console.log("Cliente:", cliente);
});

// Crear 500 comprobantes NCFC
crearComprobantesBulk("NCFC", "00000500", "2026-12-31");
```

---

## 🔍 Filtros Disponibles

### Clientes
```bash
# Incluir inactivos
GET /api/clientes?incluir_inactivos=true
```

### Proyectos
```bash
# Por cliente
GET /api/proyectos?id_cliente=1

# Incluir inactivos
GET /api/proyectos?incluir_inactivos=true
```

### Cotizaciones
```bash
# Por cliente
GET /api/cotizaciones?id_cliente=1

# Por proyecto
GET /api/cotizaciones?id_proyecto=1

# Por estado
GET /api/cotizaciones?estado=GENERADA

# Combinados
GET /api/cotizaciones?id_cliente=1&estado=APROBADA
```

### Facturas
```bash
# Por cliente
GET /api/facturas?id_cliente=1

# Por estado
GET /api/facturas?estado=PAGADA
```

### Pagos
```bash
# Por cliente
GET /api/pagos?id_cliente=1
```

---

## 📊 Respuestas de Error

```json
{
  "detail": "Cliente no encontrado"
}
```

```json
{
  "detail": "X-Tenant-Id header requerido"
}
```

```json
{
  "detail": "Error interno del servidor"
}
```

---

## 📋 Resumen de Endpoints

### Tabla de Referencia Rápida

| Categoría | Método | Endpoint | Descripción |
|-----------|--------|----------|-------------|
| **Clientes** | POST | `/api/clientes` | Crear cliente |
| | GET | `/api/clientes` | Listar clientes |
| | GET | `/api/clientes/{id}` | Obtener cliente |
| | PUT | `/api/clientes/{id}` | Actualizar cliente |
| | DELETE | `/api/clientes/{id}` | Desactivar cliente |
| | POST | `/api/clientes/{id}/restore` | Reactivar cliente |
| | POST | `/api/clientes/{id}/logo` | Subir logo |
| | GET | `/api/clientes/{id}/logo` | Obtener logo |
| **Proyectos** | POST | `/api/proyectos` | Crear proyecto |
| | GET | `/api/proyectos` | Listar proyectos |
| | GET | `/api/proyectos/{id}` | Obtener proyecto |
| | PUT | `/api/proyectos/{id}` | Actualizar proyecto |
| | DELETE | `/api/proyectos/{id}` | Desactivar proyecto |
| | POST | `/api/proyectos/{id}/restore` | Reactivar proyecto |
| | GET | `/api/clientes/{id}/proyectos` | Proyectos de cliente |
| **Cotizaciones** | POST | `/api/cotizaciones` | Crear cotización |
| | GET | `/api/cotizaciones` | Listar cotizaciones |
| | GET | `/api/cotizaciones/{id}` | Obtener cotización |
| | PUT | `/api/cotizaciones/{id}` | Actualizar cotización |
| | DELETE | `/api/cotizaciones/{id}` | Desactivar cotización |
| | POST | `/api/cotizaciones/{id}/restore` | Reactivar cotización |
| | GET | `/api/cotizaciones/{id}/full` | Cotización completa + totales |
| **Presupuestos** | GET | `/api/cotizaciones/{id}/presupuesto` | Obtener presupuesto |
| | PUT | `/api/cotizaciones/{id}/presupuesto` | Guardar presupuesto |
| | GET | `/api/cotizaciones/{id}/presupuesto/calc` | Calcular totales |
| **Notas** | POST | `/api/cotizaciones/{id}/notas` | Agregar notas |
| | GET | `/api/cotizaciones/{id}/notas` | Obtener notas |
| | PUT | `/api/cotizaciones/{id}/notas` | Actualizar notas |
| **Facturas** | POST | `/api/facturas` | Crear factura |
| | GET | `/api/facturas` | Listar facturas |
| | GET | `/api/facturas/{id}` | Obtener factura |
| | PUT | `/api/facturas/{id}` | Actualizar factura |
| | DELETE | `/api/facturas/{id}` | Desactivar factura |
| | POST | `/api/facturas/{id}/restore` | Reactivar factura |
| | POST | `/api/facturas/{id}/comprobante` | Asignar comprobante |
| | GET | `/api/facturas/{id}/full` | Factura completa |
| **Pagos** | POST | `/api/pagos` | Registrar pago |
| | GET | `/api/pagos` | Listar pagos |
| | GET | `/api/pagos/{id}` | Obtener pago |
| | PUT | `/api/pagos/{id}` | Actualizar pago |
| | DELETE | `/api/pagos/{id}` | Eliminar pago |
| | POST | `/api/pagos/{id}/asignar` | Asignar a cotización |
| | GET | `/api/cotizaciones/{id}/pagos` | Pagos de cotización |
| | **POST** | **`/api/pagos/{id}/comprobante`** | **Subir comprobante** ⭐ |
| | **GET** | **`/api/pagos/{id}/comprobante`** | **Obtener comprobante** ⭐ |
| | **GET** | **`/api/pagos/{id}/comprobante/download`** | **Descargar comprobante** ⭐ |
| | **DELETE** | **`/api/pagos/{id}/comprobante`** | **Eliminar comprobante** ⭐ |
| **Comprobantes NC** | GET | `/api/nc` | Listar comprobantes |
| | POST | `/api/nc` | Crear comprobante |
| | **POST** | **`/api/nc/bulk`** | **Crear en bulk** ⭐ |
| | GET | `/api/nc/tipos` | Tipos disponibles |
| | GET | `/api/nc/{tipo}/siguiente` | Siguiente número |
| **Tenants** | POST | `/api/tenants` | Crear tenant |
| | GET | `/api/tenants` | Listar tenants |
| | GET | `/api/tenants/{id}` | Obtener tenant |
| | PUT | `/api/tenants/{id}` | Actualizar tenant |
| | DELETE | `/api/tenants/{id}` | Desactivar tenant |
| | POST | `/api/tenants/{id}/restore` | Reactivar tenant |
| | POST | `/api/tenants/{id}/logo` | Subir logo |
| | GET | `/api/tenants/{id}/logo` | Obtener logo |
| **Catálogos** | GET | `/api/servicios` | Listar servicios |
| | GET | `/api/tipos-factura` | Tipos de factura |

**Total: 75 endpoints** | ⭐ = Endpoints nuevos agregados

### Rutas de Almacenamiento en R2

| Tipo | Bucket | Ruta |
|------|--------|------|
| **Logo Cliente** | Público | `clientes/{id}/logo.png` |
| **Logo Tenant** | Privado | `tenant/{id}/logo.png` |
| **Comprobante Pago** ⭐ | Privado | `tenant/{tenant_id}/pagos/{pago_id}/comprobante.{ext}` |
| **PDF Cotización** | Privado | `tenant/{tenant_id}/cot/{cot_id}.pdf` |
| **PDF Factura** | Privado | `tenant/{tenant_id}/fac/{fac_id}.pdf` |

---

