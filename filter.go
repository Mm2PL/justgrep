package justgrep

import (
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
	MessageType    string

	HasMessageRegex bool
	MessageRegex    *regexp.Regexp

	UserMatchType UserMatchType
	UserRegex     *regexp.Regexp
	UserName      string
	Count         int
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

func (f Filter) StreamFilter(input chan *Message, output chan *Message, cancelled *bool) []int {
	results := make([]int, ResultCount)
	for {
		if f.Count != 0 && results[ResultOk] >= f.Count {
			results[ResultMaxCountReached] = 1
			*cancelled = true
			break
		}
		msg := <-input
		if msg == nil {
			break
		}
		result := f.Filter(msg)
		results[result]++
		if result == ResultOk {
			output <- msg
		}
	}
	return results
}

func (f Filter) Filter(msg *Message) FilterResult {
	if msg.Timestamp.After(f.EndDate) {
		return ResultDate
	}

	if msg.Timestamp.Before(f.StartDate) {
		return ResultDate
	}

	if f.HasMessageType && f.MessageType != msg.Action {
		return ResultType
	}
	if f.HasMessageRegex && !f.MessageRegex.MatchString(msg.Args[len(msg.Args)-1]) {
		return ResultContent
	}
	switch f.UserMatchType {
	case DontMatch:
		break
	case MatchRegex:
		if !f.UserRegex.MatchString(msg.User) {
			return ResultUser
		}
	case MatchExact:
		if f.UserName != msg.User {
			return ResultUser
		}
	}
	return ResultOk
}
