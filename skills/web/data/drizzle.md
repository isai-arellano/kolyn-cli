---
check:
  required_deps: 
    - drizzle-orm
    - drizzle-kit
  files_exist:
    - drizzle.config.ts
---

# Drizzle ORM Pro (PostgreSQL Edition)

**Referencia:** Drizzle ORM es nuestra capa de acceso a datos preferida por ser "Type-Safe" y ligera. Siempre se usa en conjunto con **PostgreSQL**.

## Filosofía
1.  **Si compila, funciona:** Confía en la inferencia de TypeScript.
2.  **SQL-like:** La API se parece a SQL, no a un ORM mágico.
3.  **Migrations First:** Nunca edites la BD manualmente; usa `drizzle-kit`.

---

## 1. Configuración (`drizzle.config.ts`)

Asegúrate de configurar el dialecto para PostgreSQL.

```typescript
import { defineConfig } from 'drizzle-kit';

export default defineConfig({
  schema: './src/db/schema.ts',
  out: './drizzle',
  dialect: 'postgresql', // ⚠️ Importante: NO 'sqlite' ni 'mysql'
  dbCredentials: {
    url: process.env.DATABASE_URL!,
  },
});
```

---

## 2. Definición de Esquemas (`src/db/schema.ts`)

Usa los tipos de `drizzle-orm/pg-core`.

### Tipos de Datos Recomendados
*   **IDs:** Usa `uuid` o `serial` (preferiblemente `identity` si es PG 10+).
*   **Textos:** `text` es mejor que `varchar` en Postgres (mismo performance, sin límite arbitrario).
*   **Fechas:** `timestamp` con `mode: 'date'` para objetos JS Date nativos.

```typescript
import { pgTable, text, timestamp, uuid } from 'drizzle-orm/pg-core';

export const users = pgTable('users', {
  id: uuid('id').defaultRandom().primaryKey(),
  name: text('name').notNull(),
  email: text('email').notNull().unique(),
  createdAt: timestamp('created_at').defaultNow().notNull(),
});
```

### Relaciones (Drizzle Relations)
Define las relaciones para usar la API de consultas `query`.

```typescript
import { relations } from 'drizzle-orm';

export const usersRelations = relations(users, ({ many }) => ({
  posts: many(posts),
}));

export const posts = pgTable('posts', {
  id: uuid('id').defaultRandom().primaryKey(),
  authorId: uuid('author_id').references(() => users.id),
});

export const postsRelations = relations(posts, ({ one }) => ({
  author: one(users, {
    fields: [posts.authorId],
    references: [users.id],
  }),
}));
```

---

## 3. Consultas (Query API vs SQL-like)

### Query API (Recomendado para lecturas simples)
Similar a Prisma, ideal para traer relaciones anidadas.

```typescript
const result = await db.query.users.findMany({
  where: (users, { eq }) => eq(users.active, true),
  with: {
    posts: true, // Trae los posts automáticamente
  },
});
```

### SQL-like (Recomendado para performance/complejidad)
Control total sobre el SQL generado.

```typescript
import { eq } from 'drizzle-orm';

const result = await db.select()
  .from(users)
  .where(eq(users.email, 'test@example.com'));
```

---

## 4. Migraciones (Workflow)

El ciclo de vida de cambios en la BD:

1.  **Modificar:** Edita `src/db/schema.ts`.
2.  **Generar:** Crea el archivo SQL de migración.
    ```bash
    npx drizzle-kit generate
    ```
3.  **Aplicar:** Ejecuta la migración en la BD.
    ```bash
    npx drizzle-kit migrate
    ```

**⚠️ Pro Tip:** Nunca edites archivos generados en `drizzle/` manualmente.

---

## 5. Inferencia de Tipos

No escribas interfaces manuales. Infiérelas del esquema.

```typescript
import { type InferSelectModel, type InferInsertModel } from 'drizzle-orm';

// Tipo para SELECT (lo que devuelve la BD)
export type User = InferSelectModel<typeof users>;

// Tipo para INSERT (lo que necesitas para crear)
export type NewUser = InferInsertModel<typeof users>;
```

---

## Referencias
- [Drizzle ORM Docs](https://orm.drizzle.team)
- [PostgreSQL Adapter](https://orm.drizzle.team/docs/get-started-postgresql)
