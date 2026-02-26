# Slash Commands de Pacto

Este archivo define los comandos de Pacto de forma agnóstica de IDE.

## Formas soportadas

Puedes usar cualquiera de estas variantes según soporte de tu editor:

- `/pa-status` o `/pa:status`
- `/pa-new` o `/pa:new`
- `/pa-exec` o `/pa:exec`

## Contrato de cada comando

### 1) `pa-status`

Objetivo: ver estado consolidado de planes activos.

Lee:
- `pacto/README.md`
- `pacto/current/*/README.md`

Entrega:
- planes en ejecución
- progreso por plan
- bloqueadores
- siguiente acción concreta
- resultado de verificación por plan (`verified` | `partial` | `unverified`)

Regla: si hay conflicto de estado en un plan, prevalece el **delta** más reciente con fecha explícita.

#### Capa de verificación (obligatoria)

`pa-status` no confía ciegamente en el documento del plan. Debe contrastar claims contra código:

1. Extraer claims verificables del plan:
- rutas de archivos
- endpoints/rutas API
- afirmaciones de tests/checks
- porcentaje/estado declarado
- deltas recientes (scope, cambios, validación)

2. Verificar evidencia en repo:
- existencia de archivos
- presencia de símbolos/rutas por búsqueda dirigida
- coherencia entre estado declarado y último delta fechado
- cuando la plataforma lo permita, usar subagentes para dividir verificaciones por plan/área en paralelo

3. Clasificar por plan:
- `verified`: evidencia suficiente y coherente.
- `partial`: evidencia incompleta o parcialmente coherente.
- `unverified`: no hay evidencia suficiente o hay contradicción.

### 2) `pa-new`

Objetivo: crear un plan nuevo en `pacto/`.

Uso:
- `/pa-new <current|to-implement> <slug>`

Acciones mínimas:
1. Crear carpeta `pacto/<estado>/<slug>/`.
2. Crear `README.md`.
3. Crear `PLAN_<TOPIC>_<YYYY-MM-DD>.md`.
4. Usar `pacto/PLANTILLA_PACTO_PLAN.md`.
5. Actualizar `pacto/README.md` (contador, enlace y fecha).

### 3) `pa-exec`

Objetivo: ejecutar un plan existente por slices verificables.

Uso:
- `/pa-exec <ruta_plan_md>`

Acciones mínimas:
1. Extraer fases/tareas ejecutables.
2. Ejecutar en pasos pequeños.
3. Validar con tests/checks aplicables.
4. Registrar evidencia (fecha absoluta + comando + resultado).
5. Registrar un **delta** por cada slice ejecutado.
6. Actualizar progreso/estado en el plan.
7. (Opcional) crear checkpoint de snapshot al cerrar un bloque mayor.

## Convenciones

- Documentación en español.
- Naming técnico (`id`, slugs) en inglés.
