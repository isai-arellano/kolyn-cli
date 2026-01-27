---
type: scaffold
framework: nextjs
version: ">=14"
create_command: "npx create-next-app@latest {name} --typescript --tailwind --eslint --app --src-dir --no-git --import-alias '@/*'"
structure:
  directories:
    - src/lib
    - src/components/ui
    - src/hooks
    - src/types
    - src/app/(auth)
    - src/app/(dashboard)
  files:
    - path: .env.example
      content: |
        # Database
        DATABASE_URL="postgresql://user:password@localhost:5432/dbname"

        # Auth (Better Auth)
        BETTER_AUTH_SECRET="change-me-in-prod"
        BETTER_AUTH_URL="http://localhost:3000"
        
    - path: src/lib/utils.ts
      content: |
        import { clsx, type ClassValue } from "clsx"
        import { twMerge } from "tailwind-merge"

        export function cn(...inputs: ClassValue[]) {
          return twMerge(clsx(inputs))
        }
    - path: src/components/ui/.gitkeep
      content: ""
  tsconfig:
    compilerOptions:
      paths:
        "@/*": ["./src/*"]
---

# Next.js Scaffold

Estructura estándar para proyectos Next.js 14+ con App Router.

## Stack
- **Framework:** Next.js 14+ (App Router)
- **Lenguaje:** TypeScript
- **Estilos:** Tailwind CSS
- **Componentes:** Shadcn UI (compatible)

## Estructura de Carpetas

- `src/app/(auth)`: Rutas de autenticación (login, register).
- `src/app/(dashboard)`: Rutas protegidas del panel principal.
- `src/components/ui`: Componentes base (botones, inputs).
- `src/lib`: Utilidades y configuraciones (db, auth client).
- `src/hooks`: Hooks personalizados de React.
