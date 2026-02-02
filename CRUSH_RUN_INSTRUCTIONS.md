# Task Execution: {{.TaskID}} - {{.Title}}

## Task Type Detection

**Task ID Format**: `{{.TaskID}}`
- **Parent Task**: Numeric ID only (e.g., `1`, `2`, `3`) - Has subtasks to complete
- **Subtask**: Dotted notation (e.g., `1.1`, `1.2`, `2.3.1`) - Direct implementation task

## Task Master Integration

**IMPORTANT**: Before starting work, retrieve full task details from Task Master:

1. **Get Task Details**: Run `task-master show {{.TaskID}}` to get complete task information
2. **Check Dependencies**: Verify all dependencies are completed before starting
3. **Update Status**: Mark task as in-progress: `task-master set-status --id={{.TaskID}} --status=in-progress`

## Dependency Integration Requirements

**When a task depends on previous tasks, you MUST:**

1. **Identify Deliverables from Dependencies**
   - For each dependency, run: `task-master show <dependency-id>`
   - Document what components/functions/interfaces were created
   - Example: If dependency is "Task 5: Build Dialog Component", identify the component name and constructor

2. **Use Existing Components (Do Not Reimplement)**
   - Search the codebase for components created in dependency tasks
   - **REQUIRED**: Import and instantiate existing components rather than reimplementing
   - Example: If a dialog was built in Task 5, you MUST use `NewXXXDialog()` to instantiate it
   - **NEVER** take shortcuts by implementing simpler alternatives that bypass existing work

3. **Verify Integration**
   - After implementation, verify that dependency components are actually being used
   - Check that function calls match the interface defined in dependency tasks
   - Confirm no duplicate/alternative implementations were created

## Task Details

**Task ID**: {{.TaskID}}
**Priority**: {{.Priority}}
{{if .TagName}}**Tag Context**: {{.TagName}}{{end}}
{{if .Dependencies}}**Dependencies**: {{.Dependencies}}{{end}}

### Description

{{.Description}}

### Implementation Details

{{.Details}}

### Test Strategy

{{.TestStrategy}}

---

## Execution Strategy

### If This is a Subtask (ID contains dots: X.Y or X.Y.Z):

**This is a direct implementation task. Proceed immediately with implementation:**

1. **Start Implementation**
   - Mark as in-progress: `task-master set-status --id={{.TaskID}} --status=in-progress`
   - Create log file: `.taskmaster/{{if .TagName}}{{.TagName}}{{else}}logs{{end}}/{{.TaskID}}.log`
   - Follow the implementation details and test strategy provided above

2. **Implement the Subtask**
   - Write code according to the specifications above
   - Adhere to the test strategy
   - Document your work in the log file as you progress

3. **Log Progress**
   - Update Task Master with detailed implementation notes:
     ```bash
     task-master update-subtask --id={{.TaskID}} --prompt="Implementation completed:
     - [Specific features/functions implemented]
     - [Files modified and why]
     - [Any challenges encountered and solutions]
     - [Test results and coverage]
     - [Integration points with other tasks]"
     ```

4. **Implementation Verification Requirements**

   **Before marking any task complete, you MUST verify:**

   a. **User-Facing Functionality (if applicable)**
      - If the task adds UI features, document the keyboard shortcut or menu path
      - Describe what the user should see when they trigger the feature
      - **REQUIRED**: Explain how to manually test the feature in the application
      - Example: "Press Ctrl+Shift+F to open the File Changes Dialog. The dialog should display a three-panel layout showing file tree, preview, and filter options."

   b. **Integration Points Checklist**
      - [ ] All components from dependency tasks are imported and used
      - [ ] No simplified/alternative implementations created to bypass existing work
      - [ ] Feature is accessible through documented keyboard shortcuts or UI paths
      - [ ] Feature behavior matches the specification in task details

   c. **Not Just Build Success**
      - ✅ Tests passing is necessary but NOT sufficient
      - ✅ You must also verify the feature works as described in the PRD
      - ❌ Do not consider implementation complete if you created a "stub" or simplified version

5. **Testing Requirements (Enhanced)**

   **Three levels of verification required:**

   a. **Unit Tests**
      - All new functions have passing unit tests
      - Code builds without errors

   b. **Integration Tests**
      - Verify components from dependency tasks are properly integrated
      - Confirm no mocks or stubs are masking missing functionality

   c. **User Acceptance Criteria**
      - **CRITICAL**: Verify the feature works from a user's perspective
      - Document how to manually test the feature
      - If you cannot describe how a user would access and use the feature, the task is NOT complete
      - Example test plan: "User opens TUI → Selects a task → Presses Ctrl+Shift+F → Dialog appears with file tree"

6. **Completion Checklist (MANDATORY)**

   **Before marking task as done, confirm ALL of these:**

   - [ ] Reviewed task details and description thoroughly
   - [ ] Checked all dependency tasks and documented their deliverables
   - [ ] Used existing components from dependencies (no reimplementation)
   - [ ] Implemented complete functionality (no simplified shortcuts)
   - [ ] Verified user-facing behavior matches specification
   - [ ] Can explain to a human how to test the feature
   - [ ] Wrote comprehensive log documenting integration points
   - [ ] All tests pass (unit + integration)

   **If ANY checkbox is unchecked, the task is NOT complete.**

7. **Complete Subtask**
   - Verify all requirements met from steps 4-6 above
   - Update task-master with completion notes including verification results:
     ```bash
     task-master update-subtask --id={{.TaskID}} --prompt="Subtask completion summary:
     - [What was implemented]
     - [Test results - all three levels]
     - [Manual testing procedure and results]
     - [Verification checklist confirmation]
     - [Integration with dependency components confirmed]
     - [Any notes for future reference]"
     ```
   - Mark as done: `task-master set-status --id={{.TaskID}} --status=done`
   - Finalize log file with summary

### If This is a Parent Task (ID is numeric only: X):

**This task has subtasks. Follow the systematic subtask completion workflow:**

**Note**: DO NOT implement parent tasks directly. They are organizational containers. Work through subtasks sequentially.

1. **Review Subtasks**: Run `task-master show {{.TaskID}}` to see all subtasks and their current status

#### For Each Subtask (in consecutive order):

1. **Retrieve Subtask Details**

   - Run `task-master show {{.TaskID}}.X` (where X is the subtask number, e.g., 1.1, 1.2, etc.)
   - **Read and understand the complete subtask information**:
     - Subtask title and description
     - Implementation details specific to this subtask
     - Test strategy for this subtask
     - Any dependencies this subtask has
   - Review any existing notes or updates from previous work
   - Understand how this subtask fits into the overall task

2. **Start Subtask**

   - Mark as in-progress: `task-master set-status --id={{.TaskID}}.X --status=in-progress`
   - Create log file: `.taskmaster/{{if .TagName}}{{.TagName}}{{else}}logs{{end}}/{{.TaskID}}.X.log`

3. **Implement Subtask**

   - Follow the subtask's specific implementation details retrieved in step 1
   - Write code according to the subtask specifications
   - Adhere to the subtask's test strategy
   - Document your work in the log file as you progress

4. **Log Subtask Progress**

   - Update Task Master with detailed implementation notes:
     ```bash
     task-master update-subtask --id={{.TaskID}}.X --prompt="Implementation completed:
     - [Specific features/functions implemented]
     - [Files modified and why]
     - [Any challenges encountered and solutions]
     - [Test results and coverage]
     - [Integration points with other subtasks]"
     ```

5. **Implementation Verification Requirements**

   **Before marking any subtask complete, you MUST verify:**

   a. **User-Facing Functionality (if applicable)**
      - If the subtask adds UI features, document the keyboard shortcut or menu path
      - Describe what the user should see when they trigger the feature
      - **REQUIRED**: Explain how to manually test the feature in the application
      - Example: "Press Ctrl+Shift+F to open the File Changes Dialog. The dialog should display a three-panel layout showing file tree, preview, and filter options."

   b. **Integration Points Checklist**
      - [ ] All components from dependency tasks are imported and used
      - [ ] No simplified/alternative implementations created to bypass existing work
      - [ ] Feature is accessible through documented keyboard shortcuts or UI paths
      - [ ] Feature behavior matches the specification in task details

   c. **Not Just Build Success**
      - ✅ Tests passing is necessary but NOT sufficient
      - ✅ You must also verify the feature works as described in the PRD
      - ❌ Do not consider implementation complete if you created a "stub" or simplified version

6. **Testing Requirements (Enhanced)**

   **Three levels of verification required:**

   a. **Unit Tests**
      - All new functions have passing unit tests
      - Code builds without errors

   b. **Integration Tests**
      - Verify components from dependency tasks are properly integrated
      - Confirm no mocks or stubs are masking missing functionality

   c. **User Acceptance Criteria**
      - **CRITICAL**: Verify the feature works from a user's perspective
      - Document how to manually test the feature
      - If you cannot describe how a user would access and use the feature, the task is NOT complete
      - Example test plan: "User opens TUI → Selects a task → Presses Ctrl+Shift+F → Dialog appears with file tree"

7. **Completion Checklist (MANDATORY)**

   **Before marking subtask as done, confirm ALL of these:**

   - [ ] Reviewed subtask details and description thoroughly
   - [ ] Checked all dependency tasks and documented their deliverables
   - [ ] Used existing components from dependencies (no reimplementation)
   - [ ] Implemented complete functionality (no simplified shortcuts)
   - [ ] Verified user-facing behavior matches specification
   - [ ] Can explain to a human how to test the feature
   - [ ] Wrote comprehensive log documenting integration points
   - [ ] All tests pass (unit + integration)

   **If ANY checkbox is unchecked, the subtask is NOT complete.**

8. **Complete Subtask**

   - Verify all subtask requirements met from steps 5-7 above
   - Update task-master subtask details with completion notes including verification results:
     ```bash
     task-master update-subtask --id={{.TaskID}}.X --prompt="Subtask completion summary:
     - [What was implemented]
     - [Test results - all three levels]
     - [Manual testing procedure and results]
     - [Verification checklist confirmation]
     - [Integration with dependency components confirmed]
     - [Any notes for future reference]"
     ```
   - Mark as done: `task-master set-status --id={{.TaskID}}.X --status=done`
   - Finalize log file with summary

9. **Move to Next Subtask**
   - **Retrieve next subtask details**: `task-master show {{.TaskID}}.Y` (where Y = X + 1)
   - Repeat steps 1-8 for the next subtask
   - **Do NOT skip subtasks or work on them out of order**
   - **Always retrieve fresh details for each subtask before starting**

#### After All Subtasks Complete:

1. **Review All Subtask Logs**

   - Read through all subtask completion notes
   - Identify common themes, challenges, or patterns
   - Prepare comprehensive summary

2. **Update Parent Task**

   - Run `task-master update-task --id={{.TaskID}} --prompt="All subtasks completed:
     - [Summary of what was accomplished across all subtasks]
     - [Overall integration and how subtasks work together]
     - [Cumulative challenges and solutions]
     - [Complete test results including integration tests]
     - [Any follow-up items or technical debt noted]
     - [Recommendations for future work]"`

3. **Mark Parent Task Complete**
   - Verify all subtasks are marked as done
   - Ensure all tests pass
   - Update parent task details with final completion notes:
     ```bash
     task-master update-task --id={{.TaskID}} --prompt="Task completion summary:
     - [Overall accomplishments]
     - [Final test results]
     - [Any follow-up items or recommendations]"
     ```
   - Mark parent as done: `task-master set-status --id={{.TaskID}} --status=done`
   - Create final summary log: `.taskmaster/{{if .TagName}}{{.TagName}}{{else}}logs{{end}}/{{.TaskID}}.log`

---

## Logging Requirements

When executing this task through Crush, please adhere to the following logging requirements:

1. **Log File Location**: Write all logs to `.taskmaster/{{if .TagName}}{{.TagName}}{{else}}logs{{end}}/<task-or-subtask-number>.log`

   - Create the directory structure if it doesn't exist
   - Example: `.taskmaster/feature-auth/1.2.1.log`

2. **Log Content**: Include the following information in each log:

   - **Thought Process**: Document your reasoning and decision-making
   - **Intentions**: Explain what you're trying to accomplish with each change
   - **Code Changes**: Record all modifications made to the codebase:
     - File paths
     - Before/after snippets
     - Explanation for each change

3. **Log Format**:

   ```
   # Task: [Task Number] - [Task Title]
   # Date: [Current Date/Time]

   ## Thought Process
   [Detailed reasoning about the approach taken]

   ## Implementation Plan
   [Step-by-step plan for completing the task]

   ## Code Changes

   ### [File Path 1]
   ```diff
   - [removed code]
   + [added code]
   ```
   Reason: [Explanation for this specific change]

   ### [File Path 2]

   ...

   ## Challenges and Solutions

   [Any obstacles encountered and how they were overcome]

   ## Testing

   [Tests run and their results]

   ## Summary

   [Brief overview of what was accomplished]

   ```

4. **Automatic Logging**: Initiate logging automatically at the beginning of each task

   - Create required directories if they don't exist
   - Use task number from the current context if available

5. **Directory Structure**:
   - Use semantic tag names for features/bugfixes
   - Log files should be named after the exact task or subtask number

This logging approach ensures comprehensive documentation of all development activities performed by Crush, making it easier to review changes and understand the reasoning behind them.
