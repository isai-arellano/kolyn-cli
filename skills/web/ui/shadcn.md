# Shadcn UI Pro (Arquitectura de Componentes)

Eres un experto en **Design Systems** implementando **Shadcn UI**. Tu objetivo no es solo copiar y pegar, sino construir una arquitectura de UI mantenible, accesible y consistente.

## Filosofía de Diseño

1.  **Ownership:** Shadcn no es una librería npm, es *tu* código. Puedes y debes modificar los componentes en `components/ui` para adaptarlos a tu marca.
2.  **Composition:** Prefiere componentes pequeños y componibles sobre props complejas ("Prop Drilling").
3.  **Utility-First:** Usa Tailwind CSS para todo. Evita CSS modules o styles-in-js a menos que sea animación compleja.

---

## 1. La Función `cn()` (ClassNames)

El corazón de Shadcn. Úsala siempre para permitir que los componentes reciban clases externas y las mezclen inteligentemente (resolviendo conflictos de Tailwind).

```typescript
// lib/utils.ts
import { type ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

// Uso en componente
export function Card({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div 
      className={cn("rounded-xl border bg-card text-card-foreground shadow", className)} 
      {...props} 
    />
  )
}
```

---

## 2. Estructura de Componentes

### Componentes Base vs. Feature Components

*   **`components/ui/`**: Solo primitivas de Shadcn (Button, Input, Card). Son "tontos" (sin lógica de negocio) y altamente reutilizables.
*   **`app/(feature)/_components/`**: Componentes complejos que usan lógica de negocio o combinan múltiples primitivas (ej. `UserProfileCard`, `DataTable`).

```text
components/
└── ui/
    ├── button.tsx
    ├── input.tsx
    └── toast.tsx
app/
└── dashboard/
    └── _components/
        ├── metrics-card.tsx  <-- Usa Card + Button + Iconos
        └── user-nav.tsx      <-- Usa DropdownMenu + Avatar
```

---

## 3. Theming & Variables

Nunca hardcodees colores hex (`#000`). Usa variables CSS semánticas para soportar Dark Mode nativamente.

```css
/* globals.css */
@layer base {
  :root {
    --background: 0 0% 100%;
    --foreground: 222.2 84% 4.9%;
    --primary: 222.2 47.4% 11.2%; /* Azul Marino */
    --primary-foreground: 210 40% 98%;
    --destructive: 0 84.2% 60.2%;
  }
 
  .dark {
    --background: 222.2 84% 4.9%;
    --foreground: 210 40% 98%;
    --primary: 210 40% 98%; /* Blanco */
    --primary-foreground: 222.2 47.4% 11.2%;
  }
}
```

Uso en Tailwind: `bg-primary text-primary-foreground`.

---

## 4. Patrones Comunes

### Botones con Loading State
Extiende el componente Button para manejar estados de carga visualmente.

```typescript
import { Loader2 } from "lucide-react"

<Button disabled={isLoading}>
  {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
  Guardar Cambios
</Button>
```

### Formularios (React Hook Form + Zod)
El estándar para formularios en Shadcn.

```typescript
<Form {...form}>
  <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
    <FormField
      control={form.control}
      name="email"
      render={({ field }) => (
        <FormItem>
          <FormLabel>Email</FormLabel>
          <FormControl>
            <Input placeholder="tu@empresa.com" {...field} />
          </FormControl>
          <FormMessage />
        </FormItem>
      )}
    />
  </form>
</Form>
```

---

## 5. Checklist de Calidad UI

1.  [ ] **Accesibilidad:** ¿Todos los `img` tienen `alt`? ¿Los botones sin texto tienen `aria-label`?
2.  **Responsive:** ¿El diseño funciona en móvil (`sm:`) y escritorio (`lg:`)?
3.  **Dark Mode:** ¿Verificaste que los bordes y textos sean legibles en modo oscuro?
4.  **Feedback:** ¿Usas `Toast` o `Alert` para confirmar acciones del usuario?

---

## Referencias
- [Shadcn UI Docs](https://ui.shadcn.com)
- [Lucide Icons](https://lucide.dev)
