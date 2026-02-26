# Pacto de Planes

Creado por: `000geid`  
Fecha: `2026-02-26`  
Idioma oficial para documentos Pacto: **español**

## Idioma y nomenclatura

- La documentación de planes se mantiene en **español**.
- El código, identificadores técnicos y convenciones de naming se mantienen en **inglés**.
- Esto aplica a: `id` de artefactos, slugs técnicos, nombres de archivos técnicos auxiliares y ejemplos de naming.
- Las carpetas y secciones documentales de Pacto se nombran en español (ejemplo: `artefactos/`).

## Propósito

Definir una estructura formal y consistente para crear, revisar estado y ejecutar planes en `pacto`.

## Comandos Pacto

Acciones canónicas:

- `pacto:status`
- `pacto:create`
- `pacto:execute`

Slash commands recomendados (agnósticos de IDE):

- `/pa-status` o `/pa:status`
- `/pa-new` o `/pa:new`
- `/pa-exec` o `/pa:exec`

Formato de argumentos:

- `pacto:<acción> [opciones]`

## Alcance del sistema

- Directorio fuente: `pacto`
- Estados de plan: `current`, `to-implement`, `done`, `outdated`
- Índice global: `pacto/README.md`

## Estructura canónica obligatoria

Todo plan nuevo debe incluir, como mínimo:

1. Metadatos

- `Título`
- `Versión`
- `Fecha`
- `Estado`
- `Owner`

1. Marco del problema

- `Resumen`
- `Contexto`
- `Objetivos`
- `No objetivos`

1. Ejecución

- `Fases` con estado por fase
- `Duración estimada` y, cuando aplique, `duración real`
- `Progreso global` y `restante estimado`

1. Validación

- `Plan de pruebas`
- `Checklist de smoke`
- `Evidencia` (fecha + comandos/resultados clave)

1. Gobernanza

- `Decisiones` (tipo ADR breve)
- `Dependencias`
- `Riesgos y mitigaciones`

1. Cierre

- `Criterios de éxito`
- `Entregables`
- `Siguientes pasos`

## Modelo mínimo de planificación

Para mantener Pacto simple y útil, el modelo oficial es:

- `Plan` -> `Fases` -> `Tareas`
- `Plan` -> `Artefactos` (opcionales, pero tipados)
- `Plan` -> `Deltas` (unidad principal de avance y verificación)

### Jerarquía operativa

1. **Plan**

- Tiene objetivo, estado global y progreso global.

1. **Fase**

- Tiene estado propio y criterio de salida.
- Agrupa tareas ejecutables.

1. **Tarea**

- Unidad mínima de ejecución.
- Debe poder marcarse como hecha/no hecha y, cuando aplique, referenciar evidencia.

1. **Artefacto**

- Elemento documental reusable del plan.
- Se declara con una fila en inventario y puede vivir en el mismo `README` o en archivo aparte.

### Tipos de artefacto permitidos (v1 mínima)

- `diagrama`
- `checklist`
- `documentacion`
- `evidencia`

Campos mínimos por artefacto:

- `id` (slug corto, único dentro del plan)
- `tipo` (uno de los 4 tipos)
- `estado` (`draft` | `active` | `done` | `outdated`)
- `ruta` (archivo o sección del plan)
- `owner` (persona/equipo)

Convención de naming:

- `id` y slugs en inglés, lowercase, separados por `-`.

## Reglas de actualización de estado

1. Fuente de verdad

- Si hay conflicto entre tabla inicial y trazas posteriores, prevalece el **delta** más reciente con fecha explícita.

1. Fechas

- Todo cambio de estado debe incluir fecha absoluta (`YYYY-MM-DD`).

1. Evidencia mínima para marcar fase completada

- Validación técnica (tests/lint o equivalente).
- Evidencia funcional (smoke/manual/API) cuando aplique.

1. Consistencia del índice

- Al crear/mover/eliminar planes, actualizar `pacto/README.md` (conteos + enlaces).

1. Regla de simplicidad

- Cada plan inicia con máximo 1 diagrama y 1 checklist.
- Solo agregar más artefactos cuando exista una necesidad operativa explícita.

## Modelo de Deltas (obligatorio)

`Delta` es la unidad atómica de cambio en Pacto (similar a un commit/PR pequeño).

Campos mínimos por delta:

- `id` (ej: `D-2026-02-26-01`)
- `fecha` (`YYYY-MM-DD HH:MM`)
- `scope` (fase/módulo)
- `tipo` (`feat` | `fix` | `refactor` | `docs` | `test`)
- `estado` (`applied` | `partial` | `reverted`)
- `cambios` (`+` agregado, `~` modificado, `-` removido)
- `validación` (comando + resultado)
- `siguiente_delta` (acción concreta)

Reglas:

1. Cada slice ejecutado en `pa-exec` debe registrar al menos 1 delta.
2. `checkpoint` pasa a ser opcional, solo como snapshot de resumen.
3. El estado/progreso del plan debe derivar del historial de deltas, no solo de texto narrativo.

## Capa de verificación de estado (`pa-status`)

`pa-status` debe verificar que la información del plan coincide con el código/evidencia disponible.

Reglas:

1. No asumir veracidad por texto
- Una afirmación en el plan no cuenta como validada si no tiene evidencia contrastable.

2. Validaciones mínimas por plan
- Verificar existencia de rutas de archivos mencionadas.
- Verificar presencia de símbolos/endpoints declarados mediante búsqueda dirigida.
- Verificar coherencia entre estado/progreso declarado y último delta con fecha.
- Si hay soporte en la herramienta, permitir uso de subagentes para paralelizar verificación por plan o dominio técnico.

3. Clasificación obligatoria de verificación
- `verified`: evidencia suficiente y consistente.
- `partial`: evidencia incompleta o parcialmente consistente.
- `unverified`: evidencia ausente, obsoleta o contradictoria.

4. Reporte de salida
- `pa-status` debe incluir, por plan: estado, bloqueadores, siguiente acción y clasificación de verificación.
- Siempre incluir referencias concretas (ruta/claim) para sustentar la clasificación.

## Estructura recomendada por plan (opcional)

Dentro de `pacto/<estado>/<slug>/`:

- `README.md` (fuente principal del plan)
- `artefactos/diagramas/` (solo si hay archivos `.mmd` o exportes)
- `artefactos/checklists/` (solo si un checklist crece demasiado para el README)

Si el plan es pequeño, todo puede quedarse en `README.md` sin carpetas extra.

## Plantilla oficial

Para nuevos planes Pacto usar:

- `pacto/PLANTILLA_PACTO_PLAN.md`

## Evolución

Este Pacto define la versión inicial formal. Puede evolucionar con nuevas secciones (por ejemplo, contratos machine-readable) manteniendo compatibilidad con esta base.
