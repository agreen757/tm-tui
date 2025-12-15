# TaskMaster Workflow Guide - Optimized for Agent Execution

## Pre-Task Checklist (5 minutes)

Run these commands to load context once per task:

```bash
task-master show <task-id>                    # Load full task context
./bin/memory get -key "readme:taskmaster-tui" # Project overview (cache)
./bin/memory list -prefix "log:<task-id>"     # Previous attempts on this task
```

**Stop if**: Task is blocked, missing dependencies, or requires external access.

---

## Implementation Loop

### Phase 1: Plan (No coding yet)
1. Read task requirements
2. Identify what files to change (use grep/find to verify they exist)
3. Check memory for similar completed tasks: `./bin/memory list | grep -i "<topic>"`
4. Note down 3-5 specific code locations that need changes
5. Log plan once (don't repeat): `./bin/memory log -task "task-master:<task-id>" -message "PLAN: [specific steps, files, line numbers]"`

### Phase 2: Execute (Code changes)
1. Make changes following the logged plan
2. Run tests immediately after each logical change
3. If tests fail: fix root cause, don't skip
4. If stuck > 15 mins on one issue: log blockers and try different approach

### Phase 3: Verify (Before marking done)
- [ ] All code changes match the plan
- [ ] Tests pass (run full suite once)
- [ ] No broken imports or unrelated bugs introduced
- [ ] Changes follow project patterns (check similar files)

---

## Subtask Completion Template

When marking subtask as done, do both steps:

**Step 1: Update task-master**
```bash
task-master update-subtask --id=<id> --prompt="WHAT: [1 sentence]. HOW: [approach/pattern used]. TESTS: [what was verified]. CODE: <file:line> <file:line>"
task-master set-status --id=<id> --status=done
```

**Step 2: Update memory with implementation details**
```bash
./bin/memory log -task "task-master:<id>" -message "COMPLETE: <subtask-id> - <title>

SUMMARY: [2-3 sentence description of what was implemented]

APPROACH: [design pattern/strategy used and why]

CODE CHANGES:
[Include actual code snippet showing the main change]

FILES MODIFIED: <file1:line-range>, <file2:line-range>, ...

TESTS: [which tests verify this, pass rate]

RATIONALE: [why this approach over alternatives, performance impact, maintainability]

DEPENDENCIES: [what this subtask depends on, what depends on it]

BLOCKERS RESOLVED: [if any issues were encountered and solved, document the fix]"
```

**Example:**
```bash
task-master update-subtask --id=2.1 --prompt="WHAT: Added JWT validation middleware. HOW: Wrapped routes with verifyToken() from auth package. TESTS: 5 auth tests pass. CODE: internal/auth/jwt.go:45-67 internal/routes/api.go:12"
task-master set-status --id=2.1 --status=done

./bin/memory log -task "task-master:2.1" -message "COMPLETE: 2.1 - Add JWT validation middleware

SUMMARY: Implemented token validation middleware that intercepts all protected routes, validates JWT signature and expiration, and extracts user claims for downstream handlers.

APPROACH: Middleware pattern (ChainHandler wrapper). Chose this over per-route guards to centralize auth logic and reduce duplication. Stateless design allows horizontal scaling.

CODE CHANGES:
func VerifyTokenMiddleware(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    token := extractToken(r)
    claims, err := validateToken(token)
    if err != nil {
      http.Error(w, \"Unauthorized\", http.StatusUnauthorized)
      return
    }
    r.Header.Set(\"user-id\", claims.UserID)
    next.ServeHTTP(w, r)
  })
}

FILES MODIFIED: internal/auth/jwt.go:45-89, internal/routes/api.go:12-18

TESTS: 5/5 auth middleware tests pass. Edge cases covered: expired token, invalid signature, missing token.

RATIONALE: Middleware pattern is Go standard for cross-cutting concerns. Avoids code duplication across 20+ route handlers. Stateless validation enables caching at load balancer if needed.

DEPENDENCIES: Depends on JWT secret setup (task 2.0). Unblocks task 2.2 (RBAC layer).

BLOCKERS RESOLVED: Initially used per-route guards but realized maintenance burden; refactored to middleware (saved 150 LOC)."
```

---

## Main Task Completion Template

After all subtasks are done, do both steps:

**Step 1: Update task-master**
```bash
task-master update-task --id=<id> --prompt="SUMMARY: [2-3 sentences on what was built]. FILES: <5-10 changed files>. TESTS: [coverage/pass rate]. DECISIONS: [why this approach over alternatives]. ISSUES: [unresolved or future work]"
task-master set-status --id=<id> --status=done
```

**Step 2: Update memory with comprehensive implementation details**
```bash
./bin/memory log -task "task-master:<id>" -message "COMPLETE: <task-id> - <title>

SUMMARY: [3-5 sentence comprehensive description of the entire feature/implementation]

ARCHITECTURE: [high-level design, how components interact, data flow]

IMPLEMENTATION DETAILS:
[Include key code snippets showing the main logic]

FILES CHANGED:
- internal/auth/jwt.go (45-89): Token validation logic
- internal/auth/rbac.go (12-156): Role-based access control
- internal/routes/api.go (5-35): Middleware integration
- tests/auth_test.go (1-250): 18 test cases
- config/auth.yaml: Configuration additions

SUBTASKS COMPLETED: 2.1 (jwt middleware), 2.2 (rbac), 2.3 (tests)

TESTING SUMMARY:
- Unit tests: 18/18 pass (100%)
- Integration tests: 7/7 pass
- Edge cases covered: token expiry, invalid signatures, missing permissions
- No regressions in existing auth paths

KEY DECISIONS:
1. JWT over sessions: Chose for stateless scaling vs session-based approach
   - Pro: Horizontally scalable, no server state
   - Con: Cannot revoke tokens immediately (mitigated with blacklist)
   
2. Middleware pattern: Centralized auth validation
   - Avoided per-route guards (150+ LOC saved, easier maintenance)
   
3. Role-based RBAC: Extensible permission model
   - Supports dynamic role creation, inherited permissions

RATIONALE: JWT is industry standard for modern APIs; middleware pattern follows Go conventions; RBAC allows enterprise customers to define custom roles.

PERFORMANCE IMPACT:
- Token validation: ~2ms per request (negligible)
- Memory: ~500KB for role cache (minimal)
- No performance regression in baseline tests

SECURITY REVIEW:
- Tokens signed with HS256 (adequate for internal use; RS256 recommended for external APIs)
- Passwords hashed with bcrypt (cost factor 12)
- Token expiry: 24h (configurable)
- Blacklist check bypassed if Redis unavailable (safe fail: rejects token)

BLOCKERS RESOLVED:
- Database migration timing: solved by lazy-loading roles
- Token refresh logic: implemented sliding window approach

FUTURE IMPROVEMENTS:
- Redis token blacklist (currently in-memory, doesn't survive restarts)
- Switch to RS256 if integrating external OAuth providers
- Implement 2FA for admin users

DEPENDENCIES:
- Unblocks: API endpoint authorization (task 3), user management (task 4)
- Blocked by: Database schema (completed in task 1)"
```

**Example:**
```bash
task-master update-task --id=2 --prompt="SUMMARY: Implemented JWT-based authentication system with role-based access control. FILES: internal/auth/jwt.go, internal/auth/rbac.go, internal/routes/api.go, tests/auth_test.go. TESTS: 18/18 pass (100% coverage on auth paths). DECISIONS: Used JWT over sessions for stateless scaling; bcrypt for hashing due to industry standard. ISSUES: Redis caching for token blacklist marked as future optimization."
task-master set-status --id=2 --status=done

./bin/memory log -task "task-master:2" -message "COMPLETE: 2 - Implement JWT authentication with RBAC

SUMMARY: Completed full JWT-based authentication system with role-based access control. System validates tokens on all protected routes, supports dynamic role creation, and integrates with existing user database. All 18 auth tests pass with 100% coverage on auth paths.

ARCHITECTURE: 
- JWT token generation on login (payload includes user ID, roles, expiry)
- Middleware intercepts all /api/* routes, validates signature and expiry
- RBAC layer maps user roles to permissions, checked before handler execution
- Token blacklist for logout (Redis-backed, in-memory fallback)
- Role cache refreshes every 5 minutes or on explicit update

IMPLEMENTATION DETAILS:
// JWT validation in middleware
func VerifyTokenMiddleware(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    token := extractToken(r)
    claims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
      return jwtSecret, nil
    })
    if err != nil || !token.Valid {
      http.Error(w, \"Unauthorized\", 401)
      return
    }
    r.Header.Set(\"user-id\", claims.UserID)
    r.Header.Set(\"user-roles\", strings.Join(claims.Roles, \",\"))
    next.ServeHTTP(w, r)
  })
}

// RBAC permission check
func RequirePermission(perm string) func(http.Handler) http.Handler {
  return func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      userRoles := strings.Split(r.Header.Get(\"user-roles\"), \",\")
      if !hasPermission(userRoles, perm) {
        http.Error(w, \"Forbidden\", 403)
        return
      }
      next.ServeHTTP(w, r)
    })
  }
}

FILES CHANGED:
- internal/auth/jwt.go (45-120): Token generation and validation logic
- internal/auth/rbac.go (12-156): Permission checking and role management
- internal/routes/api.go (5-35): Middleware mounting and route protection
- tests/auth_test.go (1-250): 18 comprehensive test cases
- configs/auth.yaml: JWT secret, token expiry, role definitions

SUBTASKS COMPLETED:
- 2.1: JWT token generation and validation middleware
- 2.2: RBAC permission checking layer
- 2.3: Integration tests for auth flows

TESTING:
- 18/18 unit tests pass
- 7/7 integration tests pass (login, access granted, access denied scenarios)
- Coverage: 100% on auth paths, 95% overall
- Edge cases: expired token, invalid signature, missing roles, permission denial

KEY DECISIONS:
1. JWT over sessions: Horizontal scalability without server state
2. HS256 signing: Adequate for internal APIs (would use RS256 for external)
3. Middleware pattern: Avoided per-route guards (saved 150+ LOC, easier maintenance)
4. Role cache with 5min TTL: Balance between consistency and performance

RATIONALE: JWT is industry standard for modern APIs; middleware pattern follows Go/HTTP conventions; role caching prevents database hammering; bcrypt is cryptographically sound.

PERFORMANCE: Token validation ~2ms per request, negligible overhead. Role cache ~500KB memory.

SECURITY: HS256 adequate for internal use; bcrypt cost factor 12; token expiry 24h; revocation via Redis blacklist (degrades gracefully).

BLOCKERS RESOLVED: Circular dependency between auth and user migration solved by delaying role initialization.

FUTURE: Redis token blacklist (persistent), consider 2FA for admins, RS256 if adding OAuth."
```

---

## Decision Tree for Task Success

```
Is the task clear?
├─ NO: Update task with questions → mark as blocked
└─ YES: Continue to planning

Do you have all dependencies?
├─ NO: Check task dependencies, mark as blocked if missing
└─ YES: Continue to planning

Can you identify specific files to change?
├─ NO: Search codebase, read similar implementations
└─ YES: Log plan and execute

Tests passing?
├─ NO: Debug, fix root cause (don't skip)
└─ YES: Update task and mark done
```

---

## Context Optimization Rules

**DO THIS:**
- Load context once per task with `task-master show` (use result for entire task)
- Check memory for similar completed tasks before implementing
- Log plan once at start (reference it, don't repeat it)
- Use grep to verify files exist before editing
- Run tests incrementally (after each file change)

**DON'T DO THIS:**
- Re-run `task-master show` multiple times (reuse initial output)
- Log progress after every single line of code
- Ask clarifying questions without searching codebase first
- Make changes without understanding project patterns
- Skip tests or test only at the very end

---

## Emergency Recovery (If blocked > 30 mins)

```bash
# 1. Dump task state
task-master show <id>

# 2. Check what similar tasks did
./bin/memory list | grep -i "<topic>"

# 3. Verify git state
git status
git diff HEAD

# 4. Log blockers (be specific)
./bin/memory log -task "task-master:<id>" -message "BLOCKER: [specific error/missing dependency]. ATTEMPTED: [what was tried]. NEXT: [different approach to try]"

# 5. Try different approach:
#    - Change tech stack approach
#    - Refactor to simpler design
#    - Split into smaller subtasks
#    - Ask for external help with precise error output
```

---

## Memory Management Strategy

Memory logs (`./bin/memory log`) are persistent knowledge that survives across sessions and tasks. Use them to create a searchable implementation archive.

### When to Log

- ✅ **After completing each subtask**: Capture implementation details while fresh
- ✅ **After completing main task**: Full summary for future reference
- ✅ **When solving blockers**: Document the problem and solution
- ✅ **When discovering patterns**: Record reusable code patterns
- ✅ **When debugging**: Log root cause analysis and fix

### What NOT to Log

- ❌ **Duplicate what's in task-master**: Don't repeat summary from task update
- ❌ **Incomplete thoughts**: Only log when you have concrete details
- ❌ **Trivial changes**: Skip minor formatting fixes or comments
- ❌ **Every single line**: Summarize instead of blow-by-blow narration

### Memory Organization

Memory entries use this key structure:

```
log:task-master:<task-id>           # Main entry for task completion
log:task-master:<subtask-id>        # Individual subtask implementations
workflow:guide                      # This document
readme:<topic>                      # Topic-specific documentation (cached)
```

### Querying Memory for Context

Before starting a task, search for related implementations:

```bash
# Find task you're working on
./bin/memory get -key "log:task-master:<task-id>"

# Find similar completed tasks
./bin/memory list | grep "log:task-master:" | head -20
./bin/memory get -key "log:task-master:2" | grep -i "auth"

# Check if pattern exists
./bin/memory list | grep "rbac\|permission\|role"
```

### Memory Entry Structure (Complete Template)

Every subtask and main task completion should follow this structure:

```
COMPLETE: <task-id> - <short title>

SUMMARY: [2-3 sentences describing what was implemented]

APPROACH: [Design pattern, strategy, why this approach]

CODE CHANGES:
[Actual code snippet showing main logic - not pseudocode]

FILES MODIFIED:
- file1.go (line-range): Description
- file2.go (line-range): Description

TESTS:
- Test suite name: X/Y pass
- Coverage: XX%
- Edge cases covered: [list specific scenarios]

RATIONALE:
- Why this approach over alternatives
- Performance characteristics
- Maintainability and future-proofing

BLOCKERS RESOLVED:
[Any problems encountered and how they were solved]

DEPENDENCIES:
- Depends on: [tasks this needs]
- Enables: [tasks this unblocks]

FUTURE IMPROVEMENTS:
[Optimizations, tech debt, or nice-to-haves]
```

### Real-world Example: Querying for Implementation Help

```bash
# Agent working on RBAC feature (task 3.2)
./bin/memory get -key "log:task-master:2.2"

# Output includes RBAC implementation from task 2.2
# Agent can reuse patterns: permission checking middleware, role cache strategy, etc.

# Find all security-related implementations
./bin/memory list | xargs -I {} ./bin/memory get -key "{}" | grep -i "secret\|hash\|encrypt" | head -20
```

---

## Token Efficiency Guide

**Keep context loaded in one session** (avoid task switching):
- One task = one agent session
- Use `/clear` only between different major tasks
- Reference earlier outputs instead of re-running commands

**Optimize memory queries (order by speed):**
```bash
# FASTEST: Exact key lookup (O(1))
./bin/memory get -key "log:task-master:2.1"

# FAST: List all then filter with grep
./bin/memory list | grep "log:task-master:2"

# SLOWER: Prefix search (if available)
./bin/memory list -prefix "log:task-master:2"

# Use xargs for batch operations (only when needed)
./bin/memory list | xargs -I {} ./bin/memory get -key "{}"
```

**Memory best practices for token efficiency:**
- Load memory entry once per task (cache result, don't re-query)
- Reference the cached output when making decisions
- Add snippets to subtask/task updates incrementally (not all at once)
- Use relative file paths in code snippets (easier to parse, fewer tokens)

**Code changes:**
- Make one logical change per commit
- Include task reference in commit message: `git commit -m "feat: add JWT validation (task 2.1)"`
- In memory log, use relative paths: `internal/auth/jwt.go` not `/Users/.../internal/auth/jwt.go`
- Use line ranges not full file: `(45-89)` not entire 200-line file

**Reduce redundant updates:**
- Update subtask once with full details (not incremental updates)
- Update main task once with comprehensive summary
- Memory logs are permanent; include all details upfront
- Don't repeat what's already documented (reference it instead)

**Efficient memory logging:**
```bash
# GOOD: One comprehensive log after subtask completion
./bin/memory log -task "task-master:2.1" -message "COMPLETE: 2.1 - Middleware
SUMMARY: [full description]
CODE: [snippet]
FILES: [list with line ranges]
TESTS: [results]
RATIONALE: [why this approach]"

# BAD: Multiple incremental updates (wastes space, confuses future readers)
./bin/memory log -task "task-master:2.1" -message "Step 1: Added VerifyToken function..."
./bin/memory log -task "task-master:2.1" -message "Step 2: Integrated middleware..."
./bin/memory log -task "task-master:2.1" -message "Step 3: Added tests..."
```

---

## Success Metrics

Task completed successfully when:
- ✅ All subtasks marked `done` with descriptive notes
- ✅ Main task updated with summary and file references
- ✅ All tests pass
- ✅ Code follows project patterns
- ✅ No external blockers remain
