# ORGM - Changelog

Descripción general: Herramienta CLI en Go para automatizar flujos de trabajo (Docker, administración y utilidades varias) soportada por configuración mediante `viper` y menús interactivos.

## 2025-01-11
- **Eliminación de Docker**: Removidas todas las funciones de Docker del proyecto
- **Simplificación de comandos de configuración**:
  - `orgm keys [editor]`: Edita `keys.toml` localmente (sin descarga/upload automático)
  - `orgm links [editor]`: Edita `links.toml` localmente (sin descarga/upload automático)  
  - `orgm config [editor]`: Edita `config.toml` localmente (sin descarga/upload automático)
  - Descarga/upload manual con `init` y `update` respectivamente
- **Nuevo comando `orgm viper`**: Muestra todas las variables de configuración cargadas
  - Visualización organizada por categorías
  - Detección automática de archivos config.toml, links.toml y keys.toml
  - Valores formateados para fácil lectura
- **Carga automática de archivos TOML**: Viper ahora carga automáticamente links.toml y keys.toml
- **Variables de entorno**: Uso de `BUCKET_URL` y `BUCKET_TOKEN` para sincronización R2
- Nuevo comando `orgm prop` para gestión completa de propuestas con API:
  - Crear nuevas propuestas con textarea para prompts largos
  - **Workflow completo asíncrono**: Proceso completo (texto → HTML → PDF) con monitoreo en tiempo real
  - Modificar propuestas existentes con filtrado por título/ID
  - Listar propuestas ordenadas por fecha (más recientes primero)
  - **Búsqueda de propuestas**: Por términos en título y subtítulo
  - Descargar propuestas (MD, HTML, PDF) a ~/Downloads
  - **Health check de API**: Verificación del estado de la API
  - Apertura automática de carpeta de descarga
  - Integración con `links.toml` para URL de API
  - **Modelo personalizable**: Soporte para diferentes modelos de IA
  - **Soporte portapapeles**: Pegar contenido con Ctrl+V en textarea
  - **Endpoints corregidos**: Descarga usando `/proposal/{id}/{file_type}`
- Nuevo input `TextArea` en `inputs/textarea.go` para prompts multilínea
- Soporte para filtrado de propuestas con "/" en listados

## 2025-09-11
- Prefijo del directorio HOME a `viper.carpetas.apps` en `cmd/docker.go` cuando la ruta no es absoluta y expansión de `~` para asegurar que se encuentre el archivo `docker/config.toml` correctamente.
- Se eliminaron `HostPort`, `AppPort`, `DockerVolumeHostPath` y `DockerVolumeName` de `ProjectConfig`; ahora puertos y volúmenes se definen exclusivamente en `docker-compose`.
- Nuevos comandos `orgm docker nvim` y `orgm docker nano` para editar el archivo de configuración `docker/config.toml` directamente.
- Si no existe una sección para el proyecto actual en `docker/config.toml`, ahora se solicita interactivamente `HOST_NAME`, `DOCKER_IMAGE_NAME`, `DOCKER_IMAGE_TAG`, `DOCKER_SAVE_FILE` y `DOCKER_CONTAINER_NAME` con sugerencias por defecto, y se agrega al final del archivo.
- Integración inicial con Cloudflare R2: sincronización de `docker/config.toml` desde nube al directorio de configuración y preferencia por la ruta `~/.config/orgm/docker/config.toml`.
- `orgm keys` ahora soporta: `init` (descargar), `update` (subir) y `[editor]` dinámico (descarga, edita, sube) para `keys.toml`.
- `orgm docker` agrega `init` y `update` para sincronizar `docker/config.toml`; `docker nvim|nano` edita el archivo local en `viper.config_path`.
- Nuevo comando `orgm links` con `init` y `update` para sincronizar `links.toml` en la carpeta de configuración.

