# Pacto

Creado por: `pacto-cli`  
Última actualización: `2026-02-27`

## Propósito

Pacto es un flujo liviano para planes de ingeniería asistida por IA:

- escribir planes antes de implementar
- mantener estado explícito del plan
- validar afirmaciones de implementación con evidencia del repositorio
- producir estado legible por máquinas para CI/automatización

## Modelo de Workspace

Raíz canónica de planes:

- `./.pacto/plans` (creada por `pacto init`)

También soportado:

- cualquier directorio que ya contenga las 4 carpetas de estado (uso avanzado/manual)

Carpetas de estado requeridas:

- `current`
- `to-implement`
- `done`
- `outdated`

Archivos base en la raíz:

- `README.md` (vista general del workspace)
- `PACTO.md` (este contrato)
