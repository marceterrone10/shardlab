# Benchmarks

Este directorio contendrá:

- cargas de `pgbench` o `k6`;
- consultas y planes `EXPLAIN (ANALYZE, BUFFERS)`;
- metadatos del entorno y configuración;
- resultados crudos bajo `bench/results/` (ignorados por Git);
- gráficos y conclusiones reproducibles.

Cada corrida debe indicar volumen, concurrencia, semilla, versión de
PostgreSQL y commit de shardlab. No se agregan resultados manuales sin el
comando que los reproduzca.
