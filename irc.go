package justgrep

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Message struct {
	Raw       string
	Prefix    string
	Args      []string
	Action    string
	Tags      map[string]string
	Timestamp time.Time
}

func (m Message) String() string {
	return fmt.Sprintf("Message{Prefix: %q, Action: %q, Args: %q, Timestamp: %s}", m.Prefix, m.Action, m.Args, m.Timestamp)
}

func NewMessage(text string) *Message {
	output := &Message{}
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
		var prefix string // := doesn't work here?
		prefix = cpy[1:prefixIdx]
		cpy = cpy[prefixIdx+1:]
		output.Prefix = prefix
	}
	actionIndex := strings.Index(cpy, " ")
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
