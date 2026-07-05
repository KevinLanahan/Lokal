package runner

import (
	"fmt"
	"os"
	"strings"
)

// evalContext holds secrets and step outputs for expression expansion.
type evalContext struct {
	secrets map[string]string
	outputs map[string]map[string]string // stepID -> name -> value
	env     map[string]string
}

func newEvalContext(secrets map[string]string) *evalContext {
	return &evalContext{
		secrets: secrets,
		outputs: make(map[string]map[string]string),
		env:     make(map[string]string),
	}
}

// loadSecrets reads the .env file and shell environment into a map.
func loadSecrets() map[string]string {
	secrets := make(map[string]string)
	// Shell environment first.
	for _, kv := range os.Environ() {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) == 2 {
			secrets[parts[0]] = parts[1]
		}
	}
	return secrets
}

// expand replaces ${{ expr }} placeholders in a string.
func (ec *evalContext) expand(s string) string {
	for {
		start := strings.Index(s, "${{")
		if start < 0 {
			break
		}
		end := strings.Index(s[start:], "}}")
		if end < 0 {
			break
		}
		end += start + 2
		expr := strings.TrimSpace(s[start+3 : end-2])
		val := ec.resolveExpr(expr)
		s = s[:start] + val + s[end:]
	}
	return s
}

func (ec *evalContext) resolveExpr(expr string) string {
	lower := strings.ToLower(expr)

	// secrets.FOO
	if strings.HasPrefix(lower, "secrets.") {
		key := expr[len("secrets."):]
		if v, ok := ec.secrets[key]; ok {
			return v
		}
		if v, ok := ec.secrets[strings.ToUpper(key)]; ok {
			return v
		}
		return ""
	}

	// env.FOO
	if strings.HasPrefix(lower, "env.") {
		key := expr[len("env."):]
		if v, ok := ec.env[key]; ok {
			return v
		}
		if v, ok := os.LookupEnv(key); ok {
			return v
		}
		return ""
	}

	// steps.STEP_ID.outputs.NAME
	if strings.HasPrefix(lower, "steps.") {
		parts := strings.Split(expr, ".")
		if len(parts) == 4 && strings.ToLower(parts[2]) == "outputs" {
			stepID := parts[1]
			outName := parts[3]
			if outs, ok := ec.outputs[stepID]; ok {
				if v, ok := outs[outName]; ok {
					return v
				}
			}
		}
		return ""
	}

	// github.* context — return sensible defaults
	if strings.HasPrefix(lower, "github.") {
		key := strings.ToLower(expr[len("github."):])
		switch key {
		case "sha":
			return "0000000"
		case "ref":
			return "refs/heads/main"
		case "event_name":
			return "push"
		case "actor":
			return "lokal"
		case "repository":
			return "owner/repo"
		case "workspace":
			return "/workspace"
		}
		return ""
	}

	// format() function
	if strings.HasPrefix(lower, "format(") {
		return ec.evalFormat(expr)
	}

	// Bare env var reference
	if v, ok := os.LookupEnv(expr); ok {
		return v
	}

	return fmt.Sprintf("${{%s}}", expr)
}

func (ec *evalContext) evalFormat(expr string) string {
	inner := strings.TrimPrefix(expr, "format(")
	inner = strings.TrimSuffix(inner, ")")
	parts := strings.SplitN(inner, ",", 2)
	if len(parts) < 2 {
		return inner
	}
	tmpl := strings.Trim(strings.TrimSpace(parts[0]), "'\"")
	args := strings.Split(parts[1], ",")
	for i, arg := range args {
		arg = strings.TrimSpace(arg)
		val := ec.resolveExpr(strings.Trim(arg, "'\""))
		tmpl = strings.ReplaceAll(tmpl, fmt.Sprintf("{%d}", i), val)
	}
	return tmpl
}

// expandStep returns a copy of the step with all ${{ }} expressions expanded.
func expandStep(step Step, ec *evalContext) Step {
	step.Run = ec.expand(step.Run)
	step.If = ec.expand(step.If)
	step.WorkingDirectory = ec.expand(step.WorkingDirectory)

	expanded := make(map[string]string, len(step.Env))
	for k, v := range step.Env {
		expanded[ec.expand(k)] = ec.expand(v)
	}
	step.Env = expanded

	expandedWith := make(map[string]string, len(step.With))
	for k, v := range step.With {
		expandedWith[ec.expand(k)] = ec.expand(v)
	}
	step.With = expandedWith

	return step
}

// parseStepOutputs scans output for GitHub Actions ::set-output and environment file syntax.
func parseStepOutputs(output, stepID string, ec *evalContext) {
	if stepID == "" || ec == nil {
		return
	}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		// ::set-output name=foo::bar
		if strings.HasPrefix(line, "::set-output name=") {
			rest := strings.TrimPrefix(line, "::set-output name=")
			parts := strings.SplitN(rest, "::", 2)
			if len(parts) == 2 {
				name := parts[0]
				val := parts[1]
				if ec.outputs[stepID] == nil {
					ec.outputs[stepID] = make(map[string]string)
				}
				ec.outputs[stepID][name] = val
			}
		}
	}
}
