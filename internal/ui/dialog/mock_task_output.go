package dialog

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// MockScenario represents different mock execution scenarios
type MockScenario int

const (
	ScenarioQuickSuccess MockScenario = iota
	ScenarioLongRunning
	ScenarioWithError
	ScenarioWithWarnings
)

// MockTaskRunner generates realistic mock output for testing
type MockTaskRunner struct {
	outputCh chan TaskOutputMsg
	doneCh   chan struct{}
	taskID   string
	tabIndex int
	scenario MockScenario
	running  bool
}

// NewMockTaskRunner creates a new mock task runner
func NewMockTaskRunner(taskID string, tabIndex int, scenario MockScenario) *MockTaskRunner {
	return &MockTaskRunner{
		outputCh: make(chan TaskOutputMsg, 100),
		doneCh:   make(chan struct{}),
		taskID:   taskID,
		tabIndex: tabIndex,
		scenario: scenario,
		running:  false,
	}
}

// Start begins generating mock output on a goroutine
func (m *MockTaskRunner) Start() {
	if m.running {
		return
	}
	m.running = true
	go m.generateOutput()
}

// Cancel stops the output generation
func (m *MockTaskRunner) Cancel() {
	if m.running {
		m.running = false
		m.Close()
	}
}

// Close cleans up resources
func (m *MockTaskRunner) Close() {
	close(m.outputCh)
	close(m.doneCh)
}

// OutputChan returns the output channel
func (m *MockTaskRunner) OutputChan() <-chan TaskOutputMsg {
	return m.outputCh
}

// DoneChan returns the done channel
func (m *MockTaskRunner) DoneChan() <-chan struct{} {
	return m.doneCh
}

// generateOutput produces mock output based on the scenario
func (m *MockTaskRunner) generateOutput() {
	defer func() {
		close(m.outputCh)
	}()

	switch m.scenario {
	case ScenarioQuickSuccess:
		m.genQuickSuccess()
	case ScenarioLongRunning:
		m.genLongRunning()
	case ScenarioWithError:
		m.genWithError()
	case ScenarioWithWarnings:
		m.genWithWarnings()
	}
}

// genQuickSuccess generates a quick success scenario
func (m *MockTaskRunner) genQuickSuccess() {
	outputs := []string{
		"🔍 Analyzing task requirements...",
		"📦 Gathering dependencies...",
		"⚙️  Configuring environment...",
		"✅ Environment ready",
		"🚀 Starting task execution...",
		"💭 Thinking about approach...",
		"🛠️  Building solution...",
		"✓ Task completed successfully in 2.3s",
	}

	for _, output := range outputs {
		if !m.running {
			return
		}
		m.outputCh <- TaskOutputMsg{TaskID: m.taskID, Output: output}
		time.Sleep(time.Duration(200+rand.Intn(400)) * time.Millisecond)
	}
}

// genLongRunning generates a long-running scenario
func (m *MockTaskRunner) genLongRunning() {
	steps := []string{
		"🔍 Analyzing project structure...",
		"📝 Reading configuration files...",
		"🔗 Resolving dependencies...",
		"⚙️  Configuring build system...",
		"✅ Configuration complete",
		"🚀 Starting build process...",
	}

	for _, step := range steps {
		if !m.running {
			return
		}
		m.outputCh <- TaskOutputMsg{TaskID: m.taskID, Output: step}
		time.Sleep(time.Duration(300+rand.Intn(700)) * time.Millisecond)
	}

	// Simulate long-running compilation/processing
	for i := 0; i < 15; i++ {
		if !m.running {
			return
		}
		progress := fmt.Sprintf("⏳ Processing step %d/15... (%.1f%%)", i+1, float64(i+1)/15*100)
		m.outputCh <- TaskOutputMsg{TaskID: m.taskID, Output: progress}
		time.Sleep(time.Duration(1000+rand.Intn(2000)) * time.Millisecond)
	}

	m.outputCh <- TaskOutputMsg{TaskID: m.taskID, Output: "🔧 Finalizing results..."}
	time.Sleep(500 * time.Millisecond)
	m.outputCh <- TaskOutputMsg{TaskID: m.taskID, Output: "✓ Task completed successfully in 35.8s"}
}

// genWithError generates a scenario with an error
func (m *MockTaskRunner) genWithError() {
	outputs := []string{
		"🔍 Analyzing task requirements...",
		"📦 Gathering dependencies...",
		"⚙️  Configuring environment...",
		"✅ Environment ready",
		"🚀 Starting task execution...",
		"💭 Thinking about approach...",
		"🛠️  Building solution...",
		"⚠️  Building component A...",
		"⚠️  Building component B...",
		"❌ Error: Failed to compile component C",
		"   └─ Missing import: github.com/example/lib",
		"📋 Stack trace:",
		"   /project/src/component.go:42:15",
		"❌ Task failed",
	}

	for _, output := range outputs {
		if !m.running {
			return
		}
		m.outputCh <- TaskOutputMsg{TaskID: m.taskID, Output: output}
		time.Sleep(time.Duration(200+rand.Intn(500)) * time.Millisecond)
	}
}

// genWithWarnings generates a scenario with warnings
func (m *MockTaskRunner) genWithWarnings() {
	outputs := []string{
		"🔍 Analyzing task requirements...",
		"📦 Gathering dependencies...",
		"⚙️  Configuring environment...",
		"✅ Environment ready",
		"🚀 Starting task execution...",
		"💭 Thinking about approach...",
		"🛠️  Building solution...",
		"⚠️  Warning: Deprecated function used in module A",
		"⚠️  Warning: Missing test coverage in module B",
		"⚠️  Warning: Performance concern in module C",
		"   └─ Consider optimizing loop in main()",
		"✅ Build completed with 3 warnings",
		"📊 Test results: 145 passed, 2 skipped",
		"✓ Task completed successfully in 8.4s",
	}

	for _, output := range outputs {
		if !m.running {
			return
		}
		m.outputCh <- TaskOutputMsg{TaskID: m.taskID, Output: output}
		time.Sleep(time.Duration(200+rand.Intn(400)) * time.Millisecond)
	}
}

// GenerateRandomOutput creates a random output pattern
func (m *MockTaskRunner) GenerateRandomOutput() string {
	patterns := []string{
		"📝 Processing item %d...",
		"🔄 Transforming data...",
		"📊 Analyzing metrics...",
		"🔗 Linking dependencies...",
		"⚙️  Configuring module...",
		"✅ Validation passed",
		"⏳ Waiting for resource...",
		"💾 Persisting state...",
	}

	randomIdx := rand.Intn(len(patterns))
	pattern := patterns[randomIdx]

	if strings.Contains(pattern, "%d") {
		return fmt.Sprintf(pattern, rand.Intn(100))
	}
	return pattern
}

// SimulateTaskOutput sends simulated output over time
func (m *MockTaskRunner) SimulateTaskOutput(outputs []string, delayMs int) {
	for _, output := range outputs {
		if !m.running {
			return
		}
		m.outputCh <- TaskOutputMsg{TaskID: m.taskID, Output: output}
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
	}
}

// GetScenarioName returns the name of the current scenario
func (m *MockTaskRunner) GetScenarioName() string {
	switch m.scenario {
	case ScenarioQuickSuccess:
		return "Quick Success"
	case ScenarioLongRunning:
		return "Long Running"
	case ScenarioWithError:
		return "With Error"
	case ScenarioWithWarnings:
		return "With Warnings"
	default:
		return "Unknown"
	}
}
