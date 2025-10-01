
# 📦 Instalación

## 🐧 Linux / macOS

Instala ORGM con un solo comando:

```bash
curl -fsSL custom.or-gm.com/cli.sh | bash
```

## 🪟 Windows

Descarga e instala desde PowerShell o Command Prompt:

```powershell
# PowerShell
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/osmargm1202/orgm/master/install.bat" -OutFile "install.bat" && .\install.bat && del install.bat

# Command Prompt  
curl -O https://raw.githubusercontent.com/osmargm1202/orgm/master/install.bat && install.bat && del install.bat
```

### Instalación Manual

Si prefieres instalar manualmente:

#### Linux
```bash
mkdir -p ~/.local/bin
curl -L https://raw.githubusercontent.com/osmargm1202/orgm/master/orgm -o ~/.local/bin/orgm
chmod +x ~/.local/bin/orgm
```

#### Windows
```powershell
mkdir "$env:USERPROFILE\.config\orgm" -Force
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/osmargm1202/orgm/master/orgm.exe" -OutFile "$env:USERPROFILE\.config\orgm\orgm.exe"
```

# 🔄 Actualización

Para actualizar ORGM a la última versión:

```bash
orgm update
```

Este comando descarga automáticamente el instalador más reciente y actualiza el binario. El proceso es completamente automático y funciona tanto en Linux como en Windows.

# 🖥️ Interfaz Gráfica (yad)

Para usar la interfaz gráfica de propuestas (`orgm prop`), necesitas instalar `yad`:

## 🐧 Linux (Arch/Manjaro)
```bash
sudo pacman -S yad
```

## 🐧 Linux (Ubuntu/Debian)
```bash
sudo apt install yad
```

## 🐧 Linux (Fedora/RHEL)
```bash
sudo dnf install yad
```

## 🍎 macOS
```bash
brew install yad
```

Sin `yad` instalado, los comandos `orgm prop` y `orgm prop new` mostrarán errores. Los comandos de línea de comandos (`orgm prop gen`, `orgm prop find`, etc.) funcionan sin interfaz gráfica.


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
- **Interfaz gráfica con yad**: Gestión completa de propuestas con interfaz gráfica usando `yad`
- **Comando principal**: `orgm prop` muestra ayuda con subcomandos disponibles
- **Nueva propuesta**: `orgm prop new` para crear propuestas con interfaz gráfica
- **Modificar propuestas**: `orgm prop mod` para gestionar propuestas existentes con menú completo
- **Ver propuestas**: `orgm prop view` para descargar y ver archivos de propuestas
- **Lista con búsqueda**: Muestra todas las propuestas con búsqueda integrada por título
- **Menú de gestión**: Acciones disponibles para cada propuesta (modificar, ver, generar HTML/PDF, descargar)
- **Descarga automática**: Archivos se guardan en ~/Downloads y se abren automáticamente
  - **Integración API**: Conecta con API de propuestas configurada en `config.toml`
- **Modelo personalizable**: Soporte para diferentes modelos de IA (por defecto `gpt-5-chat-latest`)

#### Gestión de Archivos de Configuración
- **Config**: Gestión de `config.toml` (init/update/edit con editor)
- **Viper**: Muestra todas las variables de configuración cargadas desde el archivo TOML
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

Para usar el comando `orgm prop`, necesitas configurar la URL de la API de propuestas en `config.toml`:

```bash
# Editar la configuración
orgm config nano
```

En el archivo `config.toml`, configura:
```toml
[url]
propuestas_api = "http://localhost:8000"  # URL de tu API de propuestas
```

#### Uso del comando prop:
```bash
# Mostrar ayuda con subcomandos disponibles
orgm prop

# Crear nueva propuesta con interfaz gráfica
orgm prop new

# Modificar propuesta existente con interfaz gráfica
orgm prop mod

# Ver y descargar propuestas con interfaz gráfica
orgm prop view
```

#### Características de la Interfaz Gráfica:
- **Comando principal**: `orgm prop` muestra ayuda con todos los subcomandos disponibles
- **Nueva propuesta**: `orgm prop new` crea propuestas con cuadro de diálogo completo
- **Gestión completa**: `orgm prop mod` permite modificar, ver, generar HTML/PDF y descargar
- **Visualización**: `orgm prop view` descarga todos los archivos y permite abrirlos
- **Lista con búsqueda**: Muestra todas las propuestas con búsqueda integrada por título
- **Menú contextual**: Acciones disponibles para cada propuesta seleccionada
- **Descarga automática**: Archivos se guardan en ~/Downloads
- **Apertura automática**: Archivos se abren automáticamente en el sistema
- **Navegación intuitiva**: Interfaz fácil de usar con botones y menús claros

### Visualización de Configuración

Para ver todas las variables de configuración cargadas desde los archivos TOML:

```bash
# Mostrar todas las variables de configuración
orgm viper

# El comando mostrará:
# - Archivos de configuración encontrados (config.toml)
# - Todas las variables organizadas por categoría
# - Valores formateados para fácil lectura
```
