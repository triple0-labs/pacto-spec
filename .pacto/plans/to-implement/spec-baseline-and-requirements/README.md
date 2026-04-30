# Spec Baseline and Requirements Grammar

**Status:** To Implement  
**Date:** 2026-04-29

## Summary

Introduce a persistent capability baseline at `.pacto/specs/<capability>/spec.md`
and a structured `### Requirement` / `#### Scenario` grammar inside plan
`spec.md` files. Plans become deltas against the baseline; on `pacto move done`
the deltas are folded back into the baseline so the workspace always answers
"what is the system today?" separately from "what is this plan changing?".

## Problem

Pacto plans are the only source of truth. There is no persistent baseline that
answers "what does the system already do?", so every plan re-derives that
context from code. Requirements inside `spec.md` are free-form prose, which
prevents reliable cross-artifact analysis (requirement → tasks → evidence)
and stops `pacto status` from reporting per-requirement coverage.

This plan closes both gaps by (a) standing up a capability baseline tree and
(b) defining a strict, parseable grammar for requirements and scenarios.

## Documents

- [spec.md](./spec.md)
- [design.md](./design.md)
- [tasks.md](./tasks.md)

## Inspirations

- OpenSpec: `openspec/specs/<capability>/spec.md` baseline + `changes/<id>/specs/<capability>/spec.md` deltas with `## ADDED|MODIFIED|REMOVED Requirements`.
- spec-kit: `### Requirement` and structured FR-### / acceptance scenarios as the addressable units of a spec.
