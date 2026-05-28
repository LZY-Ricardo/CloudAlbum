# Core Security & Access Control Review Config

- **Execution mode:** inline (`superpowers:executing-plans` default)
- **Task-level review:** self-checklist only; no external review during individual tasks
- **Feature-level review:** one external code review after the full plan is complete
- **Review executor:** subagent at feature closeout only
- **Hard rules:**
  - self-checklist always runs
  - TDD failing-test step cannot be skipped
  - any plan deviation triggers one escalation review
  - no task-level external review unless a plan deviation forces escalation
