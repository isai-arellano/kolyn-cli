# Kolyn CLI 🚀

**Orquestador de Desarrollo para la Era de la IA**

Kolyn es una herramienta CLI diseñada para estandarizar flujos de trabajo en equipos modernos. Actúa como un puente entre desarrolladores y Agentes de IA, inyectando contexto (Skills, Reglas, Roles) y automatizando tareas repetitivas de infraestructura.

## 📦 Instalación

### Instalación Rápida

**Mac / Linux:**
```bash
curl -sfL https://raw.githubusercontent.com/isai-arellano/kolyn-cli/main/install.sh | sh
```

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/isai-arellano/kolyn-cli/main/install.ps1 | iex
```

### Desde Source (Go)
Si tienes Go instalado:
```bash
go install github.com/isai-arellano/kolyn-cli@latest
```

## 🚀 Getting Started

### 1. Inicializar Proyecto
Al iniciar un proyecto, Kolyn crea o actualiza el archivo `Agent.md`. Este archivo es el "cerebro" que tu Agente de IA leerá para entender cómo trabajar contigo.

```bash
cd mi-proyecto
kolyn init
```

### 2. Sincronizar Estándares del Equipo (Sync)
Kolyn permite que todo tu equipo comparta las mismas "Skills" (guías de estilo, arquitecturas, roles). Crea un archivo `.kolyn.json` en la raíz de tu proyecto:

```json
{
  "project_name": "ecommerce-platform",
  "skills_sources": [
    "https://github.com/mi-org/backend-standards"
  ]
}
```

Luego ejecuta:
```bash
kolyn sync
```
Esto descargará automáticamente las skills de tu equipo en `~/.kolyn/sources/` y las hará disponibles para la IA.

## 🛠 Herramientas (Tools)

Kolyn incluye un set de navajas suizas para tareas comunes.

### 🐳 Docker Tools
Levanta infraestructura de desarrollo en segundos sin escribir `docker-compose.yaml` manualmente.

```bash
# Levantar un servicio (menú interactivo)
kolyn tools docker up

# Listar servicios corriendo
kolyn tools docker list

# Detener un servicio
kolyn tools docker down
```
*Servicios disponibles:* n8n, PostgreSQL, Redis, MongoDB, Next.js Stack, entre otros.
*Ubicación de datos:* Los volúmenes y archivos persisten en `~/.kolyn/services/`.

### 🔑 SSH Manager
Genera llaves SSH modernas (Ed25519) y configura tu archivo `~/.ssh/config` automáticamente con una sola línea.

```bash
# Sintaxis: kolyn tools ssh create <nombre> <ip> [usuario]
kolyn tools ssh create mi-servidor 192.168.1.50 root
```
Esto:
1. Genera llaves en `~/.ssh/mi-servidor`
2. Agrega la configuración al `config` de SSH.
3. (Opcional) Copia la llave pública al servidor remoto.

## 🧠 Comandos de IA (Skills)

Comandos pensados para que los use tu Agente de IA (Windsurf, Cursor, Cline, etc):

*   `kolyn skills paths`: Muestra dónde están los archivos Markdown de contexto (Roles, Reglas, Tech).
*   `kolyn skills list`: Explorador interactivo de skills para humanos.

## 🗑 Desinstalación

Si decides irte, Kolyn limpia su desorden. El script te preguntará si quieres conservar tus Skills descargadas.

**Mac / Linux:**
```bash
curl -sfL https://raw.githubusercontent.com/isai-arellano/kolyn-cli/main/uninstall.sh | sh
```

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/isai-arellano/kolyn-cli/main/uninstall.ps1 | iex
```

## 📂 Estructura de Archivos

Kolyn mantiene tu sistema ordenado guardando todo en `~/.kolyn`:

```text
~/.kolyn/
├── services/       # Contenedores Docker y sus volúmenes
├── skills/         # Skills locales
└── sources/        # Skills sincronizadas desde Git (Sync)
```

## License
MIT
