
## WINDOWS

### Download

```
New-Item -Path "$env:USERPROFILE\.config\orgm" -ItemType Directory -Force

powershell -Command "Invoke-WebRequest -Uri https://github.com/osmargm1202/orgm/releases/latest/download/orgm.exe -OutFile '$env:USERPROFILE\.config\orgm\orgm.exe'"
```


## LINUX

### Download

```
curl -L https://github.com/osmargm1202/orgm/releases/latest/download/orgm -o ~/.local/bin/orgm
```

### Set executable permissions

```
chmod +x ~/.local/bin/orgm
```

### Set path

```
echo 'export PATH=$PATH:$HOME/.local/bin' >> ~/.bashrc
source ~/.bashrc

echo 'export PATH=$PATH:$HOME/.local/bin' >> ~/.zshrc
source ~/.zshrc

echo 'export PATH=$PATH:$HOME/.local/bin' >> ~/.profile
source ~/.profile

echo 'export PATH=$PATH:$HOME/.local/bin' >> ~/.bash_profile
source ~/.bash_profile
```

## UPDATE

To update the application to the latest version:

```
orgm update
```

This command will download the latest binary from GitHub and replace the current one.


## API Endpoints Disponibles

La aplicación ORGM proporciona acceso a los siguientes endpoints de API:

### Endpoints Principales

| Endpoint | Descripción | Uso |
|----------|-------------|-----|
| `/cot` | Gestión de cotizaciones | Crear, consultar y gestionar cotizaciones |
| `/fac` | Gestión de facturas | Crear, consultar y gestionar facturas |
| `/ai` | Servicios de Inteligencia Artificial | Interacciones con modelos GPT |

### Comandos Disponibles

La aplicación incluye los siguientes comandos principales:

#### Administración (`adm`)
- **Clientes**: Gestión de clientes
- **Proyectos**: Gestión de proyectos
- **Cotizaciones**: Crear y gestionar cotizaciones
- **Facturas**: Crear y gestionar facturas
- **Folders**: Gestión de carpetas de proyecto
- **Locations**: Gestión de ubicaciones
- **PDF**: Operaciones con documentos PDF
- **Presentaciones**: Gestión de presentaciones

#### Inteligencia Artificial (`ai`)
- **Conversaciones**: Interacción con modelos GPT
- **Gestión de historial**: Guardar y cargar conversaciones
- **Exportación**: Exportar conversaciones a TXT
- **Configuraciones**: Múltiples configuraciones de AI

#### Propuestas (`prop`)
- **Crear propuestas**: Generar nuevas propuestas con textarea para prompts largos, título y subtítulo personalizables
- **Workflow completo**: Proceso asíncrono completo (texto → HTML → PDF) con monitoreo en tiempo real
- **Modificar propuestas**: Editar propuestas existentes con filtrado por título/ID usando "/"
- **Listar propuestas**: Ver todas las propuestas ordenadas por fecha (más recientes primero)
- **Buscar propuestas**: Búsqueda por términos en título y subtítulo
- **Descargar propuestas**: Descargar archivos MD, HTML y PDF a ~/Downloads con apertura automática de carpeta
- **Verificar API**: Health check del estado de la API de propuestas
- **Generación automática**: Opción de generar HTML y PDF después de crear propuesta
- **Filtrado inteligente**: Buscar propuestas por título o ID con opción "/ Filtrar propuestas"
- **Integración API**: Conecta con API de propuestas configurada en `links.toml`
- **Modelo personalizable**: Soporte para diferentes modelos de IA (por defecto `gpt-4o-latest`)
- **Portapapeles**: Soporte para pegar contenido desde el portapapeles con Ctrl+V en textarea

#### Gestión de Archivos de Configuración
- **Keys**: Gestión de `keys.toml` (init/update/edit con editor)
- **Links**: Gestión de `links.toml` (init/update/edit con editor)  
- **Config**: Gestión de `config.toml` (init/update/edit con editor)
- **Viper**: Muestra todas las variables de configuración cargadas desde los archivos TOML
- **Sincronización R2**: Descarga y subida automática con variables `BUCKET_URL` y `BUCKET_TOKEN`

#### Otras Funcionalidades
- **Check**: Verificación de conectividad a servicios
- **Nextcloud**: Integración con Nextcloud
- **Build**: Construcción y despliegue de la aplicación

### Verificación de Conectividad

Para verificar que todos los servicios estén funcionando correctamente:

```bash
orgm check
```

Este comando verificará:
- Conectividad a los endpoints de API
- Conexión a Nextcloud
- Conexión a base de datos PostgREST

### Configuración

La aplicación requiere un archivo de configuración TOML con las siguientes secciones:
- `url.apis`: URL de los servicios de API
- `url.postgrest`: URL de PostgREST
- `nextcloud.*`: Configuración de Nextcloud
- `cloudflare.*`: Credenciales de Cloudflare Access

### Gestión de Propuestas

Para usar el comando `orgm prop`, necesitas configurar la URL de la API de propuestas en `links.toml`:

```bash
# Descargar configuración de links
orgm links init

# Editar la configuración
orgm links nano
```

En el archivo `links.toml`, configura:
```toml
[links]
propuestas_api = "http://localhost:8000"  # URL de tu API de propuestas
```

#### Uso del comando prop:
```bash
# Ejecutar el comando principal
orgm prop

# El comando te guiará a través de:
# 1. Crear nueva propuesta (con textarea para prompts largos)
# 2. Workflow completo (Background Tasks) - Proceso asíncrono completo
# 3. Modificar propuesta existente (con filtrado)
# 4. Listar todas las propuestas (ordenadas por fecha)
# 5. Buscar propuestas (por términos en título/subtítulo)
# 6. Descargar propuestas (MD, HTML, PDF a ~/Downloads)
# 7. Verificar estado de API (health check)
```

#### Características del Workflow Completo:
- **Proceso asíncrono**: Genera texto, HTML y PDF en background
- **Monitoreo en tiempo real**: Muestra progreso y estado de la tarea
- **Modelo personalizable**: Permite seleccionar el modelo de IA
- **Timeout inteligente**: Espera hasta 10 minutos para completar
- **Resultado completo**: Obtiene Proposal ID y PDF URL al finalizar

### Visualización de Configuración

Para ver todas las variables de configuración cargadas desde los archivos TOML:

```bash
# Mostrar todas las variables de configuración
orgm viper

# El comando mostrará:
# - Archivos de configuración encontrados (config.toml, links.toml, keys.toml)
# - Todas las variables organizadas por categoría
# - Valores formateados para fácil lectura
```
