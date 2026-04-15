# Evaluate OpenSpec config context model for Pacto

**Created At:** 2026-04-13 23:29  
**Updated At:** 2026-04-13 23:29

## Summary

Idea exploration workspace.

## Notes

- [2026-04-13 23:29] Idea created.
- [2026-04-13 23:29] Investigate whether Pacto should keep .pacto/context/domains/, adopt an OpenSpec-style global config context blob, or combine both. Key findings so far: Spec Kit uses shared memory/constitution rather than accumulated domain docs; OpenSpec moved from project.md to openspec/config.yaml context plus per-artifact rules; OpenSpec context is injected into all artifact instructions and is not itself a canonical knowledge folder. Follow-up should compare tradeoffs for global prompt context vs durable domain knowledge docs and identify a minimal Pacto direction.
- [2026-04-14 07:55] Created [exploration.md](file:///home/diego/sandbox/pacto/.pacto/ideas/openspec-config-context-for-pacto/exploration.md) to compare Pacto, OpenSpec, and Spec Kit approaches. Proposed a "Combined" direction that keeps domain docs while adding a rich `context` and `rules` block to `.pacto/config.yaml`.
