package http

import (
	"net/http"
	"net/url"
	"path"
	"strconv"
)

const (
	sendMessagePath = "sendMessage"
)

type Tg struct {
	url   string
	token string
}

func NewTG(url string, token string) *Tg {
	return &Tg{url: url, token: token}
}

func (t *Tg) SendMessage(chatID int, text string) error {
	q := url.Values{}
	q.Add("chat_id", strconv.Itoa(chatID))
	q.Add("text", text)

	u := url.URL{
		Scheme: "https",
		Host:   t.url,
		Path:   path.Join("bot"+t.token, sendMessagePath),
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return err
	}

	req.URL.RawQuery = q.Encode()
	_, err = http.DefaultClient.Do(req)
	return err
}
