package compare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
)

func to_terminal_safe_json_string(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func Compare(expected, actual any, url string) (string, error) {
	if info, err := exec.LookPath("compare"); err == nil && info != "" {
		// Use local compare binary

		terminal_command := fmt.Sprintf("compare -expected='%s' -actual='%s'", to_terminal_safe_json_string(expected), to_terminal_safe_json_string(actual))
		cmd := exec.Command("sh", "-c", terminal_command)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return string(output), err
		}
		return string(output), nil
	} else {
		println("compare is not a command")
	}
	if url == "" {
		url = "https://compare-production-1494.up.railway.app/compare1"
	}

	payload := map[string]any{
		"expected": expected,
		"actual":   actual,
	}

	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest("GET", url, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return string(body), nil
}
