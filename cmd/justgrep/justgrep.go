package main

import (
	"flag"
	"fmt"
	"justgrep"
	"os"
	"regexp"
	"time"
)

type arguments struct {
	url *string

	user        *string
	userIsRegex *bool

	channel      *string
	messageRegex *string

	msgOnly *bool

	start *string
	end   *string

	startTime time.Time
	endTime   time.Time

	verbose *bool
}

func (args *arguments) validateFlags() (valid bool) {
	valid = true
	if *args.channel == "" {
		fmt.Println("You need to pass the -channel argument.")
		valid = false
	}
	if *args.start == "" {
		fmt.Println("You need to pass the -start argument.")
		valid = false
	}
	if *args.end == "" {
		fmt.Println("You need to pass the -end argument.")
		valid = false
	}
	// show missing arguments and that's it
	if !valid {
		return
	}

	startTime, err := time.Parse("2006-01-02 15:04:05", *args.start)
	if err != nil {
		fmt.Printf("-start: Invalid time: %s: %s\n", *args.start, err)
		valid = false
	}
	args.startTime = startTime

	endTime, err := time.Parse("2006-01-02 15:04:05", *args.end)
	if err != nil {
		fmt.Printf("-end: Invalid time: %s: %s\n", *args.end, err)
		valid = false
	}
	args.endTime = endTime
	return
}

func main() {
	args := &arguments{}
	args.user = flag.String("user", "", "Target user")
	args.userIsRegex = flag.Bool("uregex", false, "Is the -user option a regex?")

	args.msgOnly = flag.Bool("msg-only", false, "Only want chat messages (PRIVMSGs).")

	args.channel = flag.String("channel", "", "Target channel")
	args.messageRegex = flag.String("regex", "", "Message Regex")
	args.start = flag.String("start", "", "Start time")
	args.end = flag.String("end", "", "End time")
	args.url = flag.String("url", "http://localhost:8025", "Justlog instance URL")

	args.verbose = flag.Bool("v", false, "Spam stdout a little more")
	flag.Parse()
	flagsAreValid := args.validateFlags()
	if !flagsAreValid {
		os.Exit(1)
	}

	messageExpr, err := regexp.Compile(*args.messageRegex)
	if err != nil {
		fmt.Printf("Error while compiling your message regex: %s\n", err)
	}

	var api justgrep.JustlogAPI
	if *args.user != "" && !(*args.userIsRegex) {
		api = &justgrep.UserJustlogAPI{User: *args.user, Channel: *args.channel, URL: *args.url}
	} else {
		api = &justgrep.ChannelJustlogAPI{Channel: *args.channel, URL: *args.url}
	}

	download := make(chan *justgrep.Message)

	var userRegex *regexp.Regexp
	matchMode := justgrep.DontMatch
	if *args.userIsRegex {
		matchMode = justgrep.MatchRegex
		userRegex, err = regexp.Compile(*args.user)
		if err != nil {
			fmt.Printf("Error while compiling your username regex: %s\n", err)
		}
	}
	filter := justgrep.Filter{
		StartDate: args.startTime,
		EndDate:   args.endTime,

		HasMessageType: *args.msgOnly,
		MessageType:    "PRIVMSG",

		HasMessageRegex: true,
		MessageRegex:    messageExpr,

		UserMatchType: matchMode,
		UserName:      *args.user,
		UserRegex:     userRegex,
	}
	totalResults := make([]int, justgrep.ResultCount)
	nextDate := args.endTime

	for {
		nextDate, err = justgrep.FetchForDate(api, nextDate, download)
		if err != nil {
			fmt.Printf("Error while fetching logs: %s\n", err)
			break
		}

		filtered := make(chan *justgrep.Message)
		var results []int
		go func() {
			results = filter.StreamFilter(download, filtered)
			filtered <- nil
		}()
		for {
			msg := <-filtered
			if msg == nil {
				break
			}
			fmt.Println(msg.Raw)
		}

		for result, count := range results {
			totalResults[result] += count
		}
		if nextDate.Before(args.startTime) {
			break
		}
	}
	if *args.verbose {
		fmt.Println("Summary:")
		for result, count := range totalResults {
			fmt.Printf("  %s  %d\n", justgrep.FilterResult(result), count)
		}
	}
}
