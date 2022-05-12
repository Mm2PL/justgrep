package justgrep

import (
	"context"
	"regexp"
	"strconv"
	"time"
)

type UserMatchType uint8

const (
	DontMatch UserMatchType = iota
	MatchRegex
	MatchExact
)

type Filter struct {
	// StartDate < EndDate
	StartDate time.Time
	EndDate   time.Time

	HasMessageType bool
	MessageTypes   []string

	HasMessageRegex bool
	MessageRegex    *regexp.Regexp

	UserMatchType UserMatchType

	UserRegex         *regexp.Regexp
	NegativeUserRegex *regexp.Regexp
	UserName          string
	NegativeUserName  string

	Count int
}
type FilterResult uint8

const (
	ResultOk FilterResult = iota
	ResultDate
	ResultType
	ResultContent
	ResultUser
	ResultMaxCountReached

	ResultCount
)

func (res FilterResult) String() string {
	switch res {
	case ResultOk:
		return "ok"
	case ResultDate:
		return "date"
	case ResultType:
		return "type"
	case ResultContent:
		return "content"
	case ResultUser:
		return "user"
	case ResultMaxCountReached:
		return "limit reached"
	default:
		return strconv.FormatInt(int64(res), 10)
	}
}

// StreamFilter performs Filter on every message from the input channel and puts every message that matched onto the
// output channel, if the max count of results is reached cancel() is called and results[ResultsMaxCountReached] is set.
// If the messages are too old, cancel() is called and results[ResultDate] is set.
func (f Filter) StreamFilter(
	cancel context.CancelFunc,
	input chan *Message,
	output chan *Message,
	progress *ProgressState,
) []int {
	results := make([]int, ResultCount)
	for msg := range input {
		if msg == nil {
			break
		}

		if f.Count != 0 && progress.TotalResults[ResultOk]+results[ResultOk] >= f.Count {
			results[ResultMaxCountReached] = 1
			cancel() // HTTP request is still going, kill it
			break
		}
		result := f.Filter(msg)
		results[result]++
		if result == ResultOk {
			output <- msg
		}
		if result == ResultDate {
			cancel() // HTTP request is still going, kill it
			break
		}
	}
	close(output)
	return results
}

// Filter performs all checks necessary to know if a given msg matches the Filter predicates.
func (f Filter) Filter(msg *Message) FilterResult {
	if msg.Timestamp.After(f.EndDate) {
		return ResultDate
	}

	if msg.Timestamp.Before(f.StartDate) {
		return ResultDate
	}
	if f.HasMessageType {
		ok := false
		for _, messageType := range f.MessageTypes {
			if messageType == msg.Action {
				ok = true
				break
			}
		}
		if !ok {
			return ResultType
		}
	}
	if f.HasMessageRegex && !f.MessageRegex.MatchString(msg.Args[len(msg.Args)-1]) {
		return ResultContent
	}
	switch f.UserMatchType {
	case DontMatch:
		break
	case MatchRegex:
		if f.UserName != "" && !f.UserRegex.MatchString(msg.User) {
			return ResultUser
		}

		if f.NegativeUserName != "" && f.NegativeUserRegex.MatchString(msg.User) {
			return ResultUser
		}
	case MatchExact:
		if f.UserName != "" && f.UserName != msg.User {
			return ResultUser
		}

		if f.NegativeUserName != "" && f.NegativeUserName == msg.User {
			return ResultUser
		}
	}
	return ResultOk
}
