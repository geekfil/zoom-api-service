package telegram

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

const telegramApiEndpoint = "https://api.telegram.org/bot%s/%s"

type Config struct {
	ChatId  string
	Token   string
	Timeout time.Duration
}

type SendError struct {
	Date  time.Time `json:"date"`
	Error error     `json:"error"`
}

type Telegram struct {
	config     *Config
	SendErrors []SendError
}

func NewConfig() *Config {
	return &Config{
		ChatId:  os.Getenv("TELEGRAM_CHAT_ID"),
		Token:   os.Getenv("TELEGRAM_TOKEN"),
		Timeout: time.Second * 3,
	}
}

func New(config *Config) *Telegram {

	return &Telegram{
		config:     config,
		SendErrors: []SendError{},
	}
}

func (t *Telegram) Send(text string) (*http.Response, error) {
	params := url.Values{}
	params.Set("chat_id", t.config.ChatId)
	params.Set("text", text)
	var res *http.Response
	var err error
	if res, err = http.PostForm(fmt.Sprintf(telegramApiEndpoint, t.config.Token, "sendMessage"), params); err != nil {
		t.SendErrors = append(t.SendErrors, SendError{
			Date:  time.Now(),
			Error: err,
		})
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		err = errors.New("Response status is " + res.Status)
		t.SendErrors = append(t.SendErrors, SendError{
			Date:  time.Now(),
			Error: err,
		})
		return nil, err
	}
	resCopy := res
	return resCopy, nil
}
