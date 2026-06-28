package runner

import (
	"context"
	"fmt"
)

// stepResult records what happened to a step, for the summary at the end.
type stepResult struct {
	name    string
	passed  bool
	skipped bool
	aborted bool // user hit abort — signals the job to stop and print summary
}

// Run is the main entry point called by the CLI.
func Run(workflowPath string) error {
	path, err := findWorkflow(workflowPath)
	if err != nil {
		return err
	}

	wf, err := parseWorkflow(path)
	if err != nil {
		return err
	}

	fmt.Printf("\n  cidb  ·  %s\n", wf.Name)
	fmt.Printf("  Workflow: %s\n", path)
	fmt.Printf("  Jobs: %d\n", len(wf.Jobs))

	ctx := context.Background()

	for jobID, job := range wf.Jobs {
		if err := runJob(ctx, jobID, job); err != nil {
			return err
		}
	}

	return nil
}

func runJob(ctx context.Context, jobID string, job Job) error {
	fmt.Printf("\n  ┌─ Job: %s (runs-on: %s)\n\n", jobID, job.RunsOn)

	ctr, err := startContainer(ctx, job.RunsOn)
	if err != nil {
		return err
	}
	defer ctr.stop()

	var results []stepResult

	for i, step := range job.Steps {
		name := stepName(step, i)

		if step.Uses != "" && step.Run == "" {
			fmt.Printf("\n  ─── Step %d: %s\n", i+1, name)
			fmt.Printf("  (uses: %s — action steps not supported in v1, skipping)\n", step.Uses)
			results = append(results, stepResult{name: name, skipped: true})
			printStepResult(name, false, true)
			continue
		}

		result := runStep(ctr, i+1, name, step)
		results = append(results, result)

		// Always print summary before stopping, whether aborted or failed
		if result.aborted {
			fmt.Println("\n  Aborted.")
			printSummary(results)
			return nil
		}

		if !result.passed && !result.skipped {
			printSummary(results)
			return fmt.Errorf("job %q stopped: step %q failed", jobID, name)
		}
	}

	printSummary(results)
	return nil
}

// runStep handles the pause → execute → result loop for a single step.
func runStep(ctr *Container, num int, name string, step Step) stepResult {
	for {
		action := pause(num, name, step.Run)

		switch action {
		case ActionAbort:
			return stepResult{name: name, aborted: true}

		case ActionSkip:
			printStepResult(name, false, true)
			return stepResult{name: name, skipped: true}

		case ActionShell:
			if err := ctr.dropShell(); err != nil {
				fmt.Printf("\n  Shell error: %v\n", err)
			}
			continue

		case ActionContinue:
			fmt.Println()
			exitCode, err := ctr.exec(step.Run, step.Env)
			fmt.Println()

			if err != nil {
				fmt.Printf("  Exec error: %v\n", err)
				printStepResult(name, false, false)
				return stepResult{name: name, passed: false}
			}

			if exitCode == 0 {
				printStepResult(name, true, false)
				return stepResult{name: name, passed: true}
			}

			// Step failed — pause and let the user decide what to do
			fmt.Printf("  Step exited with code %d\n", exitCode)
			printStepResult(name, false, false)

			// Loop: keep re-prompting until the user picks something decisive
			for {
				action = pause(num, name+" (failed — what next?)", "")
				switch action {
				case ActionShell:
					if err := ctr.dropShell(); err != nil {
						fmt.Printf("\n  Shell error: %v\n", err)
					}
					// Stay in the failure loop so we re-show "failed — what next?"
				case ActionAbort:
					return stepResult{name: name, aborted: true}
				case ActionContinue, ActionSkip:
					return stepResult{name: name, passed: false}
				}
			}
		}
	}
}

func stepName(step Step, index int) string {
	if step.Name != "" {
		return step.Name
	}
	if step.Uses != "" {
		return step.Uses
	}
	return fmt.Sprintf("step %d", index+1)
}
