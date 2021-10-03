package main

import (
	"flag"
	"fmt"
	"justgrep"
	"os"
	"regexp"
	"strings"
	"time"
)

type arguments struct {
	url *string

	user        *string
	userIsRegex *bool

	channel      *string
	messageRegex *string
	maxResults   *int

	msgOnly *bool

	start *string
	end   *string

	startTime time.Time
	endTime   time.Time

	verbose   *bool
	recursive *bool
}

func parseTime(input string) (output time.Time, err error) {
	output, err = time.Parse("2006-01-02 15:04:05", input)
	if err == nil {
		return
	}
	output, err = time.Parse(time.RFC3339, input)
	if err == nil {
		return
	}

	return time.Time{}, err
}

func (args *arguments) validateFlags() (valid bool) {
	valid = true
	if *args.channel == "" && !*args.recursive {
		_, _ = fmt.Fprintln(os.Stderr, "You need to pass the -channel or -r (recursive) arguments.")
		valid = false
	}
	if *args.start == "" {
		_, _ = fmt.Fprintln(os.Stderr, "You need to pass the -start argument.")
		valid = false
	}
	if *args.end == "" {
		_, _ = fmt.Fprintln(os.Stderr, "You need to pass the -end argument.")
		valid = false
	}
	// show missing arguments and that's it
	if !valid {
		return
	}

	startTime, err := parseTime(*args.start)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "-start: Invalid time: %s: %s\n", *args.start, err)
		valid = false
	}
	args.startTime = startTime

	endTime, err := parseTime(*args.end)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "-end: Invalid time: %s: %s\n", *args.end, err)
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
	args.maxResults = flag.Int("max", 0, "How many results do you want? 0 for unlimited")

	args.verbose = flag.Bool("v", false, "Spam stdout a little more")
	args.recursive = flag.Bool("r", false, "Run search on all channels.")
	flag.Parse()
	flagsAreValid := args.validateFlags()
	if !flagsAreValid {
		os.Exit(1)
	}

	messageExpr, err := regexp.Compile(*args.messageRegex)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error while compiling your message regex: %s\n", err)
		return
	}

	download := make(chan *justgrep.Message)

	var userRegex *regexp.Regexp
	matchMode := justgrep.DontMatch
	if *args.userIsRegex {
		matchMode = justgrep.MatchRegex
		userRegex, err = regexp.Compile(*args.user)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error while compiling your username regex: %s\n", err)
			return
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
		Count:         *args.maxResults,
	}
	totalResults := make([]int, justgrep.ResultCount)
	var channelsToSearch []string
	if !*args.recursive {
		channelsToSearch = strings.Split(*args.channel, ",")
	} else {
		channelsToSearch, err = justgrep.GetChannelsFromJustLog(*args.url)
		if err != nil {
			_, err := fmt.Fprintf(os.Stderr, "Error while fetching channels from justlog: %s", err)
			if err != nil {
				return
			}
			os.Exit(1)
		}
	}
	for _, channel := range channelsToSearch {
		var api justgrep.JustlogAPI
		if *args.user != "" && !(*args.userIsRegex) {
			api = &justgrep.UserJustlogAPI{User: *args.user, Channel: channel, URL: *args.url}
		} else {
			api = &justgrep.ChannelJustlogAPI{Channel: channel, URL: *args.url}
		}
		searchLogs(args, err, api, download, filter, totalResults)
	}
	if *args.verbose {
		_, _ = fmt.Fprintf(os.Stderr, "Summary:\n")
		for result, count := range totalResults {
			_, _ = fmt.Fprintf(os.Stderr, " - %s => %d\n", justgrep.FilterResult(result), count)
		}
	}
}

func searchLogs(args *arguments, err error, api justgrep.JustlogAPI, download chan *justgrep.Message, filter justgrep.Filter, totalResults []int) {
	nextDate := args.endTime
	cancelled := false
	for {
		nextDate, err = justgrep.FetchForDate(api, nextDate, download, &cancelled)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error while fetching logs: %s\n", err)
			break
		}

		filtered := make(chan *justgrep.Message)
		var results []int
		go func() {
			results = filter.StreamFilter(download, filtered, &cancelled)
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
}
