---
name: improve-skill
description: Reviews and improves existing skills for clarity, completeness, and safety. Identifies missing steps, vague triggers, and unsafe patterns. Use when refining a skill after real-world use.
version: 1.0.0
---

# Improve-Skill — Refine Skills After Use

You review existing skills and make them sharper.

## Review Checklist

1. **Triggers too vague?** Would the skill activate when it shouldn't, or miss when it should?
2. **Steps too abstract?** Can a reader follow them without guessing?
3. **Missing safety rules?** What could go wrong that isn't guarded?
4. **Over-specified?** Does it hardcode a path or pattern that won't generalize?
5. **Version stale?** Bump the patch version after each improvement.

## Improvement Process

1. **Re-read the SKILL.md** and the skill's references.
2. **Recall how it was used** — did it help? Did it get in the way?
3. **Tighten triggers** — be specific about WHEN to activate.
4. **Add missing steps** — fill gaps discovered during use.
5. **Strengthen safety rules** — add guards for failure modes seen.
6. **Bump version** — `1.0.0` → `1.0.1` for fixes, `1.1.0` for new steps.

## Anti-Patterns to Remove

- Steps that say "do the right thing" without saying what that is.
- Triggers that match too broadly (e.g. "when modifying code").
- Rules that conflict with the project's actual conventions.
