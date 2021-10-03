package justgrep

import (
	"bufio"
	"fmt"
	"os"
	"testing"
)

func assert(t *testing.T, what string, have interface{}, expect interface{}) {
	if have != expect {
		t.Errorf("assertion on %s failed: have %q, expected %q", what, have, expect)
	}
}

func assertStrSlc(t *testing.T, what string, have []string, expect []string) {
	if len(have) != len(expect) {
		t.Errorf("assertion on %s failed: length doesn't match, have %d, expected %d: %q vs %q", what, len(have), len(expect), have, expect)
	}
	for i, elem := range have {
		if elem != expect[i] {
			t.Errorf("assertion on %s[%d] failed: have %q, expected %q", what, i, elem, expect[i])
		}
	}
}
func assertStrMap(t *testing.T, what string, have map[string]string, expect map[string]string) {
	if len(have) != len(expect) {
		t.Errorf("assertion on %s failed: length doesn't match, have %d, expected %d: %q vs %q", what, len(have), len(expect), have, expect)
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
	assertStrMap(t, "Tags", m.Tags, map[string]string{
		"badge-info": "subscriber/15",
		"badges": "subscriber/12,glhf-pledge/1",
		"color": "#DAA520",
		"display-name": "Mm2PL",
		"emotes": "",
		"flags": "",
		"id": "1d7e0b34-fe74-4895-92ae-dd912046e637",
		"mod": "0",
		"room-id": "11148817",
		"subscriber": "1",
		"tmi-sent-ts": "1632058935165",
		"turbo": "0",
		"user-id": "117691339",
		"user-type": "",
	})
}

func BenchmarkNewMessage(b *testing.B) {
	b.ReportAllocs()
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
		_, _ = NewMessage(string(line))
	}
}
