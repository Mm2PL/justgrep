package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/Mm2PL/justgrep"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		msg, err := justgrep.NewMessage(scanner.Text())
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Failed to irc parse message: %s\n", err)
			os.Exit(1)
		}
		if msg.Action == "PRIVMSG" {
			fmt.Printf("[%s] %s %s: %s\n", msg.Timestamp.UTC().Format("2006-01-02 15:04:05"), msg.Args[0], msg.User, msg.Args[1])
		} else if msg.Action == "NOTICE" {
			fmt.Printf("NOTICE %s %s\n", msg.Args[0], msg.Args[1])
		} else if msg.Action == "CLEARCHAT" {
			duration := msg.Tags["ban-duration"]
			if duration == "" {
				fmt.Printf("%s was permanently banned\n", msg.Args[1])
			} else {
				fmt.Printf("%s was timed out for %s seconds\n", msg.Args[1], msg.Tags["ban-duration"])
			}
		}
	}
}
