# Exploration: OpenSpec Config Context for Pacto

## Overview
This document explores the idea of adopting an OpenSpec-style global configuration context for Pacto, potentially replacing or augmenting the current `.pacto/context/domains/` model.

## Current State of Pacto Context
Pacto currently manages system knowledge through:
1. **.pacto/config.yaml**: Basic project settings (technologies, tools, languages).
2. **.pacto/context/domains/*.md**: Markdown files for specific domains (e.g., `auth.md`, `billing.md`).
   - Pacto automatically extracts domains from plans and ensures these docs exist.
   - They contain `Summary`, `Related Plans`, `Decisions`, and `Constraints`.

## OpenSpec Inspiration
OpenSpec uses a centralized `openspec/config.yaml` which includes:
- **context**: A block of text injected into all artifact instructions (e.g., tech stack details, platform requirements).
- **rules**: Type-specific rules for `specs`, `tasks`, and `design` files.

## spec-kit Inspiration
Spec Kit emphasizes "shared memory/constitution". It uses scripts to push "rules" into agent-specific configurations (like `.windsurf/rules/specify-rules.md`).

## Comparison of Approaches

| Feature | Pacto (Current) | OpenSpec (Pure) | Combined (Proposed) |
|---------|-----------------|-----------------|----------------------|
| **Knowledge Storage** | Distributed (`domains/*.md`) | Centralized (`config.yaml`) | Both |
| **Global Rules** | Limited to `config.yaml` fields | Rich `context` + `rules` blocks | Rich `context` + `rules` blocks |
| **Domain Depth** | High (separate files) | Low (all in one file) | High (keep domain docs) |
| **Agent Injection** | Manual/Implicit | Automatic (prompt injection) | Automatic (config + domain) |

## Proposed Direction: "Context-First" Pacto
We should evolve Pacto's context model to support "Global Prompt Context" while retaining "Durable Domain Knowledge".

### 1. Enhance `.pacto/config.yaml`
Add `context` and `rules` blocks to the existing config file.
```yaml
context: |
  Tech stack: Go, CLI tools, Markdown-based plans.
  This tool must behave predictably across OSes.

rules:
  plans:
    - Domain references must use the normalized slug.
    - Every claim must have a corresponding verification step.
```

### 2. Keep Domain Docs as "Deep Knowledge"
Retain `.pacto/context/domains/*.md` but treat them as supplemental to the global context. When an agent works on a plan affecting a domain, both the global context/rules AND the domain-specific doc should be provided to the agent.

### 3. Automated Rule Injection
Implement a `pacto` command (or hook) that propagates these rules to agent-specific "system instructions" or "rules" files (e.g., `.cursorrules`, `.clauderules`), similar to Spec Kit's update scripts.

## Next Steps
- [ ] Prototoype rule injection for a specific agent (e.g., Cursor or Claude).
- [ ] Evaluate if `rules` should be segmented by plan state (`to-implement` vs `current`).
- [ ] Decide if `domains/*.md` should also be summarized into the global config or kept separate.
