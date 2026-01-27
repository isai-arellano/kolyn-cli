---
check:
  required_deps: 
    - react-icons
    - framer-motion
    - sonner
    - tailwind-merge
    - class-variance-authority
  forbidden_deps: 
    - lucide-react
    - react-hot-toast
---

# UI Tech Stack (Librerías Estándar)

Este documento define las librerías obligatorias para la capa de UI. La consistencia es clave: no mezcles librerías de iconos ni de animaciones.

## 1. Iconografía: React Icons
Usamos **React Icons** como estándar por su inmensa variedad y tree-shaking.

```typescript
// ✅ Correcto
import { FaBeer } from 'react-icons/fa';

function Component() {
  return <h3><FaBeer className="mr-2 text-primary" /> Salud</h3>
}
```

## 2. Animaciones: Framer Motion
Toda interacción significativa debe tener una animación sutil. Usamos **Framer Motion** por su API declarativa.

*   **Micro-interacciones:** Hover, Tap.
*   **Transiciones de página:** Layout animations.
*   **Listas:** `AnimatePresence` para elementos que entran/salen.

```typescript
// ✅ Patrón de animación de entrada
import { motion } from "framer-motion"

export function FadeIn({ children }: { children: React.ReactNode }) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4 }}
    >
      {children}
    </motion.div>
  )
}
```

## 3. Notificaciones: Sonner
Reemplazamos el `Toast` por defecto de Shadcn con **Sonner** por su diseño superior y facilidad de uso.

**Instalación:** `npx shadcn@latest add sonner`

```typescript
import { toast } from "sonner"

// ✅ Uso en Event Handlers
<Button 
  onClick={() => 
    toast.success("Proyecto creado", {
      description: "Domingo, 23 de Diciembre - 9:00 AM",
      action: {
        label: "Deshacer",
        onClick: () => console.log("Undo"),
      },
    })
  }
>
  Crear
</Button>
```

## 4. Estilos: Tailwind CSS + Class Variance Authority (CVA)
Nunca escribas CSS puro. Usa `cva` para manejar variantes de componentes complejos.

```typescript
const buttonVariants = cva(
  "inline-flex items-center justify-center rounded-md text-sm font-medium",
  {
    variants: {
      variant: {
        default: "bg-primary text-primary-foreground hover:bg-primary/90",
        destructive: "bg-destructive text-destructive-foreground",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
)
```
