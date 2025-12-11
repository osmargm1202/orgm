# API de Cálculos Eléctricos - Documentación

API independiente para procesamiento de cálculos eléctricos. Recibe archivos, procesa cálculos y devuelve resultados HTML/PDF.

## Base URL

```
http://localhost:8000/api/v1
```

## Autenticación

Actualmente la API no requiere autenticación. En producción, se recomienda implementar un sistema de autenticación.

## Estructura de Respuestas

### Respuestas Exitosas

Las respuestas exitosas devuelven los datos solicitados en formato JSON.

### Respuestas de Error

```json
{
  "detail": "Mensaje de error descriptivo"
}
```

### Códigos de Estado HTTP

- `200 OK` - Operación exitosa
- `201 Created` - Recurso creado exitosamente
- `400 Bad Request` - Solicitud inválida
- `404 Not Found` - Recurso no encontrado
- `500 Internal Server Error` - Error del servidor

---

## Flujo de Trabajo

1. **Subir archivos** - Subir archivos de CIRCUITOS, DU y PANELES
2. **Iniciar procesamiento** - Enviar parámetros y comenzar cálculo
3. **Consultar estado** - Verificar progreso del procesamiento
4. **Descargar resultados** - Obtener archivos HTML/PDF generados

---

## Endpoints

### 1. Subida de Archivos

#### Subir Archivos de Circuitos

```http
POST /api/v1/proyectos/{proyecto_id}/versiones/{version}/upload/circuitos
Content-Type: multipart/form-data
```

**Path Parameters:**
- `proyecto_id` (int): ID del proyecto (identificador único)
- `version` (int): Versión del cálculo

**Body (multipart/form-data):**
- `files`: Múltiples archivos .txt de circuitos

**Ejemplo con curl:**
```bash
curl -X POST \
  "http://localhost:8000/api/v1/proyectos/1/versiones/1/upload/circuitos" \
  -F "files=@/ruta/al/archivo1.txt" \
  -F "files=@/ruta/al/archivo2.txt"
```

**Response (200):**
```json
{
  "proyecto_id": 1,
  "version": 1,
  "tipo": "circuitos",
  "archivos_subidos": [
    {
      "filename": "D - BT.txt",
      "url": "https://r2.example.com/calc-bt-input/1/1/CIRCUITOS/D - BT.txt",
      "size": 12345
    }
  ],
  "total": 1
}
```

#### Subir Archivo DU

```http
POST /api/v1/proyectos/{proyecto_id}/versiones/{version}/upload/du
Content-Type: multipart/form-data
```

**Path Parameters:**
- `proyecto_id` (int): ID del proyecto
- `version` (int): Versión del cálculo

**Body (multipart/form-data):**
- `file`: Archivo .txt DU

**Ejemplo con curl:**
```bash
curl -X POST \
  "http://localhost:8000/api/v1/proyectos/1/versiones/1/upload/du" \
  -F "file=@/ruta/al/D - DU.txt"
```

**Response (200):**
```json
{
  "proyecto_id": 1,
  "version": 1,
  "tipo": "du",
  "archivo_subido": {
    "filename": "D - DU.txt",
    "url": "https://r2.example.com/calc-bt-input/1/1/DU/D - DU.txt",
    "size": 5432
  }
}
```

#### Subir Archivo PANELES

```http
POST /api/v1/proyectos/{proyecto_id}/versiones/{version}/upload/paneles
Content-Type: multipart/form-data
```

**Path Parameters:**
- `proyecto_id` (int): ID del proyecto
- `version` (int): Versión del cálculo

**Body (multipart/form-data):**
- `file`: Archivo .csv PANELES

**Ejemplo con curl:**
```bash
curl -X POST \
  "http://localhost:8000/api/v1/proyectos/1/versiones/1/upload/paneles" \
  -F "file=@/ruta/al/PANELES.csv"
```

**Response (200):**
```json
{
  "proyecto_id": 1,
  "version": 1,
  "tipo": "paneles",
  "archivo_subido": {
    "filename": "PANELES.csv",
    "url": "https://r2.example.com/calc-bt-input/1/1/PANELES/PANELES.csv",
    "size": 8765
  }
}
```

---

### 2. Procesamiento

#### Iniciar Cálculo

```http
POST /api/v1/proyectos/{proyecto_id}/versiones/{version}/calcular
Content-Type: multipart/form-data
```

**Path Parameters:**
- `proyecto_id` (int): ID del proyecto
- `version` (int): Versión del cálculo

**Body (multipart/form-data):**

**Campos requeridos:**
- Ninguno (proyecto_id y version vienen en el path)

**Campos opcionales:**
- `cliente` (string): Nombre del cliente
- `url_logo_cliente` (string): URL del logo del cliente
- `url_logo_empresa` (string): URL del logo de la empresa
- `empresa` (string): Nombre de la empresa
- `ingeniero` (string): Nombre del ingeniero
- `codia` (string): CODI del ingeniero
- `proyecto` (string): Nombre del proyecto
- `ubicacion` (string): Ubicación del proyecto
- `blanco_negro` (boolean): Generar en blanco y negro (default: false)
- `html_only` (boolean): Solo generar HTML, no PDF (default: false)
- `eficiencia` (boolean): Aplicar eficiencia a motores y aires (default: false)
- `fp` (boolean): Aplicar factor de potencia (default: false)
- `color` (string): Color semilla hexadecimal para paleta (ej: "#1a365d")
- `alimentadores_size` (string): Tamaño de página para alimentadores (ej: "11,17")
- `cuadros_size` (string): Tamaño de página para cuadros (ej: "36,24")
- `modulos_trf_size` (string): Tamaño de página para módulos TRF (ej: "24,18")

**Ejemplo con curl:**
```bash
curl -X POST \
  "http://localhost:8000/api/v1/proyectos/1/versiones/1/calcular" \
  -F "cliente=Cliente ABC" \
  -F "empresa=ORGM" \
  -F "ingeniero=Ing. Osmar Garcia" \
  -F "codia=36467" \
  -F "proyecto=Proyecto Residencial" \
  -F "ubicacion=Santo Domingo, RD" \
  -F "blanco_negro=false" \
  -F "html_only=false" \
  -F "eficiencia=false" \
  -F "fp=false" \
  -F "color=#1a365d" \
  -F "alimentadores_size=11,17" \
  -F "cuadros_size=36,24" \
  -F "modulos_trf_size=24,18"
```

**Response (200):**
```json
{
  "proyecto_id": 1,
  "version": 1,
  "estado": "calculando",
  "mensaje": "Procesamiento iniciado"
}
```

**Nota:** El procesamiento se ejecuta de forma asíncrona. Usa el endpoint de estado para verificar el progreso.

---

### 3. Estado y Resultados

#### Consultar Estado

```http
GET /api/v1/proyectos/{proyecto_id}/versiones/{version}/estado
```

**Path Parameters:**
- `proyecto_id` (int): ID del proyecto
- `version` (int): Versión del cálculo

**Response (200):**
```json
{
  "proyecto_id": 1,
  "version": 1,
  "estado": "completado",
  "mensaje": "Procesamiento completado. 15 archivos generados.",
  "assets_generados": [
    {
      "filename": "alimentadores.pdf",
      "url": "https://r2.example.com/calc-bt/1/1/alimentadores.pdf"
    },
    {
      "filename": "cuadros.pdf",
      "url": "https://r2.example.com/calc-bt/1/1/cuadros.pdf"
    },
    {
      "filename": "assets/logo_empresa.png",
      "url": "https://r2.example.com/calc-bt/1/1/assets/logo_empresa.png"
    }
  ],
  "logs": [
    {
      "timestamp": "2025-01-01T00:00:00",
      "mensaje": "Iniciando subida de 2 archivo(s) de circuitos"
    },
    {
      "timestamp": "2025-01-01T00:00:05",
      "mensaje": "✓ Archivo subido: D - BT.txt (12345 bytes)"
    },
    {
      "timestamp": "2025-01-01T00:00:10",
      "mensaje": "Iniciando procesamiento de cálculos..."
    },
    {
      "timestamp": "2025-01-01T00:01:00",
      "mensaje": "VALIDANDO JERARQUÍA DE PANELES"
    },
    {
      "timestamp": "2025-01-01T00:01:05",
      "mensaje": "✓ Jerarquía de paneles validada correctamente"
    },
    {
      "timestamp": "2025-01-01T00:10:00",
      "mensaje": "✓ Procesamiento completado exitosamente. 15 archivos generados."
    }
  ],
  "created_at": "2025-01-01T00:00:00",
  "updated_at": "2025-01-01T00:10:00"
}
```

**Estados posibles:**
- `calculando` - El procesamiento está en curso
- `completado` - El procesamiento se completó exitosamente
- `error` - Ocurrió un error durante el procesamiento
- `subiendo` - Se están subiendo archivos

**Campo `logs`:**
El campo `logs` contiene un array de mensajes de log con timestamp. Incluye:
- Mensajes de subida de archivos
- Mensajes de procesamiento (warnings, errores, información)
- Mensajes de validación
- Mensajes de generación de archivos
- Errores si ocurren

Cada entrada de log tiene:
- `timestamp`: Fecha y hora ISO del mensaje
- `mensaje`: Texto del mensaje (sin códigos ANSI)

#### Descargar Resultados (ZIP)

```http
GET /api/v1/proyectos/{proyecto_id}/versiones/{version}/resultados
```

**Path Parameters:**
- `proyecto_id` (int): ID del proyecto
- `version` (int): Versión del cálculo

**Response:** Archivo ZIP con Content-Type `application/zip`

**Headers:**
```
Content-Disposition: attachment; filename="resultados_p1_v1.zip"
```

**Nota:** Solo disponible cuando el estado es `completado`. El ZIP contiene todos los archivos generados (HTML, PDF, TXT, assets) manteniendo la estructura de carpetas.

---

## Ejemplos de Uso

### Flujo Completo con curl

```bash
# 1. Subir archivos de circuitos
curl -X POST \
  "http://localhost:8000/api/v1/proyectos/1/versiones/1/upload/circuitos" \
  -F "files=@CIRCUITOS/D - BT.txt"

# 2. Subir archivo DU
curl -X POST \
  "http://localhost:8000/api/v1/proyectos/1/versiones/1/upload/du" \
  -F "file=@DU/D - DU.txt"

# 3. Subir archivo PANELES
curl -X POST \
  "http://localhost:8000/api/v1/proyectos/1/versiones/1/upload/paneles" \
  -F "file=@PANELES/PANELES.csv"

# 4. Iniciar procesamiento
curl -X POST \
  "http://localhost:8000/api/v1/proyectos/1/versiones/1/calcular" \
  -F "cliente=Cliente ABC" \
  -F "empresa=ORGM" \
  -F "ingeniero=Ing. Osmar Garcia" \
  -F "codia=36467" \
  -F "proyecto=Proyecto Residencial" \
  -F "ubicacion=Santo Domingo, RD" \
  -F "color=#1a365d"

# 5. Consultar estado (polling)
curl "http://localhost:8000/api/v1/proyectos/1/versiones/1/estado"

# 6. Descargar resultados cuando esté completado
curl -O "http://localhost:8000/api/v1/proyectos/1/versiones/1/resultados"
```

### JavaScript/TypeScript (Fetch API)

```javascript
const BASE_URL = 'http://localhost:8000/api/v1';
const proyectoId = 1;
const version = 1;

// 1. Subir archivos de circuitos
const uploadCircuitos = async (files) => {
  const formData = new FormData();
  files.forEach(file => formData.append('files', file));
  
  const response = await fetch(
    `${BASE_URL}/proyectos/${proyectoId}/versiones/${version}/upload/circuitos`,
    {
      method: 'POST',
      body: formData
    }
  );
  return response.json();
};

// 2. Subir archivo DU
const uploadDU = async (file) => {
  const formData = new FormData();
  formData.append('file', file);
  
  const response = await fetch(
    `${BASE_URL}/proyectos/${proyectoId}/versiones/${version}/upload/du`,
    {
      method: 'POST',
      body: formData
    }
  );
  return response.json();
};

// 3. Subir archivo PANELES
const uploadPaneles = async (file) => {
  const formData = new FormData();
  formData.append('file', file);
  
  const response = await fetch(
    `${BASE_URL}/proyectos/${proyectoId}/versiones/${version}/upload/paneles`,
    {
      method: 'POST',
      body: formData
    }
  );
  return response.json();
};

// 4. Iniciar procesamiento
const iniciarCalculo = async (params) => {
  const formData = new FormData();
  Object.entries(params).forEach(([key, value]) => {
    formData.append(key, value);
  });
  
  const response = await fetch(
    `${BASE_URL}/proyectos/${proyectoId}/versiones/${version}/calcular`,
    {
      method: 'POST',
      body: formData
    }
  );
  return response.json();
};

// 5. Consultar estado (polling)
const consultarEstado = async () => {
  const response = await fetch(
    `${BASE_URL}/proyectos/${proyectoId}/versiones/${version}/estado`
  );
  return response.json();
};

// 6. Descargar resultados
const descargarResultados = async () => {
  const response = await fetch(
    `${BASE_URL}/proyectos/${proyectoId}/versiones/${version}/resultados`
  );
  const blob = await response.blob();
  const url = window.URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `resultados_p${proyectoId}_v${version}.zip`;
  a.click();
};

// Ejemplo de uso completo
(async () => {
  // Subir archivos
  const circuitosFiles = document.getElementById('circuitos').files;
  await uploadCircuitos(Array.from(circuitosFiles));
  
  const duFile = document.getElementById('du').files[0];
  await uploadDU(duFile);
  
  const panelesFile = document.getElementById('paneles').files[0];
  await uploadPaneles(panelesFile);
  
  // Iniciar cálculo
  await iniciarCalculo({
    cliente: 'Cliente ABC',
    empresa: 'ORGM',
    ingeniero: 'Ing. Osmar Garcia',
    codia: '36467',
    proyecto: 'Proyecto Residencial',
    ubicacion: 'Santo Domingo, RD',
    color: '#1a365d'
  });
  
  // Polling de estado
  const checkStatus = async () => {
    const estado = await consultarEstado();
    
    if (estado.estado === 'calculando') {
      console.log('Calculando...', estado.mensaje);
      setTimeout(checkStatus, 2000); // Revisar cada 2 segundos
    } else if (estado.estado === 'completado') {
      console.log('Completado!', estado.assets_generados);
      // Descargar resultados
      await descargarResultados();
    } else if (estado.estado === 'error') {
      console.error('Error:', estado.mensaje);
    }
  };
  
  checkStatus();
})();
```

### Python (requests)

```python
import requests
import time

BASE_URL = "http://localhost:8000/api/v1"
proyecto_id = 1
version = 1

# 1. Subir archivos de circuitos
def upload_circuitos(files):
    url = f"{BASE_URL}/proyectos/{proyecto_id}/versiones/{version}/upload/circuitos"
    files_data = [('files', open(f, 'rb')) for f in files]
    response = requests.post(url, files=files_data)
    return response.json()

# 2. Subir archivo DU
def upload_du(file_path):
    url = f"{BASE_URL}/proyectos/{proyecto_id}/versiones/{version}/upload/du"
    with open(file_path, 'rb') as f:
        files = {'file': f}
        response = requests.post(url, files=files)
    return response.json()

# 3. Subir archivo PANELES
def upload_paneles(file_path):
    url = f"{BASE_URL}/proyectos/{proyecto_id}/versiones/{version}/upload/paneles"
    with open(file_path, 'rb') as f:
        files = {'file': f}
        response = requests.post(url, files=files)
    return response.json()

# 4. Iniciar procesamiento
def iniciar_calculo(params):
    url = f"{BASE_URL}/proyectos/{proyecto_id}/versiones/{version}/calcular"
    response = requests.post(url, data=params)
    return response.json()

# 5. Consultar estado
def consultar_estado():
    url = f"{BASE_URL}/proyectos/{proyecto_id}/versiones/{version}/estado"
    response = requests.get(url)
    return response.json()

# 6. Descargar resultados
def descargar_resultados(output_path):
    url = f"{BASE_URL}/proyectos/{proyecto_id}/versiones/{version}/resultados"
    response = requests.get(url)
    with open(output_path, 'wb') as f:
        f.write(response.content)
    return output_path

# Ejemplo de uso completo
if __name__ == "__main__":
    # Subir archivos
    print("Subiendo archivos...")
    upload_circuitos(['CIRCUITOS/D - BT.txt'])
    upload_du('DU/D - DU.txt')
    upload_paneles('PANELES/PANELES.csv')
    
    # Iniciar cálculo
    print("Iniciando cálculo...")
    params = {
        'cliente': 'Cliente ABC',
        'empresa': 'ORGM',
        'ingeniero': 'Ing. Osmar Garcia',
        'codia': '36467',
        'proyecto': 'Proyecto Residencial',
        'ubicacion': 'Santo Domingo, RD',
        'color': '#1a365d'
    }
    iniciar_calculo(params)
    
    # Polling de estado
    print("Consultando estado...")
    while True:
        estado = consultar_estado()
        
        if estado['estado'] == 'calculando':
            print(f"Calculando... {estado['mensaje']}")
            time.sleep(2)
        elif estado['estado'] == 'completado':
            print(f"Completado! {len(estado['assets_generados'])} archivos generados")
            # Descargar resultados
            descargar_resultados('resultados.zip')
            break
        elif estado['estado'] == 'error':
            print(f"Error: {estado['mensaje']}")
            break
```

---

## Estructura de Archivos en R2

### Archivos de Entrada
Los archivos subidos se guardan en:
```
calc-bt-input/{proyecto_id}/{version}/CIRCUITOS/{filename}
calc-bt-input/{proyecto_id}/{version}/DU/{filename}
calc-bt-input/{proyecto_id}/{version}/PANELES/{filename}
```

### Archivos de Salida
Los resultados generados se guardan en:
```
calc-bt/{proyecto_id}/{version}/{filename}
calc-bt/{proyecto_id}/{version}/assets/{filename}
```

---

## Notas Importantes

1. **Identificadores**: `proyecto_id` y `version` son solo identificadores. No requieren base de datos ni creación previa.

2. **Procesamiento Asíncrono**: El cálculo se ejecuta en background. Usa polling en el endpoint de estado para verificar el progreso.

3. **Archivos Requeridos**: 
   - Al menos un archivo .txt en CIRCUITOS
   - Un archivo .txt DU
   - Un archivo .csv PANELES

4. **Estados**: El estado se guarda en archivo JSON local (`estados_calculos.json`). Se pierde al reiniciar el servidor.

5. **Resultados**: Los archivos generados incluyen HTML, PDF, TXT y assets (logos/imágenes). Los assets se guardan en la carpeta `assets/`.

6. **Tamaños de Página**: Los tamaños se especifican como string con formato "ancho,alto" (ej: "11,17" para 11 pulgadas x 17 pulgadas).

7. **Colores**: El color semilla debe ser un código hexadecimal válido (ej: "#1a365d"). Se usa para generar una paleta de colores Material Design.

---

## Manejo de Errores

### Error 400 - Solicitud Inválida

```json
{
  "detail": "El archivo debe ser un archivo .txt"
}
```

### Error 404 - Recurso No Encontrado

```json
{
  "detail": "Estado no encontrado para proyecto 1, versión 1"
}
```

### Error 500 - Error del Servidor

```json
{
  "detail": "Error subiendo archivo a R2"
}
```

Siempre verifica el código de estado HTTP y el campo `detail` en la respuesta para obtener información sobre el error.

---

## Health Check

```http
GET /
```

**Response (200):**
```json
{
  "status": "healthy",
  "service": "calc-bt-api"
}
```

---

## Soporte

Para más información o soporte, consulta la documentación del proyecto o contacta al equipo de desarrollo.
