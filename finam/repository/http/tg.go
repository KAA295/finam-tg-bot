package http

import (
	"bytes"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"
)

const (
	sendMessagePath = "sendMessage"
	sendPhotoPath   = "sendPhoto"
	retriesCount    = 4
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

	t.sendRequest(req)
	return nil
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

	t.sendRequest(req)

	return nil
}

func (t *Tg) sendRequest(req *http.Request) {
	client := http.Client{
		Timeout: 100 * time.Second,
	}
	bodyBytes, _ := io.ReadAll(req.Body)
	req.Body.Close()
	backoff := 2
	for i := 0; i < retriesCount; i++ {
		newReq := req.Clone(req.Context())

		newReq.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		resp, err := client.Do(newReq)
		if err != nil {
			log.Printf("[ERR] %s %s request failed: %v", newReq.Method, newReq.URL, err)
		}

		if resp.StatusCode == 200 {
			log.Printf("%s %s request; attempt: %s; success", newReq.Method, newReq.URL, i+1, resp.StatusCode)
			return
		}
		log.Printf("%s %s request; attempt: %s; status code: %s;", newReq.Method, newReq.URL, i+1, resp.StatusCode)

		backoff := backoff * backoff

		time.Sleep(time.Duration(backoff) * time.Second)
	}
}
