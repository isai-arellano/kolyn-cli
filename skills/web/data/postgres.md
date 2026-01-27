# PostgreSQL Pro (Diseño de Bases de Datos)

Diseño robusto, normalizado y eficiente para aplicaciones modernas.

## Reglas de Oro

1.  **Normalización:** Diseña en **3NF** (Tercera Forma Normal) por defecto. Evita duplicar datos. Desnormaliza solo si tienes pruebas de problemas de rendimiento.
2.  **Identificadores:**
    *   **Numéricos:** Usa `BIGINT GENERATED ALWAYS AS IDENTITY`. Es el estándar moderno (reemplaza a `SERIAL`).
    *   **Globales:** Usa `UUID` (v7 preferiblemente si PG17+, si no v4 con `gen_random_uuid()`) cuando necesites unicidad global o ids opacos.
3.  **Nulos:** Usa `NOT NULL` siempre que sea semánticamente posible. Los nulos complican la lógica y las consultas.
4.  **Fechas:** Usa siempre `TIMESTAMPTZ` (Timestamp with Time Zone). `TIMESTAMP` sin zona horaria es una bomba de tiempo.

---

## Tipos de Datos Recomendados

| Tipo de Dato | Recomendación PostgreSQL | Por qué |
| :--- | :--- | :--- |
| **Texto** | `TEXT` | `VARCHAR(n)` no ofrece ventaja real de performance en PG. |
| **Enteros** | `BIGINT` | Evita desbordamientos (overflows) futuros. `INT` es arriesgado hoy en día. |
| **Decimales** | `NUMERIC` | Para dinero y cálculos exactos. Nunca uses `FLOAT` o `REAL` para dinero. |
| **JSON** | `JSONB` | Binario, indexable y rápido. Evita `JSON` plano. |
| **Fechas** | `TIMESTAMPTZ` | Maneja conversiones de zona automáticamente. |

---

## Índices y Performance

PostgreSQL no crea índices en Foreign Keys automáticamente. **Debes hacerlo tú.**

### 1. Índices Obligatorios
*   **Foreign Keys:** Siempre indexa las columnas FK (`user_id`, `order_id`) para evitar bloqueos y acelerar JOINs.
*   **Filtros Frecuentes:** Columnas usadas en `WHERE`, `ORDER BY` o `GROUP BY`.

### 2. Índices JSONB (GIN)
Si guardas datos en JSONB y consultas por claves internas, necesitas un índice GIN.

```sql
CREATE INDEX idx_users_metadata ON users USING GIN (metadata);
-- Permite: WHERE metadata @> '{"role": "admin"}'
```

### 3. Índices Parciales
Ahorra espacio indexando solo lo que importa.

```sql
-- Solo indexa usuarios activos (ahorra espacio y acelera inserts)
CREATE INDEX idx_active_users ON users (email) WHERE active = true;
```

---

## Gotchas Comunes

*   **Identificadores (Case Sensitivity):** Postgres convierte todo a minúsculas a menos que uses comillas dobles `"Tabla"`. **Regla:** Usa siempre `snake_case` para tablas y columnas y olvídate de las comillas.
*   **Sequences:** Los IDs autogenerados pueden tener huecos (gaps) debido a rollbacks o concurrencia. **Es normal**, no intentes arreglarlo.
*   **Unique + Null:** `UNIQUE` permite múltiples `NULL`s por defecto. Si quieres solo un nulo, usa `UNIQUE NULLS NOT DISTINCT` (PG15+).

---

## Referencias
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Modern SQL](https://modern-sql.com/)
