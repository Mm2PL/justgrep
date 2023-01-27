package justgrep

import (
	"testing"
	"time"
)

func TestLogsList_Snip(t *testing.T) {
	l := LogsList{
		AvailableLogEntry{
			RawYear:  "2022",
			RawMonth: "12",
		},
		AvailableLogEntry{
			RawYear:  "2022",
			RawMonth: "11",
		},
		AvailableLogEntry{
			RawYear:  "2022",
			RawMonth: "10",
		},
		AvailableLogEntry{
			RawYear:  "2022",
			RawMonth: "9",
		},
		AvailableLogEntry{
			RawYear:  "2022",
			RawMonth: "8",
		},
		AvailableLogEntry{
			RawYear:  "2022",
			RawMonth: "7",
		},
		AvailableLogEntry{
			RawYear:  "2022",
			RawMonth: "6",
		},
		AvailableLogEntry{
			RawYear:  "2022",
			RawMonth: "5",
		},
		AvailableLogEntry{
			RawYear:  "2022",
			RawMonth: "4",
		},
		AvailableLogEntry{
			RawYear:  "2022",
			RawMonth: "3",
		},
		AvailableLogEntry{
			RawYear:  "2022",
			RawMonth: "2",
		},
		AvailableLogEntry{
			RawYear:  "2022",
			RawMonth: "1",
		},
	}
	err := l.EnsureParsed()
	assert(t, "err", err, nil)

	have, err := l.Snip(
		time.Date(2022, 4, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2022, 10, 1, 0, 0, 0, 0, time.UTC),
	)
	assert(t, "err", err, nil)
	expect := LogsList{
		AvailableLogEntry{
			RawYear:  "2022",
			RawMonth: "10",
		},
		AvailableLogEntry{
			RawYear:  "2022",
			RawMonth: "9",
		},
		AvailableLogEntry{
			RawYear:  "2022",
			RawMonth: "8",
		},
		AvailableLogEntry{
			RawYear:  "2022",
			RawMonth: "7",
		},
		AvailableLogEntry{
			RawYear:  "2022",
			RawMonth: "6",
		},
		AvailableLogEntry{
			RawYear:  "2022",
			RawMonth: "5",
		},
		AvailableLogEntry{
			RawYear:  "2022",
			RawMonth: "4",
		},
	}
	expect.EnsureParsed()
	if len(have) != len(expect) {
		t.Errorf(
			"assertion on LogsListSnip failed: length doesn't match, have %d, expected %d: %q vs %q",
			len(have),
			len(expect),
			have,
			expect,
		)
	}
	for i, elem := range have {
		if elem != expect[i] {
			t.Errorf("assertion on LogsListSnip[%d] failed: have %#v, expected %#v", i, elem, expect[i])
		}
	}
}
