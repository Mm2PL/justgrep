package justgrep

import (
	"regexp"
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
}
type FilterResult uint8

const (
	ResultOk FilterResult = iota
	ResultDate
	ResultType
	ResultContent
	ResultUser

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
	default:
		return string(res)
	}
}

func (f Filter) StreamFilter(input chan *Message, output chan *Message) []int {
	results := make([]int, ResultCount)
	for {
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
