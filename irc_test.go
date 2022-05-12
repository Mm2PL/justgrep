package justgrep

import (
	"bufio"
	"fmt"
	"os"
	"testing"
	"time"
)

func assert(t *testing.T, what string, have interface{}, expect interface{}) {
	if have != expect {
		t.Errorf("assertion on %s failed: have %q, expected %q", what, have, expect)
	}
}

func assertStrSlc(t *testing.T, what string, have []string, expect []string) {
	if len(have) != len(expect) {
		t.Errorf(
			"assertion on %s failed: length doesn't match, have %d, expected %d: %q vs %q",
			what,
			len(have),
			len(expect),
			have,
			expect,
		)
	}
	for i, elem := range have {
		if elem != expect[i] {
			t.Errorf("assertion on %s[%d] failed: have %q, expected %q", what, i, elem, expect[i])
		}
	}
}
func assertStrMap(t *testing.T, what string, have map[string]string, expect map[string]string) {
	if len(have) != len(expect) {
		t.Errorf(
			"assertion on %s failed: length doesn't match, have %d, expected %d: %q vs %q",
			what,
			len(have),
			len(expect),
			have,
			expect,
		)
	}
	for key, value := range have {
		if value != expect[key] {
			t.Errorf("assertion on %s[%q] failed: have %q, expected %q", what, key, value, expect[key])
		}
	}
}

func TestNewMessage(t *testing.T) {
	m, err := NewMessage("@badge-info=subscriber/15;badges=subscriber/12,glhf-pledge/1;color=#DAA520;display-name=Mm2PL;emotes=;flags=;id=1d7e0b34-fe74-4895-92ae-dd912046e637;mod=0;room-id=11148817;subscriber=1;tmi-sent-ts=1632058935165;turbo=0;user-id=117691339;user-type= :mm2pl!mm2pl@mm2pl.tmi.twitch.tv PRIVMSG #pajlada :-tags")
	assert(t, "error", err, nil)
	assert(t, "Action", m.Action, "PRIVMSG")
	assert(t, "Prefix", m.Prefix, "mm2pl!mm2pl@mm2pl.tmi.twitch.tv")
	assertStrSlc(t, "Args", m.Args, []string{"#pajlada", "-tags"})
	assertStrMap(
		t, "Tags", m.Tags, map[string]string{
			"badge-info":   "subscriber/15",
			"badges":       "subscriber/12,glhf-pledge/1",
			"color":        "#DAA520",
			"display-name": "Mm2PL",
			"emotes":       "",
			"flags":        "",
			"id":           "1d7e0b34-fe74-4895-92ae-dd912046e637",
			"mod":          "0",
			"room-id":      "11148817",
			"subscriber":   "1",
			"tmi-sent-ts":  "1632058935165",
			"turbo":        "0",
			"user-id":      "117691339",
			"user-type":    "",
		},
	)

	m, err = NewMessage("@badge-info=subscriber/15;badges=subscriber/12,glhf-pledge/1;color=#DAA520;display-name=Mm2PL;emotes=;flags=;id=1d7e0b34-fe74-4895-92ae-dd912046e637;mod=0;room-id=11148817;subscriber=1;tmi-sent-ts=1632058935165;turbo=0;user-id=117691339;user-type= :mm2pl!mm2pl@mm2pl.tmi.twitch.tv PRIVMSG #pajlada :-tags many words asdasd")
	assert(t, "error", err, nil)
	assert(t, "Action", m.Action, "PRIVMSG")
	assert(t, "Prefix", m.Prefix, "mm2pl!mm2pl@mm2pl.tmi.twitch.tv")
	assertStrSlc(t, "Args", m.Args, []string{"#pajlada", "-tags many words asdasd"})
	assertStrMap(
		t, "Tags", m.Tags, map[string]string{
			"badge-info":   "subscriber/15",
			"badges":       "subscriber/12,glhf-pledge/1",
			"color":        "#DAA520",
			"display-name": "Mm2PL",
			"emotes":       "",
			"flags":        "",
			"id":           "1d7e0b34-fe74-4895-92ae-dd912046e637",
			"mod":          "0",
			"room-id":      "11148817",
			"subscriber":   "1",
			"tmi-sent-ts":  "1632058935165",
			"turbo":        "0",
			"user-id":      "117691339",
			"user-type":    "",
		},
	)

	m, err = NewMessage(`@tag=spaces\sexist\sas\sdo\nnew\rlines\sand\:semicolons TEST`)
	assert(t, "error", err, nil)
	assert(t, "Action", m.Action, "TEST")
	assert(t, "Prefix", m.Prefix, "")
	assertStrSlc(t, "Args", m.Args, []string{})
	assertStrMap(
		t, "Tags", m.Tags, map[string]string{
			"tag": "spaces exist as do\nnew\rlines and;semicolons",
		},
	)
}

func BenchmarkNewMessage(b *testing.B) {
	file, err := os.Open("channel.txt")
	if err != nil {
		fmt.Println("You need to have a large logs of IRC messages named channel.txt for this benchmark to work.")
		fmt.Println("Failed to open file", err)
		b.Failed()
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			break
		}
		_, err = NewMessage(string(line))
		if err != nil {
			fmt.Printf("Failed at line: %s. Err=%s\n", line, err)
		}
	}
}
func getTestMessage() *Message {
	return &Message{
		Raw:    "@badge-info=subscriber/15;badges=subscriber/12,glhf-pledge/1;color=#DAA520;display-name=Mm2PL;emotes=;flags=;id=1d7e0b34-fe74-4895-92ae-dd912046e637;mod=0;room-id=11148817;subscriber=1;tmi-sent-ts=1632058935165;turbo=0;user-id=117691339;user-type= :mm2pl!mm2pl@mm2pl.tmi.twitch.tv PRIVMSG #pajlada :-tags many words asdasd",
		Prefix: "mm2pl!mm2pl@mm2pl.tmi.twitch.tv",
		User:   "mm2pl",
		Args:   []string{"#pajlada", "-tags many words asdasd"},
		Action: "PRIVMSG",
		Tags: map[string]string{
			"badge-info":   "subscriber/15",
			"badges":       "subscriber/12,glhf-pledge/1",
			"color":        "#DAA520",
			"display-name": "Mm2PL",
			"emotes":       "",
			"flags":        "",
			"id":           "1d7e0b34-fe74-4895-92ae-dd912046e637",
			"mod":          "0",
			"room-id":      "11148817",
			"subscriber":   "1",
			"tmi-sent-ts":  "1632058935165",
			"turbo":        "0",
			"user-id":      "117691339",
			"user-type":    "",
		},
		Timestamp: time.Date(2021, 9, 19, 15, 42, 15, 165, time.UTC),
	}
}

func testSerializeMessage(t *testing.T, rawIrc string) {
	m, err := NewMessage(rawIrc)
	assert(t, "error", err, nil)
	assert(t, "serialized output", m.Serialize(), rawIrc+"\r\n")
}
func TestMessage_Serialize(t *testing.T) {
	testSerializeMessage(
		t,
		"@badge-info=subscriber/15;badges=subscriber/12,glhf-pledge/1;color=#DAA520;display-name=Mm2PL;emotes=;flags=;id=1d7e0b34-fe74-4895-92ae-dd912046e637;mod=0;room-id=11148817;subscriber=1;tmi-sent-ts=1632058935165;turbo=0;user-id=117691339;user-type= :mm2pl!mm2pl@mm2pl.tmi.twitch.tv PRIVMSG #pajlada :-tags many words asdasd",
	)

	// test escaping tags
	testSerializeMessage(t, `@tag=spaces\sexist\sas\sdo\nnew\rlines\sand\:semicolons TEST`)
}

func BenchmarkMessage_Serialize(b *testing.B) {
	m := getTestMessage()
	for i := 0; i < 1000; i++ {
		_ = m.Serialize()
	}
}
