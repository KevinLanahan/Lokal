package runner

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const envFile = ".env"

// readEnvFile reads the .env file and returns a map of key→value and the original line order.
func readEnvFile() (map[string]string, []string, error) {
	secrets := make(map[string]string)
	var order []string

	f, err := os.Open(envFile)
	if os.IsNotExist(err) {
		return secrets, order, nil
	}
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			order = append(order, line)
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			secrets[key] = val
			order = append(order, key)
		}
	}
	return secrets, order, scanner.Err()
}

// writeEnvFile writes secrets back to .env preserving key order.
func writeEnvFile(secrets map[string]string, order []string) error {
	f, err := os.Create(envFile)
	if err != nil {
		return err
	}
	defer f.Close()

	seen := make(map[string]bool)
	for _, key := range order {
		if strings.HasPrefix(key, "#") || key == "" {
			fmt.Fprintln(f, key)
			continue
		}
		if val, ok := secrets[key]; ok {
			fmt.Fprintf(f, "%s=%s\n", key, val)
			seen[key] = true
		}
	}
	// Write any new keys not in the original order.
	for key, val := range secrets {
		if !seen[key] {
			fmt.Fprintf(f, "%s=%s\n", key, val)
		}
	}
	return nil
}

func SecretsSet(pair string) error {
	parts := strings.SplitN(pair, "=", 2)
	if len(parts) != 2 || parts[0] == "" {
		return fmt.Errorf("format must be KEY=VALUE")
	}
	key := strings.TrimSpace(parts[0])
	val := strings.TrimSpace(parts[1])

	secrets, order, err := readEnvFile()
	if err != nil {
		return err
	}
	isNew := secrets[key] == ""
	secrets[key] = val
	if isNew {
		order = append(order, key)
	}
	if err := writeEnvFile(secrets, order); err != nil {
		return err
	}
	fmt.Printf("  ✓  Set %s\n", key)
	return nil
}

func SecretsList() error {
	secrets, order, err := readEnvFile()
	if err != nil {
		return err
	}
	if len(secrets) == 0 {
		fmt.Println("  No secrets set. Use: lokal secrets set KEY=VALUE")
		return nil
	}
	fmt.Println()
	for _, key := range order {
		if strings.HasPrefix(key, "#") || key == "" {
			continue
		}
		if _, ok := secrets[key]; ok {
			masked := maskValue(secrets[key])
			fmt.Printf("  %-30s %s\n", key, masked)
		}
	}
	fmt.Println()
	return nil
}

func SecretsRemove(key string) error {
	secrets, order, err := readEnvFile()
	if err != nil {
		return err
	}
	if _, ok := secrets[key]; !ok {
		return fmt.Errorf("key %q not found in .env", key)
	}
	delete(secrets, key)
	if err := writeEnvFile(secrets, order); err != nil {
		return err
	}
	fmt.Printf("  ✓  Removed %s\n", key)
	return nil
}

// maskValue shows the first 4 chars and masks the rest.
func maskValue(val string) string {
	if len(val) <= 4 {
		return strings.Repeat("*", len(val))
	}
	stars := len(val) - 4
	if stars > 20 {
		stars = 20
	}
	return val[:4] + strings.Repeat("*", stars)
}

