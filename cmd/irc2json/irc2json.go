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
		msg, err := justgrep.NewMessage(scanner.Text())
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Failed to irc parse message: %s\n", err)
			os.Exit(1)
		}
		err = encoder.Encode(msg)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Failed to JSON encode message: %s\n", err)
			os.Exit(1)
		}
	}
}
