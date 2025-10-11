# Ejemplos de Uso - API v2.1.0

Ejemplos prácticos de cómo usar la API del sistema de cotizaciones y facturas.

## 🔑 Autenticación

Todos los endpoints requieren el header `X-Tenant-Id`:

```bash
X-Tenant-Id: 1
```

## ⚠️ Notas Importantes

### Rutas de Almacenamiento con Multitenancy
Todos los archivos en R2 incluyen `tenant_id` para evitar conflictos:

- **Logos de clientes:** `tenant/{tenant_id}/clientes/{id}/logo.png`
- **Logos de tenants:** `tenant/{id}/logo.png`
- **Comprobantes de pago:** `tenant/{tenant_id}/pagos/{id}/comprobante.{ext}`
- **PDFs:** `tenant/{tenant_id}/cot/{id}.pdf` o `fac/{id}.pdf`

**Importante:** Si tienes logos en rutas antiguas (`clientes/{id}/logo.png`), necesitarás migrarlos a las nuevas rutas.

### Tipos de Archivo Permitidos
- **Logos:** PNG, JPEG, JPG, WEBP
- **Comprobantes de pago:** PDF, PNG, JPEG, JPG
- **Formatos de respuesta:** JSON (la mayoría) o bytes directos (PDFs, imágenes)

## 📑 Índice de Contenidos

1. [Clientes](#1-clientes) - **10 endpoints** (incluye búsqueda y validación)
2. [Proyectos](#2-proyectos) - **9 endpoints** (incluye búsqueda y validación)
3. [Cotizaciones](#3-cotizaciones) - **12 endpoints** (incluye PDF, búsqueda, recientes)
4. [Presupuestos](#4-presupuestos) - 3 endpoints
5. [Notas](#5-notas) - 3 endpoints
6. [Facturas](#6-facturas) - **11 endpoints** (incluye parcial y PDF)
7. [Pagos](#7-pagos) - **15 endpoints** (incluye comprobantes y avanzados)
8. [Comprobantes (NC)](#8-comprobantes-nc) - 5 endpoints (incluye bulk)
9. [Tenants](#9-tenants) - 8 endpoints
10. [Dashboard](#10-dashboard) - **4 endpoints** ⭐ NUEVO
11. [Configuración](#11-configuración) - **2 endpoints** ⭐ NUEVO
12. [Búsqueda Avanzada](#12-búsqueda-avanzada) - Ejemplos
13. [Flujo Completo](#-flujo-completo-de-ejemplo)
14. [Testing Python](#-testing-con-python)
15. [Testing JavaScript](#-testing-con-javascriptfetch)
16. [Filtros](#-filtros-disponibles)

**Total: ~90 endpoints** (+24 nuevos)

## 📋 Ejemplos por Categoría

### 1. Clientes

#### Crear un cliente
**Request:** `ClienteCreate` (JSON)  
**Response:** `ClienteResponse` (JSON)
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
**Request:** Query params opcionales (`incluir_inactivos=bool`)  
**Response:** `List[ClienteResponse]` (JSON)

```bash
curl -X GET http://localhost:8000/api/clientes \
  -H "X-Tenant-Id: 1"
```

#### Subir logo de cliente
**Request:** `multipart/form-data` con archivo  
**Response:** JSON con URL
```bash
curl -X POST http://localhost:8000/api/clientes/1/logo \
  -H "X-Tenant-Id: 1" \
  -F "file=@/path/to/logo.png"
```

#### Buscar clientes ⭐
**Request:** Query param `q=termino`  
**Response:** `List[ClienteResponse]` (JSON)

```bash
curl -X GET "http://localhost:8000/api/clientes/search?q=ABC" \
  -H "X-Tenant-Id: 1"
```

Busca en: nombre, nombre_comercial, RNC

#### Validar datos de cliente ⭐
**Request:** `ClienteValidate` (JSON)  
**Response:** JSON con `{valid: bool, errors: []}`
```bash
curl -X POST http://localhost:8000/api/clientes/validate \
  -H "X-Tenant-Id: 1" \
  -H "Content-Type: application/json" \
  -d '{
    "nombre": "Empresa XYZ",
    "numero": "RNC-123456789"
  }'
```

**Respuesta:**
```json
{
  "valid": true,
  "errors": []
}
```

O si hay duplicados:
```json
{
  "valid": false,
  "errors": [
    {"field": "numero", "message": "Ya existe un cliente con este RNC"}
  ]
}
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

#### Validar datos de proyecto ⭐
```bash
curl -X POST http://localhost:8000/api/proyectos/validate \
  -H "X-Tenant-Id: 1" \
  -H "Content-Type: application/json" \
  -d '{
    "nombre_proyecto": "Torre ABC",
    "ubicacion": "Santo Domingo"
  }'
```

#### Crear cotización desde proyecto ⭐
```bash
curl -X POST http://localhost:8000/api/proyectos/1/crear-cotizacion \
  -H "X-Tenant-Id: 1" \
  -H "Content-Type: application/json" \
  -d '{
    "id_servicio": 1,
    "fecha": "2025-10-08",
    "descuentop": 5.0,
    "itbisp": 18.0
  }'
```

El `id_cliente` e `id_proyecto` se asignan automáticamente del proyecto.

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

#### Obtener últimas cotizaciones ⭐
```bash
curl -X GET "http://localhost:8000/api/cotizaciones/recientes?limit=10" \
  -H "X-Tenant-Id: 1"
```

#### Verificar cambios pendientes ⭐
```bash
curl -X GET http://localhost:8000/api/cotizaciones/1/has-changes \
  -H "X-Tenant-Id: 1"
```

**Respuesta:**
```json
{
  "has_changes": false,
  "message": "No hay cambios pendientes",
  "last_update": "2025-10-08T10:30:00"
}
```

#### Generar PDF de cotización ⭐
**Request:** Query param `idioma=es` (opcional)  
**Response:** `application/pdf` (bytes directos)

```bash
curl -X GET "http://localhost:8000/api/cotizaciones/1/pdf?idioma=es" \
  -H "X-Tenant-Id: 1" \
  --output cotizacion.pdf
```

#### Buscar cotización por ID ⭐
```bash
curl -X GET http://localhost:8000/api/cotizaciones/by-id/1 \
  -H "X-Tenant-Id: 1"
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
**Request:** `FacturaCreate` (JSON)  
**Response:** `FacturaResponse` (JSON)
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

#### Crear factura parcial (monto exacto) ⭐
**Request:** `FacturaParcialCreate` (JSON) con `monto_facturar`  
**Response:** `FacturaParcialResponse` (JSON) con factor y presupuesto reducido
```bash
curl -X POST http://localhost:8000/api/facturas/parcial \
  -H "X-Tenant-Id: 1" \
  -H "Content-Type: application/json" \
  -d '{
    "id_cotizacion": 1,
    "id_cliente": 1,
    "id_proyecto": 1,
    "monto_facturar": 59998.50,
    "fecha": "2025-10-08",
    "tipo_factura": "NCFC"
  }'
```

**Cómo funciona:**
- Total de cotización: $100,000
- Monto a facturar: $59,998.50 (exacto del banco)
- Factor calculado: 0.599985 (59.9985%)
- Presupuesto reducido al 59.9985%
- Total final: $59,998.50 (exacto)

**Respuesta:**
```json
{
  "factura": {...},
  "factor_aplicado": 0.599985,
  "presupuesto_reducido": {...},
  "totales_calculados": {
    "subtotal": 50845.17,
    "descuentom": 2542.26,
    "total": 59998.50
  }
}
```

#### Calcular factura parcial sin guardar ⭐
```bash
curl -X POST http://localhost:8000/api/facturas/calcular-parcial \
  -H "X-Tenant-Id: 1" \
  -H "Content-Type: application/json" \
  -d '{
    "id_cotizacion": 1,
    "monto_facturar": 60000.00
  }'
```

Preview de cómo quedaría la factura.

#### Generar PDF de factura ⭐
**Request:** Query param `idioma=es` (opcional)  
**Response:** `application/pdf` (bytes directos)

```bash
curl -X GET "http://localhost:8000/api/facturas/1/pdf?idioma=es" \
  -H "X-Tenant-Id: 1" \
  --output factura.pdf
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
**Request:** `multipart/form-data` con archivo  
**Response:** JSON con `{path, message}`  
**Formatos permitidos:** PDF, PNG, JPEG, JPG

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
**Request:** Ninguno  
**Response:** Bytes directos (`application/pdf` o `image/*`)

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

#### Obtener pagos sin asignar ⭐
**Request:** Query param `id_cliente` (opcional)  
**Response:** JSON con array de pagos con monto disponible

```bash
curl -X GET "http://localhost:8000/api/pagos/sin-asignar?id_cliente=1" \
  -H "X-Tenant-Id: 1"
```

**Respuesta:**
```json
[
  {
    "pago": {...},
    "monto_total": 100000.00,
    "monto_asignado": 50000.00,
    "monto_disponible": 50000.00
  }
]
```

#### Asignar pago por porcentaje ⭐
```bash
curl -X POST http://localhost:8000/api/pagos/1/asignar-porcentaje \
  -H "X-Tenant-Id: 1" \
  -H "Content-Type: application/json" \
  -d '{
    "id_cotizacion": 1,
    "id_pago": 1,
    "porcentaje": 50.0
  }'
```

Calcula automáticamente el 50% del total de la cotización y lo asigna.

#### Obtener resumen de asignaciones ⭐
```bash
curl -X GET http://localhost:8000/api/pagos/1/resumen-asignaciones \
  -H "X-Tenant-Id: 1"
```

**Respuesta:**
```json
{
  "pago_id": 1,
  "monto_total": 100000.00,
  "monto_asignado": 75000.00,
  "monto_disponible": 25000.00,
  "asignaciones": [
    {
      "id_asignacion": 1,
      "id_cotizacion": 1,
      "monto": 50000.00,
      "fecha_cotizacion": "2025-10-01",
      "descripcion_cotizacion": "Proyecto ABC"
    },
    {
      "id_asignacion": 2,
      "id_cotizacion": 2,
      "monto": 25000.00,
      "fecha_cotizacion": "2025-10-05",
      "descripcion_cotizacion": "Proyecto XYZ"
    }
  ]
}
```

#### Calcular monto por porcentaje ⭐
```bash
curl -X GET "http://localhost:8000/api/pagos/1/calcular-montos?porcentaje=50&cotizacion_id=1" \
  -H "X-Tenant-Id: 1"
```

**Respuesta:**
```json
{
  "cotizacion_id": 1,
  "porcentaje": 50.0,
  "monto_calculado": 50000.00
}
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
**Request:** `NCBulkCreate` (JSON) con tipo, numero_final, fecha_validez  
**Response:** JSON con cantidad_creada, numero_inicial, numero_final

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

### 10. Dashboard

#### Obtener resumen de pagos de un cliente
**Request:** Ninguno  
**Response:** `ResumenPagosCliente` (JSON) con totales, saldo, avance %
```bash
curl -X GET http://localhost:8000/api/clientes/1/resumen-pagos \
  -H "X-Tenant-Id: 1"
```

**Respuesta:**
```json
{
  "cliente_id": 1,
  "nombre_cliente": "Empresa ABC",
  "total_cotizaciones": 500000.00,
  "total_pagado": 300000.00,
  "saldo_pendiente": 200000.00,
  "cotizaciones_count": 5,
  "facturas_count": 3,
  "avance_porcentaje": 60.00,
  "cotizaciones": [...]
}
```

#### Obtener cotizaciones pendientes de pago
**Request:** Ninguno  
**Response:** `List[CotizacionPendiente]` (JSON) con saldo por cotización

```bash
curl -X GET http://localhost:8000/api/clientes/1/cotizaciones-pendientes \
  -H "X-Tenant-Id: 1"
```

**Respuesta:**
```json
[
  {
    "id": 1,
    "fecha": "2025-10-01",
    "estado": "GENERADA",
    "descripcion": "Proyecto ABC",
    "total": 100000.00,
    "pagado": 50000.00,
    "saldo": 50000.00,
    "avance": 50.00
  }
]
```

#### Obtener estados de múltiples clientes/cotizaciones
**Request:** Query params `cliente_ids=1,2,3&cotizacion_ids=1,2,3`  
**Response:** `EstadosMultiplesResponse` (JSON)

```bash
curl -X GET "http://localhost:8000/api/dashboard/estados?cliente_ids=1,2,3&cotizacion_ids=1,2,3" \
  -H "X-Tenant-Id: 1"
```

#### Generar PDF de resumen de estados
**Request:** JSON con `cliente_ids` y `cotizacion_ids` (arrays)  
**Response:** `application/pdf` (bytes directos)
```bash
curl -X POST http://localhost:8000/api/dashboard/estados/pdf \
  -H "X-Tenant-Id: 1" \
  -H "Content-Type: application/json" \
  -d '{
    "cliente_ids": [1, 2, 3],
    "cotizacion_ids": [1, 2, 3]
  }' \
  --output resumen_estados.pdf
```

---

### 11. Configuración

#### Obtener configuración del usuario
**Request:** Ninguno  
**Response:** `ConfigUsuarioResponse` (JSON) con preferencias
```bash
curl -X GET http://localhost:8000/api/config/usuario \
  -H "X-Tenant-Id: 1"
```

**Respuesta:**
```json
{
  "id": 1,
  "id_tenant": 1,
  "preferencias": {
    "moneda_default": "RD$",
    "idioma": "ES",
    "tiempo_entrega_default": "30",
    "avance_default": "60",
    "validez_default": 30,
    "itbisp_default": 18.0
  },
  "fecha_creacion": "2025-10-08T10:00:00",
  "fecha_actualizacion": "2025-10-08T10:00:00"
}
```

#### Actualizar configuración
**Request:** `ConfigUsuarioUpdate` (JSON) con preferencias  
**Response:** `ConfigUsuarioResponse` (JSON)

```bash
curl -X PUT http://localhost:8000/api/config/usuario \
  -H "X-Tenant-Id: 1" \
  -H "Content-Type: application/json" \
  -d '{
    "preferencias": {
      "moneda_default": "USD",
      "idioma": "EN",
      "itbisp_default": 16.0
    }
  }'
```

---

### 12. Búsqueda Avanzada

#### Buscar clientes
```bash
curl -X GET "http://localhost:8000/api/clientes/search?q=ABC" \
  -H "X-Tenant-Id: 1"
```

Busca en: nombre, nombre_comercial, RNC

#### Buscar cotizaciones
```bash
curl -X GET "http://localhost:8000/api/cotizaciones/search?q=Torre" \
  -H "X-Tenant-Id: 1"
```

Busca en: nombre de cliente, proyecto, servicio, descripción

#### Buscar pagos
```bash
curl -X GET "http://localhost:8000/api/pagos/search?q=TRF-123" \
  -H "X-Tenant-Id: 1"
```

Busca en: nombre de cliente, comprobante, fecha

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

# Buscar clientes
response = requests.get(
    f"{BASE_URL}/api/clientes/search?q=ABC",
    headers={"X-Tenant-Id": "1"}
)
clientes = response.json()
print(f"Clientes encontrados: {len(clientes)}")

# Obtener resumen de pagos de cliente
response = requests.get(
    f"{BASE_URL}/api/clientes/1/resumen-pagos",
    headers={"X-Tenant-Id": "1"}
)
resumen = response.json()
print(f"Total: {resumen['total_cotizaciones']}, Pagado: {resumen['total_pagado']}, Saldo: {resumen['saldo_pendiente']}")

# Crear factura parcial con monto exacto
response = requests.post(
    f"{BASE_URL}/api/facturas/parcial",
    headers=HEADERS,
    json={
        "id_cotizacion": 1,
        "id_cliente": 1,
        "id_proyecto": 1,
        "monto_facturar": 59998.50,
        "fecha": "2025-10-08"
    }
)
factura_parcial = response.json()
print(f"Factor aplicado: {factura_parcial['factor_aplicado']} ({factura_parcial['factor_aplicado']*100:.2f}%)")

# Obtener pagos sin asignar
response = requests.get(
    f"{BASE_URL}/api/pagos/sin-asignar",
    headers={"X-Tenant-Id": "1"}
)
pagos_sin_asignar = response.json()
print(f"Pagos sin asignar: {len(pagos_sin_asignar)}")

# Generar PDF de cotización
response = requests.get(
    f"{BASE_URL}/api/cotizaciones/1/pdf",
    headers={"X-Tenant-Id": "1"}
)
with open("cotizacion.pdf", "wb") as file:
    file.write(response.content)
print("PDF de cotización generado")
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

// Buscar clientes
async function buscarClientes(termino) {
  const response = await fetch(`${BASE_URL}/api/clientes/search?q=${termino}`, {
    headers: { "X-Tenant-Id": "1" }
  });
  const clientes = await response.json();
  console.log(`Clientes encontrados: ${clientes.length}`);
  return clientes;
}

// Obtener resumen de pagos
async function obtenerResumenPagos(clienteId) {
  const response = await fetch(`${BASE_URL}/api/clientes/${clienteId}/resumen-pagos`, {
    headers: { "X-Tenant-Id": "1" }
  });
  const resumen = await response.json();
  console.log(`Avance: ${resumen.avance_porcentaje}%`);
  return resumen;
}

// Crear factura parcial
async function crearFacturaParcial(cotizacionId, montoFacturar) {
  const response = await fetch(`${BASE_URL}/api/facturas/parcial`, {
    method: "POST",
    headers: HEADERS,
    body: JSON.stringify({
      id_cotizacion: cotizacionId,
      id_cliente: 1,
      id_proyecto: 1,
      monto_facturar: montoFacturar,
      fecha: "2025-10-08"
    })
  });
  const result = await response.json();
  console.log(`Factor: ${(result.factor_aplicado * 100).toFixed(2)}%`);
  return result;
}

// Generar PDF de cotización
async function generarPDFCotizacion(cotizacionId) {
  const response = await fetch(`${BASE_URL}/api/cotizaciones/${cotizacionId}/pdf`, {
    headers: { "X-Tenant-Id": "1" }
  });
  const blob = await response.blob();
  
  // Descargar
  const url = window.URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = `cotizacion_${cotizacionId}.pdf`;
  a.click();
}

// Usar funciones
buscarClientes("ABC");
obtenerResumenPagos(1);
crearFacturaParcial(1, 59998.50);
generarPDFCotizacion(1);
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
| | **GET** | **`/api/clientes/search?q=`** | **Buscar clientes** ⭐ |
| | **POST** | **`/api/clientes/validate`** | **Validar datos** ⭐ |
| **Proyectos** | POST | `/api/proyectos` | Crear proyecto |
| | GET | `/api/proyectos` | Listar proyectos |
| | GET | `/api/proyectos/{id}` | Obtener proyecto |
| | PUT | `/api/proyectos/{id}` | Actualizar proyecto |
| | DELETE | `/api/proyectos/{id}` | Desactivar proyecto |
| | POST | `/api/proyectos/{id}/restore` | Reactivar proyecto |
| | GET | `/api/clientes/{id}/proyectos` | Proyectos de cliente |
| | **GET** | **`/api/proyectos/by-id/{id}`** | **Búsqueda directa** ⭐ |
| | **POST** | **`/api/proyectos/validate`** | **Validar datos** ⭐ |
| | **POST** | **`/api/proyectos/{id}/crear-cotizacion`** | **Crear cotización** ⭐ |
| **Cotizaciones** | POST | `/api/cotizaciones` | Crear cotización |
| | GET | `/api/cotizaciones` | Listar cotizaciones |
| | GET | `/api/cotizaciones/{id}` | Obtener cotización |
| | PUT | `/api/cotizaciones/{id}` | Actualizar cotización |
| | DELETE | `/api/cotizaciones/{id}` | Desactivar cotización |
| | POST | `/api/cotizaciones/{id}/restore` | Reactivar cotización |
| | GET | `/api/cotizaciones/{id}/full` | Cotización completa + totales |
| | **GET** | **`/api/cotizaciones/recientes?limit=10`** | **Últimas N** ⭐ |
| | **GET** | **`/api/cotizaciones/{id}/has-changes`** | **Verificar cambios** ⭐ |
| | **GET** | **`/api/cotizaciones/{id}/pdf`** | **Generar PDF** ⭐ |
| | **GET** | **`/api/cotizaciones/by-id/{id}`** | **Búsqueda directa** ⭐ |
| | **GET** | **`/api/cotizaciones/search?q=`** | **Buscar** ⭐ |
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
| | **POST** | **`/api/facturas/parcial`** | **Factura parcial (monto exacto)** ⭐ |
| | **POST** | **`/api/facturas/calcular-parcial`** | **Calcular parcial** ⭐ |
| | **GET** | **`/api/facturas/{id}/pdf`** | **Generar PDF** ⭐ |
| **Pagos** | POST | `/api/pagos` | Registrar pago |
| | GET | `/api/pagos` | Listar pagos |
| | GET | `/api/pagos/{id}` | Obtener pago |
| | PUT | `/api/pagos/{id}` | Actualizar pago |
| | DELETE | `/api/pagos/{id}` | Eliminar pago |
| | POST | `/api/pagos/{id}/asignar` | Asignar a cotización |
| | GET | `/api/cotizaciones/{id}/pagos` | Pagos de cotización |
| | POST | `/api/pagos/{id}/comprobante` | Subir comprobante |
| | GET | `/api/pagos/{id}/comprobante` | Obtener comprobante |
| | GET | `/api/pagos/{id}/comprobante/download` | Descargar comprobante |
| | DELETE | `/api/pagos/{id}/comprobante` | Eliminar comprobante |
| | **GET** | **`/api/pagos/sin-asignar`** | **Pagos sin asignar** ⭐ |
| | **POST** | **`/api/pagos/{id}/asignar-porcentaje`** | **Asignar por %** ⭐ |
| | **GET** | **`/api/pagos/{id}/resumen-asignaciones`** | **Resumen asignaciones** ⭐ |
| | **GET** | **`/api/pagos/{id}/calcular-montos?porcentaje=`** | **Calcular monto por %** ⭐ |
| | **GET** | **`/api/pagos/search?q=`** | **Buscar pagos** ⭐ |
| **Comprobantes NC** | GET | `/api/nc` | Listar comprobantes |
| | POST | `/api/nc` | Crear comprobante |
| | POST | `/api/nc/bulk` | Crear en bulk |
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
| **Dashboard** ⭐ | **GET** | **`/api/clientes/{id}/resumen-pagos`** | **Resumen de pagos** |
| | **GET** | **`/api/clientes/{id}/cotizaciones-pendientes`** | **Cotizaciones pendientes** |
| | **GET** | **`/api/dashboard/estados?cliente_ids=&cotizacion_ids=`** | **Estados múltiples** |
| | **POST** | **`/api/dashboard/estados/pdf`** | **PDF de estados** |
| **Configuración** ⭐ | **GET** | **`/api/config/usuario`** | **Obtener config** |
| | **PUT** | **`/api/config/usuario`** | **Actualizar config** |

**Total: ~90 endpoints** | ⭐ = Endpoints nuevos agregados (+24)

### Rutas de Almacenamiento en R2

| Tipo | Bucket | Ruta |
|------|--------|------|
| **Logo Cliente** | Público | `tenant/{tenant_id}/clientes/{id}/logo.png` |
| **Logo Tenant** | Privado | `tenant/{id}/logo.png` |
| **Comprobante Pago** | Privado | `tenant/{tenant_id}/pagos/{pago_id}/comprobante.{ext}` |
| **PDF Cotización** | Privado | `tenant/{tenant_id}/cot/{cot_id}.pdf` |
| **PDF Factura** | Privado | `tenant/{tenant_id}/fac/{fac_id}.pdf` |

---

## 📋 Referencia de Request/Response por Endpoint

### Clientes
| Endpoint | Method | Request | Response |
|----------|--------|---------|----------|
| `/api/clientes` | POST | `ClienteCreate` (JSON) | `ClienteResponse` (JSON) |
| `/api/clientes` | GET | Query params | `List[ClienteResponse]` (JSON) |
| `/api/clientes/{id}` | GET | - | `ClienteResponse` (JSON) |
| `/api/clientes/{id}` | PUT | `ClienteUpdate` (JSON) | `ClienteResponse` (JSON) |
| `/api/clientes/{id}` | DELETE | - | JSON message |
| `/api/clientes/{id}/restore` | POST | - | JSON message |
| `/api/clientes/{id}/logo` | POST | `multipart/form-data` | JSON con URL |
| `/api/clientes/{id}/logo` | GET | - | JSON con URL |
| `/api/clientes/search` | GET | Query `q=` | `List[ClienteResponse]` (JSON) |
| `/api/clientes/validate` | POST | `ClienteValidate` (JSON) | JSON `{valid, errors}` |

### Proyectos
| Endpoint | Method | Request | Response |
|----------|--------|---------|----------|
| `/api/proyectos` | POST | `ProyectoCreate` (JSON) | `ProyectoResponse` (JSON) |
| `/api/proyectos` | GET | Query params | `List[ProyectoResponse]` (JSON) |
| `/api/proyectos/{id}` | GET | - | `ProyectoResponse` (JSON) |
| `/api/proyectos/{id}` | PUT | `ProyectoUpdate` (JSON) | `ProyectoResponse` (JSON) |
| `/api/proyectos/validate` | POST | `ProyectoValidate` (JSON) | JSON `{valid, errors}` |
| `/api/proyectos/{id}/crear-cotizacion` | POST | `CotizacionCreate` (JSON) | `CotizacionResponse` (JSON) |

### Cotizaciones
| Endpoint | Method | Request | Response |
|----------|--------|---------|----------|
| `/api/cotizaciones` | POST | `CotizacionCreate` (JSON) | `CotizacionResponse` (JSON) |
| `/api/cotizaciones` | GET | Query params | `List[CotizacionResponse]` (JSON) |
| `/api/cotizaciones/{id}/full` | GET | - | `CotizacionFullResponse` (JSON) |
| `/api/cotizaciones/{id}/pdf` | GET | Query `idioma=` | `application/pdf` (bytes) |
| `/api/cotizaciones/recientes` | GET | Query `limit=` | `List[CotizacionResponse]` (JSON) |
| `/api/cotizaciones/search` | GET | Query `q=` | `List[CotizacionResponse]` (JSON) |
| `/api/cotizaciones/{id}/has-changes` | GET | - | JSON `{has_changes, message}` |

### Presupuestos
| Endpoint | Method | Request | Response |
|----------|--------|---------|----------|
| `/api/cotizaciones/{id}/presupuesto` | GET | - | `PresupuestoResponse` (JSON) |
| `/api/cotizaciones/{id}/presupuesto` | PUT | `PresupuestoData` (JSON) | `PresupuestoResponse` (JSON) |
| `/api/cotizaciones/{id}/presupuesto/calc` | GET | - | JSON con totales calculados |

### Facturas
| Endpoint | Method | Request | Response |
|----------|--------|---------|----------|
| `/api/facturas` | POST | `FacturaCreate` (JSON) | `FacturaResponse` (JSON) |
| `/api/facturas/parcial` | POST | `FacturaParcialCreate` (JSON) | `FacturaParcialResponse` (JSON) |
| `/api/facturas/calcular-parcial` | POST | `CalculoFacturaParcial` (JSON) | JSON con factor y totales |
| `/api/facturas/{id}/pdf` | GET | Query `idioma=` | `application/pdf` (bytes) |
| `/api/facturas/{id}/comprobante` | POST | `NCAsignacion` (JSON) | `FacturaResponse` (JSON) |
| `/api/facturas/{id}/full` | GET | - | `FacturaFullResponse` (JSON) |

### Pagos
| Endpoint | Method | Request | Response |
|----------|--------|---------|----------|
| `/api/pagos` | POST | `PagoRecibidoCreate` (JSON) | `PagoRecibidoResponse` (JSON) |
| `/api/pagos/{id}/comprobante` | POST | `multipart/form-data` | JSON con path |
| `/api/pagos/{id}/comprobante/download` | GET | - | Bytes (PDF/imagen) |
| `/api/pagos/sin-asignar` | GET | Query `id_cliente=` | JSON array |
| `/api/pagos/{id}/asignar-porcentaje` | POST | `AsignacionPorPorcentaje` (JSON) | `AsignacionPagoResponse` (JSON) |
| `/api/pagos/{id}/resumen-asignaciones` | GET | - | `ResumenAsignaciones` (JSON) |
| `/api/pagos/{id}/calcular-montos` | GET | Query `porcentaje=&cotizacion_id=` | JSON con monto |
| `/api/pagos/search` | GET | Query `q=` | `List[PagoRecibidoResponse]` (JSON) |

### Comprobantes (NC)
| Endpoint | Method | Request | Response |
|----------|--------|---------|----------|
| `/api/nc` | POST | `NCCreate` (JSON) | `NCResponse` (JSON) |
| `/api/nc/bulk` | POST | `NCBulkCreate` (JSON) | JSON con resumen |
| `/api/nc/{tipo}/siguiente` | GET | - | JSON con siguiente número |
| `/api/nc/tipos` | GET | - | JSON con tipos disponibles |

### Dashboard
| Endpoint | Method | Request | Response |
|----------|--------|---------|----------|
| `/api/clientes/{id}/resumen-pagos` | GET | - | `ResumenPagosCliente` (JSON) |
| `/api/clientes/{id}/cotizaciones-pendientes` | GET | - | `List[CotizacionPendiente]` (JSON) |
| `/api/dashboard/estados` | GET | Query `cliente_ids=&cotizacion_ids=` | `EstadosMultiplesResponse` (JSON) |
| `/api/dashboard/estados/pdf` | POST | JSON con IDs | `application/pdf` (bytes) |

### Configuración
| Endpoint | Method | Request | Response |
|----------|--------|---------|----------|
| `/api/config/usuario` | GET | - | `ConfigUsuarioResponse` (JSON) |
| `/api/config/usuario` | PUT | `ConfigUsuarioUpdate` (JSON) | `ConfigUsuarioResponse` (JSON) |

---

## 🔧 Tipos de Content-Type

### Requests
- **JSON:** `Content-Type: application/json`
- **Form Data:** `Content-Type: multipart/form-data`
- **Query params:** En la URL

### Responses
- **JSON:** `application/json` (mayoría)
- **PDF:** `application/pdf` (endpoints `/pdf`)
- **Imágenes:** `image/png`, `image/jpeg`, etc. (endpoints `/download`)

---

## ⚠️ Correcciones Aplicadas

### Multitenancy en Rutas de Logos (IMPORTANTE)

**Cambio aplicado:** Las rutas de logos de clientes ahora incluyen `tenant_id`

**Antes (INCORRECTO):**
```
❌ clientes/55/logo.png
   Tenant 1 y Tenant 2 → Mismo archivo (se sobrescriben)
```

**Después (CORRECTO):**
```
✅ tenant/1/clientes/55/logo.png
✅ tenant/2/clientes/55/logo.png
   Cada tenant tiene su propio logo
```

**Si tienes logos en rutas antiguas:** Necesitarás migrarlos manualmente a las nuevas rutas.

---

**Para más información, visita la documentación interactiva en `/docs` cuando el servidor esté corriendo.**

