package http

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
)

const (
	sendMessagePath = "sendMessage"
	sendPhotoPath   = "sendPhoto"
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

func (t *Tg) SendPhoto(chatID int, caption string, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	_ = writer.WriteField("chat_id", strconv.Itoa(chatID))
	_ = writer.WriteField("caption", caption)

	part, err := writer.CreateFormFile("photo", filepath.Base(filename))
	if err != nil {
		return err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}

	err = writer.Close()
	if err != nil {
		return err
	}

	u := url.URL{
		Scheme: "https",
		Host:   t.url,
		Path:   path.Join("bot"+t.token, sendPhotoPath),
	}

	req, err := http.NewRequest("POST", u.String(), &b)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	_, err = http.DefaultClient.Do(req)
	return err
}
