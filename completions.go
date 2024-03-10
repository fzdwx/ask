package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type completionsCallback func(r *CompletionsResponse, done bool, err error)

func completions(text, model, token, apiUrl string, cb completionsCallback) error {
	return (&Completions{
		Text:   text,
		Model:  model,
		Token:  token,
		ApiUrl: apiUrl,
	}).Call(cb)
}

type Completions struct {
	Text   string
	Model  string
	Token  string
	ApiUrl string
}

func (c *Completions) Call(cb completionsCallback) error {
	client := http.DefaultClient
	client.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}

	body, err := c.buildBody()
	if err != nil {
		return fmt.Errorf("building request body: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/v1/chat/completions", c.ApiUrl), bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	with(func(r []byte, done bool, err error) {
		if err != nil {
			cb(nil, done, err)
			return
		}
		if done {
			cb(nil, true, nil)
			return
		}

		text := strings.ReplaceAll(string(r), "data: ", "")
		if text == "[DONE]" {
			cb(nil, true, nil)
			return
		}
		if text == "" {
			return
		}

		var resp CompletionsResponse
		if err := json.Unmarshal([]byte(text), &resp); err != nil {
			cb(nil, false, fmt.Errorf("unmarshalling response: %w", err))
			return
		}

		cb(&resp, false, nil)
	})(client.Do(req))
	return nil
}

func (c *Completions) buildBody() ([]byte, error) {
	body := CompletionsBody{
		FrequencyPenalty: 0,
		Temperature:      0.7,
		PresencePenalty:  0,
		Model:            c.Model,
		Stream:           true,
		TopP:             1,
		Messages: []CompletionsMessage{
			{
				Role:    "user",
				Content: c.Text,
			},
		},
	}

	return json.Marshal(body)
}

type CompletionsBody struct {
	FrequencyPenalty float64              `json:"frequency_penalty"`
	Temperature      float64              `json:"temperature"`
	PresencePenalty  float64              `json:"presence_penalty"`
	Model            string               `json:"model"`
	Stream           bool                 `json:"stream"`
	Messages         []CompletionsMessage `json:"messages"`
	TopP             float64              `json:"top_p"`
}

type CompletionsMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type CompletionsResponse struct {
	Id                string      `json:"id"`
	Object            string      `json:"object"`
	Created           int         `json:"created"`
	Model             interface{} `json:"model"`
	SystemFingerprint string      `json:"system_fingerprint"`
	Choices           []struct {
		Index int `json:"index"`
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason interface{} `json:"finish_reason"`
	} `json:"choices"`
}
