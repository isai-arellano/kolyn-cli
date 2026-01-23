# Kolyn CLI

Herramienta CLI simple que ayuda a agentes IA con contexto del proyecto y acceso a skills.

## Comandos

```
kolyn init           Inicializa kolyn y agrega contexto al Agent.md
kolyn skills         Retorna JSON con skills disponibles para la IA
kolyn skills list    Lista skills y permite ver/editar contenido
kolyn skills paths   Retorna solo las rutas de skills
```

## Instalación

### Instalación rápida (Linux/Mac)

```bash
curl -sfL https://raw.githubusercontent.com/isai-arellano/kolyn-cli/main/install.sh | sh
```

### Manual

Descarga el binario desde [Releases](https://github.com/isai-arellano/kolyn-cli/releases) para tu sistema operativo.

### Desde source

```bash
go install github.com/isai-arellano/kolyn-cli@latest
```

## Desinstalación

Para desinstalar Kolyn y limpiar sus archivos de configuración:

```bash
# Si instalaste usando el script
curl -sfL https://raw.githubusercontent.com/isai-arellano/kolyn-cli/main/uninstall.sh | sh

# O si tienes el repo clonado:
./uninstall.sh
```

## Uso

### Inicializar proyecto

```bash
cd tu-proyecto
kolyn init
```

Agrega contexto de kolyn al `Agent.md` para que la IA cómo usar los comandos sepa.

### Obtener skills para la IA

```bash
kolyn skills
```

Retorna JSON con todas las skills disponibles:

```json
{
  "total_skills": 3,
  "skills_dir": "/Users/tu-usuario/.kolyn/skills",
  "skills": [
    {
      "name": "commits_y_convenciones",
      "category": "general",
      "path": "/Users/tu-usuario/.kolyn/skills/general/commits_y_convenciones.md",
      "description": "Conventional Commits simplificado"
    }
  ]
}
```

### Modo interactivo

```bash
kolyn skills list
```

Lista todas las skills y permite ver o editar su contenido con tu editor default.

## Skills

Las skills se guardan en `~/.kolyn/skills/` organizadas por categoría:

```
~/.kolyn/skills/
├── general/
│   └── commits_y_convenciones.md
└── web/
    ├── backend_routehandlers.md
    ├── database_drizzle.md
    ├── devops_dokploy.md
    └── frontend_ui.md
```

## License

MIT
