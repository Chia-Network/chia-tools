package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ConfirmAction waits for the user to confirm with "yes" or "y"
func ConfirmAction(prompt string) bool {
	fmt.Printf("%s ", prompt)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "yes" || response == "y"
}
