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
		fmt.Printf("[%s] %s ", msg.Timestamp.UTC().Format("2006-01-02 15:04:05"), msg.Args[0])
		if msg.Action == "PRIVMSG" {
			fmt.Printf("%s: %s\n", msg.User, msg.Args[1])
		} else if msg.Action == "NOTICE" {
			fmt.Printf("NOTICE %s\n", msg.Args[1])
		} else if msg.Action == "CLEARCHAT" {
			if len(msg.Args) < 2 {
				fmt.Println("Chat has been cleared")
			} else {
				duration := msg.Tags["ban-duration"]
				if duration == "" {
					fmt.Printf("%s was permanently banned\n", msg.Args[1])
				} else {
					fmt.Printf("%s was timed out for %s seconds\n", msg.Args[1], duration)
				}
			}
		}
	}
}
