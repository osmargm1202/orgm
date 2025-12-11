# Worker CLI - Documentación de Uso

## Descripción

Worker de Cloudflare Workers que proporciona acceso a valores almacenados en KV mediante una API REST simple. El worker permite consultar configuraciones y valores mediante una llave (key) como parámetro.

## URL Pública

```
https://cli-config.or-gm.com/?key=<nombre_llave>
```

## Uso

### Endpoint

**GET** `https://cli-config.or-gm.com/?key=<nombre_llave>`

### Parámetros

- `key` (requerido): Nombre de la llave que se desea consultar en KV

### Ejemplo de Uso

```bash
# Consultar una llave específica
curl "https://cli-config.or-gm.com/?key=api_calc_bt"

# Desde el navegador
https://cli-config.or-gm.com/?key=api_calc_bt
```

## Respuestas

### Éxito (200 OK)

Cuando la llave existe y se encuentra el valor:

```json
{
  "success": true,
  "key": "api_calc_bt",
  "value": "valor_almacenado"
}
```

### Error: Llave faltante (400 Bad Request)

Cuando no se proporciona el parámetro `key`:

```json
{
  "success": false,
  "error": "Missing key parameter",
  "message": "Usage: ?key=your_key_name"
}
```

### Error: Llave no encontrada (404 Not Found)

Cuando la llave no existe en KV:

```json
{
  "success": false,
  "error": "Key not found",
  "key": "api_calc_bt",
  "message": "The key \"api_calc_bt\" does not exist in the configuration"
}
```

### Error: Error interno (500 Internal Server Error)

Cuando ocurre un error en el servidor:

```json
{
  "success": false,
  "error": "Internal server error",
  "message": "Error message details"
}
```

## CORS

El worker incluye headers CORS que permiten requests desde cualquier origen:

- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: GET, OPTIONS`
- `Access-Control-Allow-Headers: Content-Type`

Esto permite que aplicaciones web y CLIs puedan consumir la API sin problemas de CORS.

## Gestión de Llaves KV

### Agregar una llave

```bash
# Usando el binding
wrangler kv:key put "nombre_llave" "valor" --binding=CLI_CONFIG

# Ejemplo
wrangler kv:key put "api_calc_bt" "sk-1234567890" --binding=CLI_CONFIG
```

### Listar todas las llaves

```bash
wrangler kv:key list --binding=CLI_CONFIG
```

### Obtener el valor de una llave

```bash
wrangler kv:key get "api_calc_bt" --binding=CLI_CONFIG
```

### Eliminar una llave

```bash
wrangler kv:key delete "api_calc_bt" --binding=CLI_CONFIG
```

## Desarrollo Local

### Iniciar servidor de desarrollo

```bash
npm run dev
# o
wrangler dev
```

El worker estará disponible en `http://localhost:8787` (o el puerto indicado).

### Probar localmente

```bash
curl "http://localhost:8787?key=api_calc_bt"
```

**Nota**: Para desarrollo local, asegúrate de agregar las llaves también al namespace de preview:

```bash
wrangler kv:key put "api_calc_bt" "valor" --binding=CLI_CONFIG --preview
```

## Deploy

```bash
npm run deploy
# o
wrangler deploy
```

## Ejemplos de Integración

### JavaScript/TypeScript

```javascript
async function getConfig(key) {
  const response = await fetch(`https://cli-config.or-gm.com/?key=${key}`);
  const data = await response.json();
  
  if (data.success) {
    return data.value;
  } else {
    throw new Error(data.message);
  }
}

// Uso
const apiKey = await getConfig('api_calc_bt');
console.log(apiKey);
```

### Python

```python
import requests

def get_config(key):
    url = f"https://cli-config.or-gm.com/?key={key}"
    response = requests.get(url)
    data = response.json()
    
    if data['success']:
        return data['value']
    else:
        raise Exception(data['message'])

# Uso
api_key = get_config('api_calc_bt')
print(api_key)
```

### cURL

```bash
# Obtener valor
curl "https://cli-config.or-gm.com/?key=api_calc_bt"

# Con formato JSON bonito (requiere jq)
curl "https://cli-config.or-gm.com/?key=api_calc_bt" | jq
```

## Estructura del Proyecto

```
worker-cli/
├── src/
│   └── index.js          # Código del worker
├── wrangler.jsonc        # Configuración de Wrangler
├── package.json          # Dependencias
└── worker.md            # Esta documentación
```

## Configuración KV

- **Namespace**: `cli`
- **Binding**: `CLI_CONFIG`
- **ID**: `fbe316f8b73e4f4ca5da41b2e3cb0488`

