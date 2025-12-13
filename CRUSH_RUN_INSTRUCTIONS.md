# Crush Run Instructions

## Logging Requirements

When executing tasks through Crush, please adhere to the following logging requirements:

1. **Log File Location**: Write all logs to `.taskmaster/<tag-name>/<task-or-subtask-number>.log`
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