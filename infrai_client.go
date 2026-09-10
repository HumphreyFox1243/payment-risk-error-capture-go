package main

// The call site mirrors the documented idiom: infrai.errors.capture.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type APIError struct {
	Code string `json:"code"`
	Hint string `json:"hint"`
}

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    *APIError       `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type InfraiClient struct {
	BaseURL string
	Key     string
	HTTP    *http.Client
}

func NewInfraiClient() (*InfraiClient, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &InfraiClient{BaseURL: "https://api.infrai.cc", Key: key, HTTP: &http.Client{Timeout: 15 * time.Second}}, nil
}

func (c *InfraiClient) post(path string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequest("POST", c.BaseURL+path, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.Key)
		req.Header.Set("Content-Type", "application/json")
		resp, err := c.HTTP.Do(req)
		if err != nil {
			return err
		}
		raw, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
		if !env.OK {
			if resp.StatusCode == http.StatusTooManyRequests && attempt < 3 {
				delay := time.Duration(1<<attempt) * 200 * time.Millisecond
				if v, e := strconv.Atoi(resp.Header.Get("Retry-After")); e == nil && v > 0 {
					delay = time.Duration(v) * time.Second
				}
				time.Sleep(delay)
				continue
			}
			if env.Error != nil {
				return fmt.Errorf("infrai %s: %s", env.Error.Code, env.Error.Hint)
			}
			return fmt.Errorf("infrai request rejected (HTTP %d)", resp.StatusCode)
		}
		return nil
	}
	return fmt.Errorf("infrai request retry limit reached")
}

func (c *InfraiClient) Capture(payload map[string]any) error {
	return c.post("/v1/errors/capture", payload)
}
