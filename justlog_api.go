package justgrep

import (
	"bufio"
	"fmt"
	"net/http"
	"time"
)

type JustlogAPI interface {
	MakeURL(date time.Time) string
	NextLogFile(currentDate time.Time) time.Time
}

func fetch(url string, output chan *Message) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	go func() {
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)

		for scanner.Scan() {
			output <- NewMessage(scanner.Text())
		}
		output <- nil
	}()
	return nil
}

func FetchForDate(api JustlogAPI, date time.Time, output chan *Message) (time.Time, error) {
	url := api.MakeURL(date)
	fmt.Printf("Fetching %s from %s\n", date.Format(time.RFC3339), url)
	err := fetch(url, output)
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
		return fmt.Sprintf("%s/channel/%s/userid/%s/%d/%d?raw", api.URL, api.Channel, api.User, date.Year(), date.Month())
	}
	return fmt.Sprintf("%s/channel/%s/user/%s/%d/%d?raw", api.URL, api.Channel, api.User, date.Year(), date.Month())
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
	return fmt.Sprintf("%s/channel/%s/%d/%d/%d?raw", api.URL, api.Channel, date.Year(), date.Month(), date.Day())
}
