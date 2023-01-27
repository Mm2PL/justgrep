package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Mm2PL/justgrep"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	i := 0
	for scanner.Scan() {
		i += 1
		msg, err := justgrep.NewMessage(scanner.Text())
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Line %d: Failed to irc parse message: %s\n", i, err)
			os.Exit(1)
		}
		err = encoder.Encode(msg)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Failed to JSON encode message: %s\n", err)
			os.Exit(1)
		}
	}
}
