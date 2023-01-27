package justgrep

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type JustlogAPI interface {
	// MakeURL creates a URL to download the data from justlog
	MakeURL(date time.Time) string

	// NextLogFile is deprecated. It returns currentDate.Add(api.GetApproximateOffset)
	NextLogFile(currentDate time.Time) time.Time

	// GetApproximateOffset describes roughly how often new files are made in justlog for this api.
	// This function shouldn't be treated as anything more than a UI suggestion, use GetAvailableLogs for precise data
	// instead
	GetApproximateOffset() time.Duration

	// GetAvailableLogs fetches logs available from justlog
	GetAvailableLogs(ctx context.Context, client *http.Client) (LogsList, error)
}

type ProgressState struct {
	TotalResults []int `json:"total_results"`

	CountLines int `json:"count_lines"`
	CountBytes int `json:"count_bytes"`

	BeginTime time.Time `json:"begin_time"`
}

func fetch(ctx context.Context, url string, client *http.Client, output chan *Message, progress *ProgressState) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", UserAgent)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
		if scanner.Scan() {
			return errors.New(fmt.Sprintf("justlog instance responded with %d: %q", resp.StatusCode, scanner.Text()))
		}
		return errors.New(fmt.Sprintf("justlog instance responded with unexpected %d status code", resp.StatusCode))
	}

	go func() {
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)

		for scanner.Scan() {
			msg, err := NewMessage(scanner.Text())
			progress.CountLines += 1
			if err != nil {
				output <- nil
				_, _ = fmt.Fprintf(os.Stderr, "Error while fetching from %s: %s\n", url, err)
				break
			}
			progress.CountBytes += len(msg.Raw)
			output <- msg
			if ctx.Err() != nil {
				break
			}
		}
		close(output)
	}()
	return nil
}

func FetchForDate(
	ctx context.Context,
	api JustlogAPI,
	date time.Time,
	output chan *Message,
	progress *ProgressState,
	client *http.Client,
) (time.Time, error) {
	u := api.MakeURL(date)
	err := fetch(ctx, u, client, output, progress)
	if err != nil {
		return time.Time{}, err
	} else {
		return api.NextLogFile(date), nil
	}
}

func FetchForLogEntry(
	ctx context.Context,
	api JustlogAPI,
	logs AvailableLogEntry,
	output chan *Message,
	progress *ProgressState,
	client *http.Client,
) error {
	_, err := FetchForDate(ctx, api, logs.ToDate(), output, progress, client)
	return err
}

type UserJustlogAPI struct {
	JustlogAPI

	Channel string
	User    string
	URL     string
	IsId    bool
}

func (api UserJustlogAPI) NextLogFile(currentDate time.Time) time.Time {
	return currentDate.AddDate(0, -1, 0)
}

func (api UserJustlogAPI) MakeURL(date time.Time) string {
	if api.IsId {
		return fmt.Sprintf(
			"%s/channel/%s/userid/%s/%d/%d?raw&reverse",
			api.URL,
			api.Channel,
			api.User,
			date.Year(),
			date.Month(),
		)
	}
	return fmt.Sprintf(
		"%s/channel/%s/user/%s/%d/%d?raw&reverse",
		api.URL,
		api.Channel,
		api.User,
		date.Year(),
		date.Month(),
	)
}

func (api UserJustlogAPI) GetApproximateOffset() time.Duration {
	return time.Hour * 24 * 30
}

type ChannelJustlogAPI struct {
	JustlogAPI
	Channel string
	URL     string
}

func (api ChannelJustlogAPI) NextLogFile(currentDate time.Time) time.Time {
	return currentDate.AddDate(0, 0, -1)
}

func (api ChannelJustlogAPI) MakeURL(date time.Time) string {
	return fmt.Sprintf(
		"%s/channel/%s/%d/%d/%d?raw&reverse",
		api.URL,
		api.Channel,
		date.Year(),
		date.Month(),
		date.Day(),
	)
}

type channelsResp struct {
	Channels []struct {
		UserID string `json:"userID"`
		Name   string `json:"name"`
	} `json:"channels"`
}

func GetChannelsFromJustLog(ctx context.Context, client *http.Client, url string) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url+"/channels", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	output := channelsResp{}
	err = json.NewDecoder(resp.Body).Decode(&output)
	if err != nil {
		return nil, err
	}
	channels := make([]string, 0, 32)
	for _, channel := range output.Channels {
		channels = append(channels, channel.Name)
	}
	return channels, nil
}

// AvailableLogEntry describes an element from justlog's /list api array
type AvailableLogEntry struct {
	RawYear  string `json:"year"`
	RawMonth string `json:"month"`

	// Only for /list without a user
	RawDay string `json:"day"`

	Year  int
	Month int

	// As with RawDay, Day only makes sense for non-user logs, otherwise will be 0
	Day int
}

func (l *AvailableLogEntry) Parse() error {
	if l.Year != 0 {
		return nil
	}
	year, err := strconv.ParseInt(l.RawYear, 10, 64)
	if err != nil {
		return err
	}
	month, err := strconv.ParseInt(l.RawMonth, 10, 64)
	if err != nil {
		return err
	}
	l.Year = int(year)
	l.Month = int(month)
	if l.RawDay != "" {
		day, err := strconv.ParseInt(l.RawDay, 10, 64)
		if err != nil {
			return err
		}
		l.Day = int(day)
	}
	return nil
}

// ToDate converts the AvailableLogEntry to a time.Time with 0 seconds past midnight on the day (or the first of the month).
// It may panic if Parse() wasn't called before and parsing fails
func (l *AvailableLogEntry) ToDate() time.Time {
	if err := l.Parse(); err != nil {
		log.Panicf(
			"Unexpectidly errored while converting a log entry to a date: %s, "+
				"call justgrep.AvailableLogEntry.Parse() explicitly to avoid this",
			err,
		)
	}
	day := l.Day
	if l.Day == 0 {
		day = 1
	}
	return time.Date(
		l.Year,
		time.Month(l.Month),
		day,
		0,
		0,
		0,
		0,
		time.UTC,
	)
}

type availableLogsResponse struct {
	AvailableLogs []AvailableLogEntry `json:"availableLogs"`
}
type LogsList []AvailableLogEntry

func (l LogsList) EnsureParsed() error {
	for i, logs := range l {
		err := logs.Parse()
		if err != nil {
			return err
		}
		// copy by value in iterator?
		l[i] = logs
	}
	return nil
}

func (l LogsList) Snip(early time.Time, late time.Time) (LogsList, error) {
	if early.After(late) {
		log.Panicf("THIS SHOULD NOT HAPPEN, early > late: %#v > %#v!!!", early, late)
	}
	err := l.EnsureParsed()
	if err != nil {
		return nil, err
	}
	out := LogsList{}

	for _, logs := range l {
		if logs.RawDay != "" { // will not be present on non-user requests
			dayBegin := time.Date(
				logs.Year,
				time.Month(logs.Month),
				logs.Day,
				0,
				0,
				0,
				0,
				time.UTC,
			)
			dayEnd := dayBegin.AddDate(0, 0, 1).Add(-time.Second)

			if fitsInRange(dayBegin, early, late) || fitsInRange(dayEnd, early, late) {
				out = append(out, logs)
			}
		} else {
			monthBegin := time.Date(
				logs.Year,
				time.Month(logs.Month),
				1,
				0,
				0,
				0,
				0,
				time.UTC,
			)
			monthEnd := monthBegin.AddDate(0, 1, 0).Add(-time.Second)
			if fitsInRange(late, monthBegin, monthEnd) ||
				fitsInRange(early, monthBegin, monthEnd) ||
				fitsInRange(monthEnd, early, late) {
				out = append(out, logs)
			}
		}
	}
	return out, nil
}

// fitsInRange checks if early < t < late
func fitsInRange(t time.Time, early time.Time, late time.Time) bool {
	if t.After(late) {
		return false
	}
	if t.Before(early) {
		return false
	}
	return true
}

func (api ChannelJustlogAPI) GetAvailableLogs(ctx context.Context, client *http.Client) (LogsList, error) {
	u, err := url.Parse(api.URL)
	if err != nil {
		return nil, err
	}
	list, _ := u.Parse("/list")
	q := list.Query()
	q.Add("channel", api.Channel)

	list.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", list.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		errorBytes, _ := io.ReadAll(resp.Body)
		errorMsg := string(errorBytes)
		return nil, fmt.Errorf("%s: %s", resp.Status, strings.Trim(errorMsg, "\r\n"))
	}
	output := availableLogsResponse{}
	err = json.NewDecoder(resp.Body).Decode(&output)
	if err != nil {
		return nil, err
	}
	return output.AvailableLogs, nil

}

func (api UserJustlogAPI) GetAvailableLogs(ctx context.Context, client *http.Client) (LogsList, error) {
	u, err := url.Parse(api.URL)
	if err != nil {
		return nil, err
	}
	list, _ := u.Parse("/list")
	q := list.Query()
	q.Add("channel", api.Channel)
	if api.IsId {
		q.Add("userid", api.User)
	} else {
		q.Add("user", api.User)
	}

	list.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", list.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		errorBytes, _ := io.ReadAll(resp.Body)
		errorMsg := string(errorBytes)
		return nil, fmt.Errorf("%s: %s", resp.Status, strings.Trim(errorMsg, "\r\n"))
	}
	output := availableLogsResponse{}
	err = json.NewDecoder(resp.Body).Decode(&output)
	if err != nil {
		return nil, err
	}
	return output.AvailableLogs, nil

}

func (api ChannelJustlogAPI) GetApproximateOffset() time.Duration {
	return time.Hour * 24
}

var UserAgent = "justgrep/1.0 (log-searcher)"
