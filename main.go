package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"flag"
)

func main() {
	lines := flag.Int("lines", 0, "number of log lines to fetch")
	urlStr := *flag.String("url", "https://api.na-01.cloud.solarwinds.com", "swo api url (default: https://api.na-01.cloud.solarwinds.com)")
	token := os.Getenv("PAPERTRAIL_TOKEN")
	if token == "" {
		slog.Error("PAPERTRAIL_TOKEN env var is empty")
		os.Exit(1)
	}

	client := http.DefaultClient

	urlStr, err := url.JoinPath(urlStr, "v1/logs")
	if err != nil {
		slog.Error("Could not parse endpoint", "error", err)
		os.Exit(1)
	}

	endpoint, err := url.Parse(urlStr)
	if err != nil {
		slog.Error("Could not parse endpoint", "error", err)
		os.Exit(1)
	}

	header := http.Header{}
	header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	header.Add("accept", "application/json")

	data, err := json.Marshal(map[string]any{
		"requestPageSize": lines,
	})
	if err != nil {
		slog.Error("Could not parse request parameters", "error", err)
		os.Exit(1)
	}

	resp, err := client.Do(&http.Request{
		Method: "GET",
		URL:    endpoint,
		Header: header,
		Body:   io.NopCloser(bytes.NewReader(data)),
	})
	if err != nil {
		slog.Error("Could not send https request", "error", err)
		os.Exit(1)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			slog.Error("Could not close http body", "error", err)
			os.Exit(1)
		}
	}()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Could not read http response", "error", err)
		os.Exit(1)
	}

	type event struct {
		Message string `json:"message"`
	}
	type response struct {
		Events []event `json:"events"`
	}

	var jsonContent response
	err = json.Unmarshal(content, &jsonContent)
	if err != nil {
		slog.Error("Could not unmarshal response body to json")
		os.Exit(1)
	}

	for _, e := range jsonContent.Events {
		fmt.Println(e.Message)
	}
}
