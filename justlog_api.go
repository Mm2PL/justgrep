package justgrep

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
)

type JustlogAPI interface {
	MakeURL(date time.Time) string
	NextLogFile(currentDate time.Time) time.Time
	GetApproximateOffset() time.Duration
}

type ProgressState struct {
	TotalResults []int `json:"total_results"`

	CountLines int `json:"count_lines"`
	CountBytes int `json:"count_bytes"`

	BeginTime time.Time `json:"begin_time"`
}

func fetch(ctx context.Context, url string, output chan *Message, progress *ProgressState) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	client := http.Client{}
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
) (time.Time, error) {
	url := api.MakeURL(date)
	err := fetch(ctx, url, output, progress)
	if err != nil {
		return time.Time{}, err
	} else {
		return api.NextLogFile(date), nil
	}
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

func GetChannelsFromJustLog(ctx context.Context, url string) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url+"/channels", nil)
	if err != nil {
		return nil, err
	}
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
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

func (api ChannelJustlogAPI) GetApproximateOffset() time.Duration {
	return time.Hour * 24
}
