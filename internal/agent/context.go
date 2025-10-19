package agent

import "strings"

func SummarizeForLLM(observation string) string {
	lines := strings.Split(observation, "\n")
	if len(lines) > 200 {
		lines = lines[:200]
	}
	return strings.Join(lines, "\n")
}
