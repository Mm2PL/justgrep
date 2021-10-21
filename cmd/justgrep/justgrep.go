package main

import (
	"flag"
	"fmt"
	"justgrep"
	"math"
	"os"
	"regexp"
	"strings"
	"time"
)

type arguments struct {
	url *string

	user        *string
	notUser     *string
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
	output, err = time.Parse("2006-01-02 15:04:05-07:00", input)
	if err == nil {
		return
	}
	output, err = time.Parse(time.RFC3339, input)
	if err == nil {
		return
	}

	return time.Time{}, err
}

func (args *arguments) validateAndProcessFlags() (valid bool) {
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
	args.notUser = flag.String("notuser", "", "Negative match on username")
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
	flagsAreValid := args.validateAndProcessFlags()
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
	var negativeRegex *regexp.Regexp
	matchMode := justgrep.DontMatch
	if *args.user != "" || *args.notUser != "" {
		matchMode = justgrep.MatchExact
	}

	if *args.userIsRegex {
		matchMode = justgrep.MatchRegex
		userRegex, err = regexp.Compile(*args.user)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error while compiling your username regex: %s\n", err)
			return
		}
		negativeRegex, err = regexp.Compile(*args.notUser)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error while compiling your negative username regex: %s\n", err)
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

		UserName:         *args.user,
		NegativeUserName: *args.notUser,

		NegativeUserRegex: negativeRegex,
		UserRegex:         userRegex,

		Count: *args.maxResults,
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
	for currentIndex, channel := range channelsToSearch {
		if *args.verbose {
			_, _ = fmt.Fprintf(os.Stderr, "Now scanning #%s %d/%d\n", channel, currentIndex+1, len(channelsToSearch))
		}
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

const progressSize = 50

func makeProgressBar(totalSteps float64, stepsLeft float64) string {
	var fracDone float64
	if totalSteps == 0 {
		fracDone = 0
		stepsLeft = 1
		totalSteps = 2
	} else {
		fracDone = 1 - stepsLeft/totalSteps
	}
	done := strings.Repeat("=", int(math.Floor(progressSize*fracDone)))
	left := strings.Repeat(" ", int(math.Ceil(progressSize*(1-fracDone))))
	return fmt.Sprintf("[%s>%s] %.2f%%", done, left, fracDone*100)
}
func searchLogs(args *arguments, err error, api justgrep.JustlogAPI, download chan *justgrep.Message, filter justgrep.Filter, totalResults []int) {
	nextDate := args.endTime
	cancelled := false
	var channel string
	step := api.GetApproximateOffset()
	switch api.(type) {
	default:
		channel = fmt.Sprintf("[unknown] (%t)", api)
		step = time.Hour * 24
	case *justgrep.UserJustlogAPI:
		channel = api.(*justgrep.UserJustlogAPI).Channel
	case *justgrep.ChannelJustlogAPI:
		channel = api.(*justgrep.ChannelJustlogAPI).Channel
	}
	totalSteps := float64(args.endTime.Sub(args.startTime) / step)

	for {
		stepsLeft := float64(nextDate.Sub(args.startTime) / step)
		if *args.verbose {
			_, _ = fmt.Fprintf(os.Stderr, "Found %d matching messages... Downloading #%s at %s %s.\n", totalResults[justgrep.ResultOk], channel, nextDate.Format("2006-01-02"), makeProgressBar(totalSteps, stepsLeft))
		}
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
		if nextDate.Before(args.startTime) || results[justgrep.ResultMaxCountReached] != 0 {
			break
		}
	}
}
