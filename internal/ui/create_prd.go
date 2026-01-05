package ui

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"os/exec"
	"regexp"

	"github.com/agreen757/tm-tui/internal/config"
	"github.com/agreen757/tm-tui/internal/pathutil"
	"github.com/agreen757/tm-tui/internal/ui/dialog"
	tea "github.com/charmbracelet/bubbletea"
)

// openCreatePrdWorkflow initiates the PRD creation workflow.
// Starts with showing the PRD input form dialog.
func (m *Model) openCreatePrdWorkflow() tea.Cmd {
	m.addLogLine("Create PRD workflow initiated (Alt+Shift+P)")

	// Check if in a Task Master workspace
	if m.config.TaskMasterPath == "" {
		m.showErrorDialog("Task Master Workspace Required", "Not in a Task Master workspace. Please open a Task Master project first.")
		return nil
	}

	// Initialize PRD creation state in both Model and AppState
	if m.prdCreationState == nil {
		m.prdCreationState = NewPrdCreationState()
	}
	m.appState.InitPrdCreationState() // Also initialize in AppState for dialog navigation

	// Show PRD input dialog
	return m.showPrdInputDialog()
}

// showPrdInputDialog displays the PRD input form and handles submission or cancellation
func (m *Model) showPrdInputDialog() tea.Cmd {
	inputDialog := NewPrdInputDialog(m.config)
	dm := m.dialogManager()
	if dm != nil {
		dialog.ApplyStyleToDialog(inputDialog, dm.Style)
	}

	m.appState.AddDialog(inputDialog, func(value interface{}, err error) tea.Cmd {
		if err != nil {
			m.showErrorDialog("Create PRD", err.Error())
			return nil
		}

		// Check if form was cancelled
		result, ok := value.(PrdFormValues)
		if !ok {
			return nil
		}

		// Store PRD values in app state for later retrieval
		if m.prdCreationState != nil {
			m.prdCreationState.UpdateFromFormValues(result)
		}

		// Show model selection dialog
		return m.showModelSelectionForPrd()
	})

	return inputDialog.Init()
}

// showModelSelectionForPrd displays the model selection dialog for PRD creation
func (m *Model) showModelSelectionForPrd() tea.Cmd {
	modelDialog := dialog.NewModelSelectionDialogSimple()
	dm := m.dialogManager()
	if dm != nil {
		dialog.ApplyStyleToDialog(modelDialog, dm.Style)
	}

	m.appState.AddDialog(modelDialog, func(value interface{}, err error) tea.Cmd {
		if err != nil {
			m.showErrorDialog("Model Selection", err.Error())
			return nil
		}

		// Check if model was selected
		msg, ok := value.(dialog.ListSelectionMsg)
		if !ok || msg.SelectedItem == nil {
			// If cancelled, return to PRD input dialog with values preserved
			return m.showPrdInputWithState()
		}

		// Extract selected model
		if listItem, ok := msg.SelectedItem.(*dialog.ModelSelectionListItem); ok {
			opt := listItem.GetOption()
			m.addLogLine("Model selected: " + opt.DisplayName)

			// Save model configuration
			if err := config.SaveModelConfig(opt.Provider, opt.ModelID); err != nil {
				m.showErrorDialog("Save Model Configuration", "Failed to save model selection: "+err.Error())
				return nil
			}

			m.addLogLine("Model configuration saved")
			
			// Execute Crush for PRD creation with selected model
			return m.executeCrushForPrd(opt.Provider, opt.ModelID)
		}

		return nil
	})

	return modelDialog.Init()
}

// showPrdInputWithState displays the PRD input dialog with previously entered values preserved
func (m *Model) showPrdInputWithState() tea.Cmd {
	// Use existing state to pre-fill form
	var inputDialog *dialog.FormDialog
	if m.prdCreationState != nil && !m.prdCreationState.IsEmpty() {
		inputDialog = NewPrdInputDialogWithState(m.config, m.prdCreationState)
	} else {
		inputDialog = NewPrdInputDialog(m.config)
	}

	dm := m.dialogManager()
	if dm != nil {
		dialog.ApplyStyleToDialog(inputDialog, dm.Style)
	}

	m.appState.AddDialog(inputDialog, func(value interface{}, err error) tea.Cmd {
		if err != nil {
			m.showErrorDialog("Create PRD", err.Error())
			return nil
		}

		result, ok := value.(PrdFormValues)
		if !ok {
			return nil
		}

		// Store updated PRD values in app state
		if m.prdCreationState != nil {
			m.prdCreationState.UpdateFromFormValues(result)
		}

		// Show model selection again
		return m.showModelSelectionForPrd()
	})

	return inputDialog.Init()
}

// executeCrushForPrd executes the Crush tool to generate PRD content
// with the saved PRD state values and selected model
func (m *Model) executeCrushForPrd(provider, modelID string) tea.Cmd {
	return func() tea.Msg {
		// Log PRD creation execution with preserved state
		m.addLogLine(fmt.Sprintf("Preparing to generate PRD with %s model", modelID))
		
		if m.prdCreationState != nil {
			m.addLogLine(fmt.Sprintf("PRD Title: %s", m.prdCreationState.Title))
			m.addLogLine(fmt.Sprintf("PRD Summary: %s", m.prdCreationState.Summary))
			m.addLogLine(fmt.Sprintf("Output file: %s", m.prdCreationState.Filename))
		}
		
		// Validate the Crush binary
		m.addLogLine("Validating Crush binary...")
		if err := dialog.ValidateCrushBinary(); err != nil {
			// Check if it's a Crush binary error and handle with specific messaging
			if _, isCrushErr := err.(*dialog.CrushBinaryError); isCrushErr {
				m.showCrushDependencyError(err.Error())
				m.addLogLine("Returning to PRD input form due to Crush binary error")
				// Return to PRD input with state preserved for recovery
				return m.showPrdInputWithState()
			}
			// For other errors, show generic error message and return to form
			m.showErrorDialog("Crush Execution Failed", fmt.Sprintf("Crush execution failed: %v\n\nPlease try again.", err))
			m.addLogLine("Returning to PRD input form due to Crush execution error")
			return m.showPrdInputWithState()
		}
		m.addLogLine("✓ Crush binary validated")
		
		// Generate the PRD prompt from state
		m.addLogLine("Building PRD prompt...")
		state := m.prdCreationState
		prompt := dialog.GenerateCrushPrdPrompt(state.Title, state.Summary, state.Scope)
		m.addLogLine("✓ PRD prompt generated")
		
		// Create and show the modal for streaming output
		// Calculate modal dimensions based on available space
		modalWidth := m.width - 4
		if modalWidth < 40 {
			modalWidth = 40
		}
		if modalWidth > 120 {
			modalWidth = 120
		}
		
		modalHeight := m.height - 6
		if modalHeight < 10 {
			modalHeight = 10
		}
		if modalHeight > 30 {
			modalHeight = 30
		}
		
		dm := m.dialogManager()
		modal := dialog.NewTaskRunnerModal(modalWidth, modalHeight, dm.Style)
		if dm != nil {
			dialog.ApplyStyleToDialog(modal, dm.Style)
		}
		
		// Reset output buffer for this run
		state.OutputBuffer.Reset()
		state.InProgress = true
		
		// Add dialog with callback to handle completion
		m.appState.AddDialog(modal, func(value interface{}, err error) tea.Cmd {
			state.InProgress = false
			
			if err != nil {
				m.showErrorDialog("PRD Generation", "Execution was cancelled or failed")
				return nil
			}
			
			// PRD generation completed successfully
			m.addLogLine("✓ PRD generation completed")
			
			// Validate PRD output content
			if !m.validatePrdOutput(state) {
				return nil
			}
			
			// Store the generated content in state
			state.GeneratedContent = state.OutputBuffer.String()
			
			// Proceed to file save
			return m.savePrdToFile()
		})
		
		// Start Crush execution in background
		go func() {
			// Execute Crush with streaming output to buffer and modal
			execErr := m.executeCrushSubprocessWithOutput(modelID, prompt, modal, state.OutputBuffer)
			
			if execErr != nil {
				m.addLogLine(fmt.Sprintf("✗ Crush execution failed: %v", execErr))
				state.InProgress = false
			}
		}()
		
		return modal.Init()
	}
}

// executeCrushSubprocessWithOutput runs the Crush subprocess and streams output to the modal and buffer
func (m *Model) executeCrushSubprocessWithOutput(model, prompt string, modal *dialog.TaskRunnerModal, outputBuffer *bytes.Buffer) error {
	// Use existing Crush runner implementation
	// The ExecuteCrushSubprocess function in the dialog package handles the actual subprocess execution
	// and streams output through BubbleTea messages
	// For PRD generation, we'll call the Crush tool and capture output
	
	ctx := context.Background()
	
	// Create the command - crush run takes the prompt as argument
	cmd := exec.CommandContext(ctx, "crush", "run", prompt)
	
	// Log the exact command being executed
	m.addLogLine(fmt.Sprintf("Executing: crush run %q", prompt))
	
	// Set up stdout and stderr pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		m.addLogLine(fmt.Sprintf("Failed to create stdout pipe: %v", err))
		m.showErrorDialog("Crush Execution Error", fmt.Sprintf("Failed to set up output stream: %v\n\nPlease try again.", err))
		return err
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		m.addLogLine(fmt.Sprintf("Failed to create stderr pipe: %v", err))
		m.showErrorDialog("Crush Execution Error", fmt.Sprintf("Failed to set up error stream: %v\n\nPlease try again.", err))
		return err
	}
	
	// Start the command
	if err := cmd.Start(); err != nil {
		m.addLogLine(fmt.Sprintf("Failed to start Crush: %v", err))
		m.showErrorDialog("Crush Execution Error", fmt.Sprintf("Failed to start Crush process: %v\n\nVerify Crush is installed and accessible in your PATH.", err))
		return err
	}
	
	// Stream stdout to both modal and buffer
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			
			// Write to buffer
			outputBuffer.WriteString(line + "\n")
			
			// Log output for debugging
			m.addLogLine(line)
		}
	}()
	
	// Stream stderr to log
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			
			// Write to buffer with error prefix
			outputBuffer.WriteString("[ERR] " + line + "\n")
			
			// Log error for debugging
			m.addLogLine("[ERR] " + line)
		}
	}()
	
	// Wait for process to complete
	if err := cmd.Wait(); err != nil {
		m.addLogLine(fmt.Sprintf("Crush execution error: %v", err))
		// Don't return error - the process may have completed despite exit code
	}
	
	return nil
}

// stripAnsiCodes removes ANSI escape codes from text
func stripAnsiCodes(text string) string {
	ansiRegex := regexp.MustCompile("\x1b\\[[0-9;]*m")
	return ansiRegex.ReplaceAllString(text, "")
}

// validatePrdOutput validates the generated PRD content for emptiness and minimum length
func (m *Model) validatePrdOutput(state *PrdCreationState) bool {
	// Check if output buffer is empty
	if state.OutputBuffer == nil || state.OutputBuffer.Len() == 0 {
		m.showErrorDialog("PRD Generation Failed", "No PRD content was generated. Please try again.\n\nCheck that Crush binary is installed and accessible.")
		return false
	}

	// Check minimum content length
	content := state.OutputBuffer.String()
	if len(content) < 100 {
		m.showErrorDialog("PRD Generation Failed", "Generated PRD content is too short. Please try again.\n\nThe PRD prompt may have been too brief. Review and expand your PRD summary.")
		return false
	}

	return true
}

// handlePrdExecutionError handles errors during PRD generation and provides recovery options
func (m *Model) handlePrdExecutionError(errorMsg string, recoveryHint string) tea.Cmd {
	// Log the error for debugging
	m.addLogLine(fmt.Sprintf("PRD Execution Error: %s", errorMsg))
	
	// Show error with recovery hint
	fullMessage := errorMsg
	if recoveryHint != "" {
		fullMessage = errorMsg + "\n\n" + recoveryHint
	}
	m.showErrorDialog("PRD Generation Error", fullMessage)
	
	// Return to PRD input dialog with state preserved
	return m.showPrdInputWithState()
}

// savePrdToFile saves the generated PRD content to a file
func (m *Model) savePrdToFile() tea.Cmd {
	return func() tea.Msg {
		if m.prdCreationState == nil {
			m.showErrorDialog("Save PRD", "Internal error: PRD state not found")
			return nil
		}
		
		state := m.prdCreationState
		
		if state.GeneratedContent == "" {
			m.showErrorDialog("Save PRD", "No content to save")
			return nil
		}
		
		// Determine output path
		if state.Filename == "" {
			m.showErrorDialog("Save PRD", "Output filename not specified")
			return nil
		}
		
		// Resolve docs directory using unified helper
		docsDir, err := pathutil.GetPrdDirectory(m.config, "")
		if err != nil {
			m.showErrorDialog("Save PRD", fmt.Sprintf("Failed to create PRD directory: %v", err))
			return nil
		}
		
		filePath := filepath.Join(docsDir, state.Filename)
		
		// Check if file exists
		if _, err := os.Stat(filePath); err == nil {
			// File exists, show confirmation dialog
			dm := m.dialogManager()
			if dm != nil {
				confirmDialog := dialog.NewConfirmationDialog(
					"File Already Exists",
					fmt.Sprintf("Overwrite %s?", state.Filename),
					60, 8,
				)
				confirmDialog.SetYesText("Overwrite")
				confirmDialog.SetNoText("Cancel")
				
				m.appState.AddDialog(confirmDialog, func(value interface{}, err error) tea.Cmd {
					if err != nil {
						return nil
					}
					
					msg, ok := value.(dialog.ConfirmationMsg)
					if !ok {
						return nil
					}
					
					if msg.Result == dialog.ConfirmationResultYes {
						return m.writeFileAndShowSuccess(filePath)
					} else {
						// Return to PRD input dialog
						return m.showPrdInputWithState()
					}
				})
				
				return confirmDialog.Init()
			}
			
			// Without dialog manager, ask via error dialog
			m.showErrorDialog("File Exists", "Cannot prompt for confirmation. Please use a different filename.")
			return nil
		}
		
		// File doesn't exist, write directly
		return m.writeFileAndShowSuccess(filePath)()
	}
}

// writeFileAndShowSuccess writes the PRD to file and shows success dialog
func (m *Model) writeFileAndShowSuccess(filePath string) tea.Cmd {
	return func() tea.Msg {
		state := m.prdCreationState
		
		// Strip ANSI codes from output
		cleanOutput := stripAnsiCodes(state.GeneratedContent)
		
		// Write to file
		if err := os.WriteFile(filePath, []byte(cleanOutput), 0644); err != nil {
			// Provide specific error messages based on error type
			errorMsg := fmt.Sprintf("Failed to write PRD file: %v", err)
			if os.IsPermission(err) {
				errorMsg = fmt.Sprintf("Permission denied writing to file.\nPath: %s\nPlease check directory permissions.", filePath)
			} else if os.IsNotExist(err) {
				errorMsg = fmt.Sprintf("Directory does not exist.\nPath: %s\nPlease try again.", filePath)
			}
			m.showErrorDialog("Save PRD", errorMsg)
			return nil
		}
		
		// Log success
		m.addLogLine(fmt.Sprintf("✓ PRD saved successfully to %s", filePath))
		m.lastPrdPath = filePath
		
		// Show success dialog
		dm := m.dialogManager()
		if dm != nil {
			successDialog := dialog.NewConfirmationDialog(
				"PRD Created Successfully",
				fmt.Sprintf("PRD saved to: %s", filePath),
				70, 8,
			)
			successDialog.SetYesText("Done")
			successDialog.SetNoText("Cancel")
			successDialog.SetYesDefault(true)
			
			m.appState.AddDialog(successDialog, func(value interface{}, err error) tea.Cmd {
				// Clear PRD creation state
				if m.prdCreationState != nil {
					m.prdCreationState.Clear()
					m.prdCreationState = nil
				}
				return nil
			})
			
			return successDialog.Init()
		}
		
		// Without dialog manager, just clear state
		if m.prdCreationState != nil {
			m.prdCreationState.Clear()
			m.prdCreationState = nil
		}
		
		return nil
	}
}
