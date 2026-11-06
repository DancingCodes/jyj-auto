package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

const ocrURL = "http://ddddocr:9103"

type ocrRequest struct {
	Image string `json:"image"`
}

type ocrResponse struct {
	Text string `json:"text"`
}

func RecognizeCaptcha(ctx context.Context, image string) (string, error) {
	body, err := json.Marshal(ocrRequest{Image: image})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ocrURL+"/recognize", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", io.ErrUnexpectedEOF
	}

	var result ocrResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	result.Text = strings.TrimSpace(result.Text)
	if result.Text == "" {
		return "", io.ErrUnexpectedEOF
	}
	return result.Text, nil
}
