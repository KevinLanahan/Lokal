// runner/prompt.go handles the terminal UI — the pause prompt between steps
// and the per-step pass/fail output.
package runner

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Action represents what the user chose to do at a pause prompt.
type Action int

const (
	ActionContinue Action = iota // run this step
	ActionSkip                   // skip this step entirely
	ActionShell                  // drop into a container shell before running
	ActionRetry                  // re-show this prompt (user typed something invalid)
	ActionAbort                  // stop the whole run
)

// pause shows a step header and waits for the user to choose an action.
// It loops until valid input is received.
func pause(stepNum int, stepName string, command string) Action {
	fmt.Println()
	fmt.Println("  ─────────────────────────────────────────────────")
	fmt.Printf("  Step %d: %s\n", stepNum, stepName)
	fmt.Println("  ─────────────────────────────────────────────────")

	if command != "" {
		fmt.Println("  Command:")
		for _, line := range strings.Split(strings.TrimSpace(command), "\n") {
			fmt.Printf("    $ %s\n", line)
		}
		fmt.Println()
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("  [c]ontinue  [s]kip  [sh]ell  [a]bort  > ")
		if !scanner.Scan() {
			return ActionAbort
		}
		switch strings.TrimSpace(strings.ToLower(scanner.Text())) {
		case "c", "continue", "":
			return ActionContinue
		case "s", "skip":
			return ActionSkip
		case "sh", "shell":
			return ActionShell
		case "a", "abort", "q", "quit", "exit":
			return ActionAbort
		default:
			fmt.Println("  Unknown command. Options: c, s, sh, a")
		}
	}
}

// printStepResult prints a single ✓/✗/⏭ line after a step completes.
func printStepResult(name string, passed, skipped bool) {
	switch {
	case skipped:
		fmt.Printf("  ⏭  SKIP  %s\n", name)
	case passed:
		fmt.Printf("  ✓  PASS  %s\n", name)
	default:
		fmt.Printf("  ✗  FAIL  %s\n", name)
	}
}

// printSummary shows the full pass/fail table at the end of a job.
func printSummary(results []stepResult) {
	var passed, failed, skipped int

	fmt.Println()
	fmt.Println("  ─── Summary ────────────────────────────────────")
	for _, r := range results {
		printStepResult(r.name, r.passed, r.skipped)
		switch {
		case r.skipped:
			skipped++
		case r.passed:
			passed++
		default:
			failed++
		}
	}
	fmt.Println("  ─────────────────────────────────────────────────")
	fmt.Printf("  %d passed  %d failed  %d skipped\n\n", passed, failed, skipped)
}
