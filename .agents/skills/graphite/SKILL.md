---
name: graphite
description: Manage changes in stacked PRs using Graphite CLI (gt). Use for pushing changes, creating stacked PRs, submitting code for review, rebasing stacks, navigating branches, syncing with trunk, and dependent/stacked Git changes. Includes recommended workflow patterns.
---

# Stacked Git Workflow with Graphite CLI (gt)

Graphite simplifies stacked PR workflows on GitHub. Each branch only has one commit, and is in a stack builds on top of the previous one, keeping changes small, focused, and reviewable.

## Commands:
- `git` Git CLI
- `gt` Graphite CLI
- `gh` GitHub CLI — **only** for fixing PR description bodies on GitHub, or marking an already-open draft PR ready (see below)

## Key Concepts

- **Stack**: A sequence of PRs, each building off its parent. e.g. `main ← PR1 ← PR2 ← PR3`
- **Trunk**: The base branch stacks merge into (usually `main`)
- **Downstack**: PRs below the current one (ancestors)
- **Upstack**: PRs above the current one (descendants)

## Setup

```bash
git init  # Initialize git repository
```

## Discover Graphite CLI Features

```bash
gt --help
```

## Creating & Stacking All Changes

```bash
# Create a new branch with staged changes (stacks on current branch)
gt create -m "description of changes"

# Follow the step of syncing-and-rebasing
...

# Push to remote (publish = ready for review, not draft)
gt submit --stack --publish
```

## Syncing & Rebasing

```bash
# Sync trunk from remote, rebase all stacks, clean up merged branches
gt sync
```

## Resolve Merge Conflict

```bash
# Restack all branches in the current stack (fix parent history)
gt sync && gt restack

# Iteratively resolve conflict
...

# Continue the rebase and restacking
gt continue
```

## Submit Current PR

```bash
gt submit --publish
```

## Submitting Entire PRs Stack

```bash
gt submit --stack --publish
```

In non-interactive mode (`--no-interactive`, Cursor Agent), `gt submit` creates **draft** PRs unless you pass `--publish`. Always include `--publish` when submitting from an agent session.

After submit, **always share the Graphite PR link** from the terminal output (e.g. `https://app.graphite.com/github/pr/OWNER/REPO/NUMBER`) with the user — not just the GitHub URL.

## Navigating Stacks

```bash
gt up [steps]              # Move to child branch
gt down [steps]            # Move to parent branch
gt top                     # Jump to tip of current stack
gt bottom                  # Jump to branch closest to trunk
gt checkout [branch]       # Interactive branch selector (or specify branch)
gt log                     # Visual graph of current stack
gt log short               # Compact view
gt log long                # Detailed view
```

## Branch Management

```bash
gt delete [name]           # Delete branch, restack children onto parent
gt rename [name]           # Rename branch and update metadata
gt fold                    # Fold branch's changes into its parent
gt split                   # Split current branch into multiple branches
gt squash                  # Squash all commits in branch into one
gt track [branch]          # Start tracking an existing branch with Graphite
gt untrack [branch]        # Stop tracking a branch
gt move                    # Rebase current branch onto a different target
gt reorder                 # Interactively reorder branches in the stack
gt pop                     # Delete current branch but keep working tree changes
gt undo                    # Undo the most recent Graphite mutation
```

## Freezing (Exclude from Submit)

```bash
gt freeze [branch]         # Freeze branch + downstack (excluded from submit)
gt unfreeze [branch]       # Unfreeze branch + upstack
```

## Conflict Resolution

```bash
gt continue                # Continue after resolving a rebase conflict
gt abort                   # Abort the current rebase
```

## Branch Info

```bash
gt info [branch]           # Display info about current/specified branch
gt parent                  # Show parent branch
gt children                # Show child branches
gt trunk                   # Show trunk branch
```

## GitHub / Graphite Web

```bash
gt pr [branch]             # Open PR page in browser
gt dash                    # Open Graphite dashboard
gt merge                   # Merge PRs from trunk to current branch via Graphite
gt get [branch]            # Sync a branch/PR from remote
```

## Commit Message Format

Use `[<tag>]` (square brackets) as the prefix tag, **not** `<tag>:` (colon suffix).

```
# ✅ Correct
gt create -m "[chore] add justfile lint comments"
gt create -m "[feat] add trace export"
gt create -m "[fix] correct span attribute name"
gt create -m "[ci] add branch protection rules"

# ❌ Wrong
gt create -m "chore: add justfile lint comments"
gt create -m "feat: add trace export"
```

Common tags: `[feat]`, `[fix]`, `[chore]`, `[ci]`, `[docs]`, `[refactor]`, `[test]`

## PR Description Format

Use this outline for every PR description:

```markdown
# Why?
...

# Before
...

# After
...

# Test
...
```

**Section rules:**
- **Why** — Brief and precise; state the high-level goal only.
- **Before** — Brief code snippet showing the prior behavior or state.
- **After** — Brief code snippet showing the new behavior or state.
- **Test** — What was verified; usually which unit tests pass (e.g. `bazel test //path/to:target`).

**Example:**

```markdown
# Why?
Rename the skill so `/graphite` matches the Graphite CLI workflow it documents.

# Before
    ## Pull requests — always use `/stacked-git-workflow`

# After
    ## Pull requests — always use `/graphite`

# Test
n/a (docs-only change)
```

### Updating PR descriptions (`gt` vs `gh`)

**Default:** set the Why / Before / After / Test body when you create or first submit with Graphite (`gt create`, `gt submit`). That is the normal path.

**Use `gh` when the GitHub body is wrong and `gt` will not fix it**, for example:
- Cursor Bugbot replaced the body with a `<!-- CURSOR_SUMMARY -->` block
- `gt submit --no-edit` pushed code but left the old description on GitHub
- `gt submit --edit-description` is blocked (non-interactive / `CURSOR_AGENT=1` skips the editor)
- Editing `.git/.graphite_pr_info` locally did not stick after `gt submit`

In those cases, **overwrite the body on GitHub with `gh`**. Prefer the REST API — `gh pr edit --body-file` can fail with a GraphQL Projects (classic) deprecation error:

```bash
# One PR
gh api --method PATCH "repos/OWNER/REPO/pulls/PR_NUMBER" \
  -f "body=$(cat path/to/body.md)"

# Stack (example)
for pr in 53 54 55; do
  gh api --method PATCH "repos/tcw165/my-oops-stacks/pulls/${pr}" \
    -f "body=$(cat /tmp/pr-bodies/${pr}.md)"
done
```

**Do not** use `gh` for create/update/submit/rebase — keep using `gt` for all of that. `gh` here is only for syncing the description text on an existing PR, or recovering from a draft PR:

```bash
# If a PR was already created as draft (e.g. submit without --publish)
gh pr ready
```

After a manual `gh` fix, optionally sync local Graphite metadata so a later `gt submit` does not resurrect the old body:

```bash
# Update .git/.graphite_pr_info "body" for each PR, then:
gt submit --stack --no-edit
```

---

## ⚠️ STRICT WORKFLOW - Always Follow This Flow

**Never use raw git commands like `git commit` or `git push origin main`. Always use Graphite.**

### Step 1: Stage Changes with Git
```bash
git add <changes>
```
This is the ONLY place you use raw `git` - to stage files.

### Step 2: Commit with Graphite
**For a new PR:**
```bash
gt create -m "description of changes"
```

**For changes to an existing PR:**
```bash
gt modify
```

### Step 3: Submit to Cloud with Graphite
```bash
gt submit --stack --publish
```

This creates PR(s) on GitHub + Graphite dashboard with the PR URL. Changes are NOT in trunk (`main`) until the PR is reviewed and merged. **Share the Graphite link** from the submit output with the user.

---

**Key Rules:**
- ✅ Use `git add` to stage
- ✅ Use `gt create` for new PRs
- ✅ Use `gt modify` for existing PRs
- ✅ Use `gt submit --stack --publish` to push (ready for review, not draft)
- ❌ Never use `git commit`
- ❌ Never use `git push origin main`
- ❌ Never use `git commit --amend`

---

## Workflow Scenarios

### Scenario 1: New Feature (Single PR)

```bash
# Step 1: Make changes
echo "new code" >> src/feature.ts

# Step 1: Stage changes
git add src/feature.ts

# Step 2: Create PR with Graphite (NOT git commit!)
gt create -m "add new feature"

# Step 3: Submit to cloud (ready for review)
gt submit --publish

# PR URL shown in terminal — share for review
# After approval & merge on GitHub, changes appear in main
```

### Scenario 2: Feature Stack (Multiple Related PRs)

```bash
# Start from trunk
gt checkout main

# PR 1: Foundation
echo "base API" >> src/api.ts
git add src/api.ts
gt create -m "add base API endpoint"

# PR 2: Stack on top of PR 1
echo "business logic" >> src/logic.ts
git add src/logic.ts
gt create -m "add business logic"

# PR 3: Stack on top of PR 2
echo "ui component" >> src/ui.ts
git add src/ui.ts
gt create -m "add UI component"

# Step 3: Submit entire stack at once (ready for review)
gt submit --stack --publish

# All three PRs created in dependency chain
# Review & merge in order: PR1 → PR2 → PR3
```

### Scenario 3: Updating an Existing PR

```bash
# Make changes
echo "bug fix" >> src/feature.ts

# Step 1: Stage changes
git add src/feature.ts

# Step 2: Update current PR with Graphite (NOT git commit --amend!)
gt modify

# Step 3: Resubmit
gt submit --publish
```

### Scenario 4: Reviewing & Merging

```bash
# View your stacks
gt log

# Open PR in browser for review
gt pr

# After approval on GitHub, merge via Graphite or GitHub UI
# Then sync locally
gt sync

# Cleanup: gt sync removes merged branches automatically
```

---

## Typical Full Workflow

```bash
# Start from trunk
gt checkout main

# PR 1: API foundation
# (make changes)
git add src/api.ts
gt create -m "add API endpoint"

# PR 2: Frontend (stacked on PR 1)
# (make changes)
git add src/frontend.ts
gt create -m "add frontend for API"

# PR 3: Documentation (stacked on PR 2)
# (make changes)
git add docs/api.md
gt create -m "add API docs"

# Step 3: Submit entire stack
gt submit --stack --publish

# After review feedback, update the PR that needs changes
gt checkout "add API endpoint"
# (make changes)
git add src/api.ts
gt modify

# Resubmit the updated PR and its stack
gt submit --publish

# After all PRs are merged on GitHub, sync trunk
gt sync
```

## Tips

- Always use `gt sync` before starting new work to stay up to date
- Use `gt log` frequently to visualize your stack
- Prefer small, focused branches — that's the whole point of stacking
- **Always use `gt modify` for updates** (never `git commit --amend`)
- **Always use `gt create` for new PRs** (never `git commit`)
- **Always use `gt submit --stack --publish`** (never `git push origin main`)
- Use `gt submit --stack --publish` to push the entire stack at once (ready for review)
- **Always share the Graphite PR link** from `gt submit` output with the user
- If a rebase conflicts, resolve and run `gt continue`
- Use `gt undo` if something goes wrong
