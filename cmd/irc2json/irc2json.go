package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"justgrep"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for scanner.Scan() {
		msg := justgrep.NewMessage(scanner.Text())
		err := encoder.Encode(msg)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Failed to JSON encode message: %s", err)
			os.Exit(1)
		}
	}
}
