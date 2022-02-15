package justgrep

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Message struct {
	Raw       string            `json:"raw,omitempty"`
	Prefix    string            `json:"prefix,omitempty"`
	User      string            `json:"user,omitempty"`
	Args      []string          `json:"args,omitempty"`
	Action    string            `json:"action,omitempty"`
	Tags      map[string]string `json:"tags,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

func (m Message) Serialize() (output string) {
	if m.Tags != nil && len(m.Tags) != 0 {
		output += "@"
		// this is all to sort tags alphabetically to produce constant output
		keys := make([]string, len(m.Tags))
		i := 0
		for k := range m.Tags {
			keys[i] = k
			i += 1
		}
		sort.Strings(keys)

		maxIdx := len(keys) - 1
		for i, k := range keys {
			v := m.Tags[k]
			if i == maxIdx {
				output += k + "=" + escapeValue(v)
			} else {
				output += k + "=" + escapeValue(v) + ";"
			}
		}
		output += " "
	}
	if m.Prefix != "" {
		output += ":" + m.Prefix + " "
	}
	output += m.Action
	if len(m.Args) != 0 {
		output += " "
		for i := 0; i < len(m.Args); i++ {
			arg := m.Args[i]
			if i == len(m.Args)-1 {
				output += ":" + arg + "\r\n"
				return
			} else {
				output += arg + " "
			}
		}
	}
	output += "\r\n"
	return // this should only be hit if there are no args
}
func (m Message) String() string {
	return fmt.Sprintf("Message{Prefix: %q, Action: %q, Args: %q, Timestamp: %s}", m.Prefix, m.Action, m.Args, m.Timestamp)
}

func NewMessage(text string) (*Message, error) {
	if len(text) == 0 {
		return nil, errors.New("parser error: empty input")
	}
	output := &Message{Raw: text}
	cpy := text
	if cpy[0] == '@' {
		cpy = cpy[1:]
		// has tags
		idx := strings.Index(cpy, " ")
		if idx == -1 {
			return nil, errors.New("parser error: unable to find a space after tags, looks like input was trimmed")
		}
		tagsRaw := cpy[:idx]
		output.Tags = make(map[string]string, 16)
		for _, pair := range strings.Split(tagsRaw, ";") {
			equalsIdx := strings.IndexRune(pair, '=')
			if equalsIdx == -1 || len(pair) <= equalsIdx {
				return nil, errors.New("parser error: invalid tag key value pair")
			}
			output.Tags[pair[:equalsIdx]] = unescapeValue(pair[equalsIdx+1:])
		}
		cpy = cpy[idx+1:]
	}
	if cpy[0] == ':' {
		prefixIdx := strings.Index(cpy, " ")
		if prefixIdx == -1 {
			return nil, errors.New("parser error: unable to find a space after the prefix, looks like input was trimmed")
		}
		prefix := cpy[1:prefixIdx]
		cpy = cpy[prefixIdx+1:]
		output.Prefix = prefix
		accountSep := strings.Index(prefix, "!")
		if accountSep != -1 {
			output.User = prefix[:accountSep]
		}
	}
	actionIndex := strings.Index(cpy, " ")
	if actionIndex == -1 {
		output.Action = cpy
	} else {
		output.Action = cpy[:actionIndex]
		cpy = cpy[actionIndex+1:]
		for {
			nextSpace := strings.Index(cpy, " ")
			if nextSpace == -1 {
				// has to be last arg!
				nextSpace = len(cpy) - 1
			}
			if cpy == "" {
				break
			}
			currentArg := cpy[:nextSpace]
			if currentArg[0] == ':' {
				// last argument.
				output.Args = append(output.Args, cpy[1:])
				break
			}
			output.Args = append(output.Args, currentArg)
			cpy = cpy[nextSpace+1:]
		}
	}
	ts, hasTs := output.Tags["tmi-sent-ts"]
	if hasTs {
		parsedInt, err := strconv.ParseInt(ts, 10, 64)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("parser error: unable to parse time (@tmi-sent-ts): %q: %s", ts, err))
		}
		output.Timestamp = time.Unix(parsedInt/1000, parsedInt%1000*1000000)
	} else {
		ts, hasTs = output.Tags["time"]
		if hasTs {
			stamp, err := time.Parse(time.RFC3339, ts)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("parser error: unable to parse time (@time): %q: %s", ts, err))
			}
			output.Timestamp = stamp
		}
	}
	return output, nil
}

func unescapeValue(s string) string {
	nextEscaped := false
	unescaper := func(r rune) rune {
		if nextEscaped {
			nextEscaped = false
			switch r {
			case ':':
				return ';'
			case 'r':
				return '\r'
			case 'n':
				return '\n'
			case 's':
				return ' '
			default:
				return r
			}
		} else if r == '\\' {
			nextEscaped = true
			return -1
		} else {
			return r
		}
	}
	return strings.Map(unescaper, s)
}

func escapeValue(s string) string {
	out := make([]rune, 1024)
	usedCount := 0
	for _, chr := range s {
		switch chr {
		case ';':
			out[usedCount] = '\\'
			out[usedCount+1] = ':'
			usedCount += 2
		case '\r':
			out[usedCount] = '\\'
			out[usedCount+1] = 'r'
			usedCount += 2
		case '\n':
			out[usedCount] = '\\'
			out[usedCount+1] = 'n'
			usedCount += 2
		case ' ':
			out[usedCount] = '\\'
			out[usedCount+1] = 's'
			usedCount += 2
		default:
			out[usedCount] = chr
			usedCount += 1
		}
	}
	return string(out[:usedCount])
}