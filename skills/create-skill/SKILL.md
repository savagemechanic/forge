---
name: create-skill
description: Creates new reusable skills from completed work sessions. Distills a workflow into a SKILL.md with triggers, steps, and safety rules. Use when the user says "create a skill from what we just did."
version: 1.0.0
---

# Create-Skill — Distill Workflows Into Reusable Skills

You turn a completed work session into a reusable skill.

## Process

1. **Identify the reusable pattern.** What did we just do that could apply to other projects?
2. **Strip project-specific details.** Remove names, paths, and data unique to this session.
3. **Write the SKILL.md** with:
   - **Name:** verb-noun, lowercase-hyphenated (e.g. `extract-interface`)
   - **Description:** one sentence starting with a verb
   - **Triggers:** what prompts should activate this skill
   - **Steps:** the workflow, generalized
   - **Safety rules:** what must never happen
4. **Install it** to the appropriate scope:
   - `folder` — project-specific (`.forge/skills/`)
   - `global` — cross-project (`~/.forge/skills/`)
   - `builtin` — ships with Forge (requires PR)

## SKILL.md Template

```markdown
---
name: <verb-noun>
description: <one sentence, verb-first>
version: 1.0.0
---

# <Title>

<When to use this skill>

## Steps
1. ...
2. ...

## Safety Rules
- Never ...
```

## Scope Decision

- Does this apply to ONE project? → **folder**
- Does this apply to YOUR workflow across projects? → **global**
- Does this apply to ALL Forge users? → **builtin** (open a PR)
