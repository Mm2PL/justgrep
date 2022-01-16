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
	MessageTypes   []string

	HasMessageRegex bool
	MessageRegex    *regexp.Regexp

	UserMatchType UserMatchType

	UserRegex        *regexp.Regexp
	NegativeUserRegex *regexp.Regexp
	UserName         string
	NegativeUserName string

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
