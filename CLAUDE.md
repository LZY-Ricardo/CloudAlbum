## Enhanced Superpowers Workflow

This project uses the Enhanced Superpowers workflow. The following rules are MANDATORY for all AI agents.

### Mandatory Rules

1. **Before writing ANY code**, invoke `brainstorming` skill (Superpowers). No exceptions.
2. **Before executing**, invoke `writing-plans` skill (Superpowers). No exceptions.
3. **After each task commit**, invoke `documenting-execution` skill.
4. **After each verification run**, invoke `documenting-verification` skill.
5. **After each code review cycle**, invoke `documenting-review` skill.
6. **After resolving any bug**, invoke `documenting-debugging` skill.
7. **Before merge or PR**, invoke `documenting-completion` skill.
8. **For large requirements** (3+ features or subsystems), invoke `decomposing-requirements` skill BEFORE brainstorming.

### Strict Prohibitions

- Do NOT write code before brainstorming is approved by the human partner.
- Do NOT claim work is complete without running verification commands.
- Do NOT silently drop review findings. Every finding = FIXED, DEFERRED (with reason), or REJECTED (with evidence).
- Do NOT merge or create PR before the completion summary is written.
- Do NOT leave documentation updates uncommitted at session end.

### Documentation

- All workflow docs go in `docs/superpowers/` — see `docs/superpowers/README.md` for full details.
- Documentation is MANDATORY when using this workflow. Ad-hoc changes under 30 minutes are exempt.
