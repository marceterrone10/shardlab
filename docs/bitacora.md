# Bitácora

## 2026-10-09 — Skeleton inicial

- Se creó un módulo único de Go para compartir contratos entre coordinador,
  analizador y generador.
- El ruteo inicial usa `(cliente_id - 1) % cantidad_de_shards`.
- El coordinador sólo expone decisiones de ruteo; todavía no abre conexiones a
  PostgreSQL.
- El generador produce CSV deterministas, pero la carga y separación física por
  shard queda como siguiente paso.

Las mediciones futuras deben registrar versión de PostgreSQL, hardware,
volumen, concurrencia, configuración y semilla.
