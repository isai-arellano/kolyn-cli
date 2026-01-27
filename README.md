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
irm https://raw.githubusercontent.com/isai-arellano/kolyn-cli/main/uninstall.ps1 | iex
```

### Desde Source (Go)
Si tienes Go instalado:
```bash
go install github.com/isai-arellano/kolyn-cli@latest
```

### ⚙️ Configuración Global (Zero Config)
Kolyn usa una configuración centralizada para que no tengas que repetir tus preferencias en cada proyecto.

```bash
# Inicia el asistente de configuración
kolyn config init
```

Esto te permitirá definir:
1. **Idioma Preferido:** Español (México) o Inglés.
2. **Repositorio de Skills:** Define una fuente única de verdad para tu equipo (ej. `tu-org/skills`).
3. **Preferencias:** Almacenadas en `~/.kolyn/config.json`.

---

## 🚀 Flujo de Trabajo (Workflow)

### 1. Inicializar Proyecto
Al iniciar un proyecto, Kolyn crea o actualiza el archivo `Agent.md`. Este archivo es el "cerebro" que tu Agente de IA leerá para entender cómo trabajar contigo.

```bash
cd mi-proyecto
kolyn init
```

### 2. Sincronizar Skills (Sync)
Kolyn inyecta conocimiento técnico estandarizado a tu IA.

```bash
kolyn sync
```
*   Si configuraste un repo global, descargará las skills desde ahí.
*   Si el proyecto tiene un `.kolyn.json` específico, usará esa configuración.
*   Soporta **repositorios privados** (vía SSH/HTTPS).

### 3. Auditar Proyecto (Check)
Verifica que tu proyecto cumpla con los estándares definidos en tus skills.

```bash
kolyn check
```
Esta herramienta lee los archivos markdown de tus skills y busca reglas definidas en el frontmatter (archivos requeridos, dependencias prohibidas, etc).

---

## 🛠 Herramientas (Tools)

Kolyn incluye un set de navajas suizas para tareas comunes.

### 🐳 Docker Tools
Levanta infraestructura de desarrollo en segundos usando templates pre-configurados.

```bash
# Levantar un servicio (menú interactivo)
kolyn up
# O usar el alias:
kolyn docker up

# Listar servicios corriendo
kolyn tools docker list

# Detener un servicio
kolyn tools docker down
```

**Personalización:**
Kolyn busca templates `.yml` en `~/.kolyn/templates/`.
Puedes agregar tus propios archivos ahí y aparecerán automáticamente en el menú.

*Templates incluidos por defecto:* n8n, PostgreSQL, Redis, MongoDB.
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

---

## 🧠 Comandos de IA (Skills)

Comandos pensados para que los use tu Agente de IA (Windsurf, Cursor, Cline, etc):

*   `kolyn skills paths`: Muestra rutas absolutas a los archivos de contexto (Roles, Reglas, Tech).
*   `kolyn skills list`: Explorador interactivo de skills para humanos.

### Estructura Recomendada de Skills
Kolyn sugiere organizar tu repositorio de skills de la siguiente manera:

```text
skills/
├── backend/
│   ├── go/
│   └── python/
├── web/
│   ├── framework/ (nextjs, react)
│   ├── ui/        (shadcn, tailwind)
│   └── data/      (drizzle, prisma)
├── mobile/
└── devops/
```

---

## 🗑 Desinstalación

Si decides irte, Kolyn limpia su desorden.

```bash
kolyn uninstall
```
O manualmente:
```bash
# Mac / Linux
curl -sfL https://raw.githubusercontent.com/isai-arellano/kolyn-cli/main/uninstall.sh | sh
```

## 📂 Estructura de Archivos

Kolyn mantiene tu sistema ordenado guardando todo en `~/.kolyn`:

```text
~/.kolyn/
├── config.json     # Configuración global (Idioma, Repo Default)
├── services/       # Contenedores Docker y sus volúmenes
├── templates/      # Templates .yml para docker up (Editable)
├── skills/         # Skills locales descargadas
└── sources/        # Repositorios clonados (Cache)
```

## License
MIT
