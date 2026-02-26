# Slash Commands de Pacto

Formas soportadas:

- `/pa-status` o `/pa:status`
- `/pa-new` o `/pa:new`
- `/pa-exec` o `/pa:exec`

## Contrato mínimo

### `pa-status`

- Lee planes y reporta estado consolidado.
- Incluye verificación por plan: `verified|partial|unverified`.

### `pa-new`

- Crea carpeta de plan en estado objetivo.
- Genera `README.md` y `PLAN_<TOPIC>_<YYYY-MM-DD>.md`.
- Actualiza índice raíz.

### `pa-exec`

- Ejecuta plan por slices verificables.
- Registra evidencia y deltas.
