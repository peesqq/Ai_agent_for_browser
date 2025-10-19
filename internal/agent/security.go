package agent

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func ConfirmDangerous(action string) bool {
	fmt.Printf("[SECURITY] Подтвердите действие: %s (y/n): ", action)
	reader := bufio.NewReader(os.Stdin)
	ans, _ := reader.ReadString('\n')
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(ans)), "y")
}
