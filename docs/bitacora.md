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

## 2026-10-12 — Carga repartida

- `make load` genera cada cliente y lo inserta en el baseline y en un shard.
- El destino sale de `ShardForCustomer`. Las filas hijas acompañan a su
  `cliente_id`.
- Al terminar, el comando comprueba que el cliente 4 está en el baseline y en
  `shard0`, y que la suma de los shards coincide con el baseline.
- Los CSV quedan como salida de inspección. La carga no los lee.
