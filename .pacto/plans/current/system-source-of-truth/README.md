# System Source of Truth

**Status:** In Progress (Current)  
**Date:** 2026-04-08

## Summary

Give pacto a lightweight system-level source of truth so plans can detect conflicts, accumulate knowledge, and store it in a per-domain context tree instead of a single merged markdown document.

## Problem

Pacto plans are currently siloed. Two active plans can contradict each other without any shared reference point, and decisions from completed work disappear into historical artifacts instead of informing future plans.

This plan adds a small, markdown-first context workspace inside `.pacto/`, with one document per domain so knowledge can grow without turning into a noisy monolith.

## Documents

- [spec.md](./spec.md)
- [design.md](./design.md)
- [tasks.md](./tasks.md)
