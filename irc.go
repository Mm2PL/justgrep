package justgrep

import (
	"fmt"
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

func (m Message) String() string {
	return fmt.Sprintf("Message{Prefix: %q, Action: %q, Args: %q, Timestamp: %s}", m.Prefix, m.Action, m.Args, m.Timestamp)
}

func NewMessage(text string) *Message {
	output := &Message{Raw: text}
	cpy := text
	if cpy[0] == '@' {
		cpy = cpy[1:]
		// has tags
		idx := strings.Index(cpy, " ")
		tagsRaw := cpy[:idx]
		output.Tags = make(map[string]string, 16)
		for _, pair := range strings.Split(tagsRaw, ";") {
			splitPair := strings.Split(pair, "=")
			output.Tags[splitPair[0]] = unescapeValue(splitPair[1])
		}
		cpy = cpy[idx+1:]
	}
	if cpy[0] == ':' {
		prefixIdx := strings.Index(cpy, " ")
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
		if err == nil {
			output.Timestamp = time.Unix(parsedInt/1000, parsedInt%1000*1000000)
		}
	}
	return output
}

func unescapeValue(s string) string {
	nextEscaped := false
	output := ""
	for _, chr := range s {
		if nextEscaped {
			switch chr {
			case ':':
				output += ";"
			case 'r':
				output += "\r"
			case 'n':
				output += "\n"
			default:
				output += "\\" + string(chr)
			}
			nextEscaped = false
		} else if chr == '\\' {
			nextEscaped = true
		} else {
			output += string(chr)
		}
	}
	return output
}
