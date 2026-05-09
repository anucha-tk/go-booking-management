---
description: Execute a full BMad development cycle for a specific epic. Triggered by /bmad-full-cycle {epic}.
---

# /bmad-full-cycle - Full BMad Development Cycle

Use this workflow to implement a full development cycle for an epic, covering story creation, development, review, and validation.

## Prerequisites
- User must specify epic as argument: `/bmad-full-cycle {{epic}}`
- Project context `_bmad-output/project-context.md` must be available.

## Instructions

When `/bmad-full-cycle {{epic}}` is called, follow these steps strictly and autonomously:

### 0. Preparation
- **Mode**: Activate `/caveman full` mode. Respond terse. No fluff.
- **Context**: Read `_bmad-output/project-context.md`. Follow all rules (TDD, DI, sqlc, etc.).
- **Environment**: Always prepend `export PATH="/opt/homebrew/bin:/Users/anucha-tk/go/bin:$PATH"` to commands.

### 1. Create Story
- Use `bmad-create-story` skill for `{{epic}}`.
- Define clear requirements and acceptance criteria.
- **State Gate**: Ensure story is in `ready-for-dev` in `sprint-status.yaml`.

### 2. Dev Story (Implementation)
- **Status**: Move story to `in-progress` in `sprint-status.yaml`.
- Use `bmad-dev-story` skill to implement the story.
- **TDD Mandatory**: Write tests first. Achieve >= 80% coverage.
- Follow DDD layers and project patterns.

### 3. Code Review (CRITICAL GATE)
- **Status**: Move story to `review` in `sprint-status.yaml`.
- **Action**: Use `bmad-code-review` skill.
- **Artifact**: Create a Review Report in `_bmad-output/reviews/{{story_id}}-review.md`.
- **Adversarial Analysis**: Identify security, architectural, and quality issues. 
- **NO SKIPPING**: You cannot move to 'done' without this report and a 'review' status transition.

### 4. Improve Code
- Fix all issues identified in the Review Report.
- Refactor for maximum clarity and maintainability.
- Update the Review Report to mark issues as resolved.

### 5. Boost Test Coverage & Final Validation
- Check coverage. Must be **>= 80%**.
- Run `task lint`. Fix any issues.
- Run `task test`. Ensure 100% pass rate.
- Final check against `_bmad-output/project-context.md`.

### 6. Completion
- **Status**: Move story to `done` in `sprint-status.yaml` ONLY after all above steps are verified.
- **Cleanup**: Remove any temporary debug files.

## Definition of Done (DoD)
- [ ] Story defined with AC and in `ready-for-dev`.
- [ ] Code implemented with tests and in `in-progress`.
- [ ] Formal Review Report exists in `_bmad-output/reviews/` and story was in `review`.
- [ ] All review issues addressed and verified.
- [ ] Coverage >= 80%, Lint passes, Tests 100% pass.
- [ ] Status set to `done`.

## Workflow Rules
- **Autonomy**: Do not ask for confirmation between steps.
- **Strict Sequencing**: Follow Step 1 -> 6. No skipping to 6 from 2.
- **Final Code Focus**: The user wants to see the final, polished code.
- **Caveman Style**: Keep prose minimal. Focus on technical substance.
