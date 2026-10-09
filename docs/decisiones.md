# Decisiones de diseño

## ADR-001 — Un módulo de Go

**Estado:** aceptada.

El coordinador, el analizador y el generador viven en un solo módulo. Esto
permite compartir contratos y ejecutar `go test ./...` desde la raíz. Si sus
ciclos de despliegue divergen, se puede separar el módulo sin cambiar los
paquetes públicos.

## ADR-002 — Ruteo inicial por módulo

**Estado:** aceptada para el skeleton.

La función `(cliente_id - 1) % N` ofrece una regla fácil de reproducir en el
generador y el coordinador. No se asume que sea compatible con el hash interno
de PostgreSQL. Antes de cargar datos, el loader deberá usar exactamente esta
misma regla.

**Alternativas consideradas:** hash consistente, rangos y hash nativo de
PostgreSQL. Se posponen hasta poder medir el costo de rebalanceo y la
distribución real.

## ADR-003 — Interfaces antes que ejecución distribuida

**Estado:** aceptada.

El primer corte publica health, ruteo y scatter-plan, sin simular resultados
SQL. La conexión, cancelación, límites de concurrencia y merge se incorporarán
con pruebas de integración en fases separadas.
