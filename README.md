# Kolyn CLI 🚀

**Orquestador de Desarrollo para la Era de la IA**

Kolyn es una herramienta CLI diseñada para estandarizar flujos de trabajo en equipos modernos. Actúa como un puente entre desarrolladores y Agentes de IA (Windsurf, Cursor, Cline), inyectando contexto técnico (Skills, Reglas, Arquitectura) y automatizando tareas repetitivas.

---

## 🧠 Arquitectura: Cerebro y Músculo

Kolyn separa la lógica de la herramienta del conocimiento técnico.

1.  **El Músculo (Kolyn CLI):** Binario que instalas en tu máquina. Sabe cómo auditar código, levantar Docker y generar archivos.
2.  **El Conocimiento (Skills Repo):** Un repositorio Git (privado o público) donde tu equipo define *cómo* se hacen las cosas (Reglas de Linting, Stack Tecnológico, Convenciones).
3.  **El Cerebro del Proyecto (Agent.md):** Un archivo generado en la raíz de cada proyecto que le dice a la IA exactamente qué herramientas y reglas aplican para *ese* proyecto específico.

---

## 📦 Instalación

### Mac / Linux
```bash
curl -sfL https://raw.githubusercontent.com/isai-arellano/kolyn-cli/main/install.sh | sh
```

### Windows (PowerShell)
```powershell
irm https://raw.githubusercontent.com/isai-arellano/kolyn-cli/main/uninstall.ps1 | iex
```

### ⚙️ Configuración Inicial (Zero Config)
La primera vez que uses Kolyn, ejecuta esto para conectarlo con el "cerebro" de tu equipo (repositorio de skills):

```bash
kolyn config init
```
*Te pedirá idioma y la URL del repo de skills (ej. `git@github.com:tu-org/skills.git`).*

---

## 🚀 Flujo de Trabajo (Workflow)

### 1. Crear Nuevo Proyecto (Scaffold)
Crea proyectos desde cero siguiendo las mejores prácticas de tu equipo.

```bash
kolyn scaffold
```
1. Seleccionas el tipo de proyecto (ej. Next.js).
2. Kolyn genera la estructura de carpetas y archivos base.
3. **Automáticamente** inicia la configuración de contexto (`Agent.md`).

### 2. Inicializar Proyecto Existente
Si ya tienes código, genera el contexto para tu IA:

```bash
kolyn init
```
Kolyn detectará tu stack (Next.js, Go, Python) y te hará preguntas clave:
*   *¿Usas base de datos?*
*   *¿Tienes autenticación?*
*   *¿Consumes APIs externas?*

El resultado es un archivo `Agent.md` optimizado que tu IA leerá para entender el proyecto.

### 3. Auditar (Check)
Verifica que tu código cumpla con las reglas definidas en tus skills.

```bash
kolyn check
```
Kolyn lee el `Agent.md`, ve qué "Capabilities" (capacidades) activaste (ej. Database, Auth) y audita solo lo necesario.
*   ✅ Verifica dependencias requeridas (ej. `drizzle-orm`).
*   ✅ Verifica archivos de configuración (ej. `drizzle.config.ts`).
*   ❌ Alerta sobre dependencias prohibidas.

---

## 🧩 Conceptos Clave

### Capabilities (Capacidades)
En lugar de validar todo contra todo, Kolyn usa "Capabilities" para entender qué hace tu proyecto:

| Capability | Descripción | Skills que activa |
|------------|-------------|-------------------|
| `core` | Estructura base del framework | Linting, Config básica |
| `ui` | Componentes visuales | Shadcn/UI, Tailwind, Iconos |
| `database` | Persistencia de datos | ORMs (Drizzle, Prisma), Drivers |
| `auth` | Usuarios y Sesiones | Better Auth, NextAuth |
| `api` | Consumo de servicios | Axios, React Query, Zod |
| `devops` | CI/CD y Deploy | GitHub Actions, Dockerfiles |

### Skills
Archivos Markdown que viven en tu repositorio y definen las reglas. Ejemplo de frontmatter:

```yaml
---
name: Drizzle ORM
applies_to: [nextjs, node]
capability: database
check:
  required_deps: [drizzle-orm]
  files_exist_any: [drizzle.config.ts]
---
# Drizzle ORM Guidelines...
```

---

## 🛠 Herramientas (Tools)

### 🐳 Docker Manager
Levanta servicios de infraestructura (BDs, Cache) en segundos.

```bash
kolyn up           # Menú interactivo para levantar servicios
kolyn status       # Ver qué está corriendo
kolyn down         # Apagar todo
```
*Templates incluidos:* PostgreSQL, Redis, MongoDB, n8n, Supabase.

### 🔑 SSH Manager
Genera llaves SSH modernas y configura tu `~/.ssh/config` automáticamente.

```bash
kolyn tools ssh create mi-servidor 192.168.1.50 root
```

---

## 📂 Estructura de Archivos

```text
~/.kolyn/
├── config.json     # Configuración global
├── sources/        # Repositorios de skills clonados (Cache)
├── services/       # Volúmenes de Docker persistentes
└── templates/      # Tus archivos docker-compose.yml personalizados
```

## License
MIT
