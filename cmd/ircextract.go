package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"justgrep"
	"os"
	"strings"
	"time"
)

type extractArguments struct {
	timestamp      *bool
	action         *bool
	user           *bool
	messageText    *bool
	messageChannel *bool
	tagsRaw        *string
	tags           []string

	ignoreEmpty *bool
}

func main() {
	args := &extractArguments{}

	args.timestamp = flag.Bool("ts", false, "Extract the message timestamp")
	args.action = flag.Bool("action", false, "Extract the action field")
	args.user = flag.Bool("user", false, "Extract the user")
	args.messageText = flag.Bool("text", false, "Extract the message text")
	args.messageChannel = flag.Bool("channel", false, "Extract the message channel")
	args.tagsRaw = flag.String("tags", "", "Extract tags from the message")

	args.ignoreEmpty = flag.Bool("empty", true, "Don't ignore empty rows")
	flag.Parse()
	args.tags = strings.Split(*args.tagsRaw, ",")

	scanner := bufio.NewScanner(os.Stdin)

	writer := csv.NewWriter(os.Stdout)
	for scanner.Scan() {
		msg := justgrep.NewMessage(scanner.Text())
		line := make([]string, 0)
		meaningful := false
		if *args.timestamp {
			line = append(line, msg.Timestamp.Format(time.RFC3339))
			meaningful = true
		}
		if *args.action {
			line = append(line, msg.Action)
			meaningful = true
		}
		if *args.user {
			line = append(line, msg.User)
			if msg.User != "" {
				meaningful = true
			}
		}
		if *args.messageText {
			if msg.Action == "PRIVMSG" {
				line = append(line, msg.Args[len(msg.Args)-1])
				meaningful = true
			} else {
				line = append(line, "")
			}
		}
		if *args.messageChannel {
			if msg.Action == "PRIVMSG" {
				line = append(line, msg.Args[0])
				meaningful = true
			} else {
				line = append(line, "")
			}
		}
		if len(*args.tagsRaw) != 0 {
			for _, tag := range args.tags {
				val := msg.Tags[tag]
				if val != "" {
					meaningful = true
				}
				line = append(line, val)
			}
		}
		if !meaningful {
			continue
		}
		err := writer.Write(line)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	writer.Flush()
	err := writer.Error()
	if err != nil {
		fmt.Println(err)
		return
	}
}
