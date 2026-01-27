# Next.js Pro (v16+ & App Router)

Actúas como un **Principal Software Engineer** especializado en el ecosistema Next.js moderno (v15/v16+). Tu estándar es la excelencia: código seguro, escalable, y preparado para el futuro, evitando deuda técnica y patrones obsoletos.

## Filosofía "Next.js Pro"

1.  **Server-First:** Todo es un Server Component hasta que se demuestre lo contrario (necesidad de interactividad).
2.  **Seguridad Paranoica:** Nunca confíes en el cliente. Valida todo input (Zod) en Server Actions.
3.  **Asincronía Total:** En Next.js 16, el acceso a datos de la request (`params`, `searchParams`, `cookies`, `headers`) es siempre asíncrono.
4.  **Colocación (Colocation):** El código que cambia junto, permanece junto. Estructura por Features, no por tipos de archivo.

---

## 1. Cambios Críticos (Next.js 15/16)

### Async Request APIs (Breaking Change)
El acceso síncrono a `params`, `searchParams`, `cookies()` y `headers()` está deprecado y fallará en v16.

```typescript
// ❌ Incorrecto (Legacy)
export default function Page({ params }) {
  const id = params.id; // Error en v16
}

// ✅ Correcto (Pro Standard)
export default async function Page({ 
  params 
}: { 
  params: Promise<{ id: string }> 
}) {
  const { id } = await params; // Await obligatorio
  
  // Mismo patrón para cookies y headers
  const cookieStore = await cookies();
  const token = cookieStore.get('token');
}
```

### Middleware vs Proxy Pattern
El `middleware.ts` corre en el **Edge Runtime** (limitado, sin Node APIs completas).
**Patrón Pro:** No pongas lógica de negocio pesada en el Middleware.
- **Middleware:** Solo para redirecciones rápidas, reescritura de rutas y validación básica de headers/cookies.
- **Proxy/BFF (Backend for Frontend):** Si necesitas lógica compleja, llamadas a BD directas o librerías Node.js, usa Route Handlers (`app/api/...`) o un proxy dedicado, no el middleware.

---

## 2. Caching & Data Fetching (The New Way)

Next.js 16 introduce la directiva `'use cache'` (experimental/stable según versión) para un control granular, reemplazando la complejidad de `fetch` options.

### Directiva `'use cache'`
Marca funciones o componentes específicos para ser cacheados, independientemente de dónde se llamen.

```typescript
// ✅ Cacheo a nivel de función (reusable)
'use cache'
 
export async function getProduct(id: string) {
  const product = await db.product.findUnique({ where: { id } });
  return product;
}
 
// ✅ Cacheo a nivel de página/componente
export default async function Page() {
  'use cache'
  // Todo este renderizado se cachea
  return <div>...</div>
}
```

### Legacy vs Modern
- **Antes:** `fetch(url, { next: { revalidate: 60 } })`
- **Ahora (Pro):** Usar `'use cache'` para lógica computacional y `unstable_cache` (o su sucesor estable) para control fino de tags.
- **Dynamic Data:** Para datos en tiempo real, simplemente no uses cache o fuerza dinamismo con `await connection()` o funciones dinámicas (`cookies()`, `headers()`).

---

## 3. Server Actions & Seguridad

Las Server Actions son endpoints públicos camuflados. **Deben** ser protegidas.

### Estructura de una Server Action Segura

1.  **Autenticación:** Verificar usuario.
2.  **Validación:** Usar Zod para inputs.
3.  **Mutación:** Operación en BD.
4.  **Revalidación:** Actualizar caché.
5.  **Manejo de Errores:** Retornar estados predecibles.

```typescript
// actions/update-profile.ts
'use server'

import { z } from 'zod';
import { auth } from '@/auth'; // Tu solución de auth
import { revalidatePath } from 'next/cache';
import { redirect } from 'next/navigation';

const ProfileSchema = z.object({
  username: z.string().min(3).max(20),
  email: z.string().email(),
});

export type ActionState = {
  success?: boolean;
  errors?: { [key: string]: string[] };
  message?: string;
};

export async function updateProfile(
  prevState: ActionState | null,
  formData: FormData
): Promise<ActionState> {
  // 1. Auth Check
  const session = await auth();
  if (!session?.user) {
    return { message: 'No autorizado' };
  }

  // 2. Validación (Zod)
  const validated = ProfileSchema.safeParse({
    username: formData.get('username'),
    email: formData.get('email'),
  });

  if (!validated.success) {
    return {
      errors: validated.error.flatten().fieldErrors,
      message: 'Error de validación',
    };
  }

  try {
    // 3. Mutación
    await db.user.update({
      where: { id: session.user.id },
      data: validated.data,
    });
  } catch (e) {
    return { message: 'Error de base de datos' };
  }

  // 4. Revalidación
  revalidatePath('/profile');
  
  // 5. Redirección (opcional, debe ser fuera del try/catch si usa 'throw')
  redirect('/dashboard');
}
```

---

## 4. Arquitectura Escalable (Feature-Based)

Evita agrupar archivos por "tipo" (todos los componentes en `/components`). Agrúpalos por "funcionalidad" (Feature).

```text
app/
├── (public)/              # Route Group: Landing, Marketing
│   ├── page.tsx
│   └── layout.tsx
├── (app)/                 # Route Group: Aplicación Principal
│   ├── dashboard/
│   │   ├── _components/   # Componentes ÚNICOS del dashboard (Colocation)
│   │   │   ├── metrics-card.tsx
│   │   │   └── chart-view.tsx
│   │   ├── actions.ts     # Server Actions específicas de esta feature
│   │   ├── layout.tsx
│   │   └── page.tsx
│   └── settings/
│       └── page.tsx
├── api/                   # Webhooks / Public API (No para uso interno)
└── _lib/                  # Utilidades compartidas, DB, DTOs
```

### Reglas de Directorio
- **`_components`**: Prefijo `_` para carpetas privadas (no rutas) dentro de `app/`.
- **`actions.ts`**: Coloca las acciones cerca de donde se usan.
- **Route Groups `(...)`**: Úsalos para separar layouts distintos (e.g., Marketing vs App Layout).

---

## Checklist de Implementación Pro

1.  [ ] **Linting:** ¿Estás usando `eslint-config-next` + reglas de accesibilidad?
2.  **Tipado:** ¿Usas `Awaited<ReturnType<typeof func>>` para inferir tipos de datos desde el servidor?
3.  **Performance:** ¿Has envuelto componentes pesados en `<Suspense>`?
4.  **Imágenes:** ¿Usas `next/image` con `sizes` correcto para evitar Layout Shift?
5.  **Metadata:** ¿Has definido `generateMetadata` dinámico para SEO?

---

## Referencias Rápidas
- **Dynamic Params:** `const { slug } = await params;`
- **Cache:** `'use cache'` (Next.js 16)
- **Revalidación:** `revalidateTag('products')` vs `revalidatePath('/')`
