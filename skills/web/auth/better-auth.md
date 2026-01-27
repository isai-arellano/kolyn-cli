---
check:
  required_deps: 
    - better-auth
  forbidden_deps: 
    - next-auth
    - @auth/core
---

# Better Auth Pro (Autenticación Moderna)

**Referencia Oficial:** [better-auth.com](https://better-auth.com/docs)

Better Auth es el estándar de autenticación para nuestros proyectos. Es TypeScript-first, agnóstico del framework y modular mediante plugins.

## 1. Configuración Base (`auth.ts`)

El archivo de configuración debe vivir en `lib/auth.ts` o `auth.ts` en la raíz de `src`.

### Variables de Entorno Críticas
*   `BETTER_AUTH_SECRET`: Llave de encriptación (min 32 caracteres). Generar con `openssl rand -base64 32`.
*   `BETTER_AUTH_URL`: URL base de la app (ej. `http://localhost:3000`).

### Configuración con Drizzle ORM

Recomendamos usar Drizzle ORM para la persistencia.

```typescript
import { betterAuth } from "better-auth";
import { drizzleAdapter } from "better-auth/adapters/drizzle";
import { db } from "@/db"; // Tu instancia de Drizzle

export const auth = betterAuth({
  database: drizzleAdapter(db, {
    provider: "pg", // PostgreSQL
    // Mapeo de tablas si es necesario:
    // schema: { user: "users", session: "sessions" }
  }),
  
  // 📧 Email & Password
  emailAndPassword: {
    enabled: true,
    requireEmailVerification: true, // Recomendado para prod
  },

  // 🌐 Social Providers (Opcional)
  socialProviders: {
    google: {
      clientId: process.env.GOOGLE_CLIENT_ID!,
      clientSecret: process.env.GOOGLE_CLIENT_SECRET!,
    },
  },

  // 🛡️ Seguridad
  advanced: {
    useSecureCookies: process.env.NODE_ENV === "production",
  },
});
```

---

## 2. Base de Datos (Gotchas)

**⚠️ IMPORTANTE:** Better Auth usa los nombres de **modelos** abstractos (`user`, `session`), no necesariamente los nombres de tablas de tu BD.

Comandos útiles:
*   `npx @better-auth/cli generate`: Genera el esquema para tu ORM (Drizzle/Prisma) basado en la config.
*   `npx @better-auth/cli migrate`: Aplica cambios si usas el adaptador built-in.

---

## 3. Cliente (React)

Usa el cliente optimizado para React para manejar sesiones y hooks.

```typescript
// lib/auth-client.ts
import { createAuthClient } from "better-auth/react"

export const authClient = createAuthClient({
  baseURL: "http://localhost:3000" // Opcional si usas el mismo dominio
})
```

### Uso en Componentes

```typescript
'use client'
import { authClient } from "@/lib/auth-client"

export function UserProfile() {
  const { data: session, isPending } = authClient.useSession()

  if (isPending) return <div>Cargando...</div>
  
  if (!session) return <div>No autenticado</div>

  return (
    <div>
      <h1>Hola, {session.user.name}</h1>
      <button onClick={() => authClient.signOut()}>
        Cerrar Sesión
      </button>
    </div>
  )
}
```

---

## 4. Plugins (Modularidad)

Better Auth brilla por sus plugins. No reinventes la rueda.

**Plugins Comunes:**
*   `twoFactor`: Autenticación de dos factores (TOTP).
*   `organization`: Manejo de equipos y roles (Multi-tenant).
*   `magicLink`: Login sin password vía email.
*   `username`: Permitir login con username además de email.

**Ejemplo de implementación:**

```typescript
import { twoFactor } from "better-auth/plugins/two-factor";

export const auth = betterAuth({
  // ... config base
  plugins: [
    twoFactor({
      issuer: "Mi App",
    }),
  ],
});
```

---

## 5. Middleware y Protección de Rutas

Protege tus rutas usando middleware o checks en Server Components.

```typescript
// middleware.ts
import { auth } from "@/auth"; // Tu instancia de auth
import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

export async function middleware(request: NextRequest) {
  const session = await auth.api.getSession({ headers: request.headers });
  
  if (!session) {
    return NextResponse.redirect(new URL("/login", request.url));
  }
  
  return NextResponse.next();
}

export const config = {
  matcher: ["/dashboard/:path*"],
};
```
