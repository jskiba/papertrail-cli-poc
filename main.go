package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

func main() {
	lines := flag.Int("lines", 5, "number of log lines that should be fetched")
	endpoint := flag.String("url", "https://api.na-01.cloud.solarwinds.com", "api url")
	flag.Parse()
	token := os.Getenv("SWOKEN")
	if token == "" {
		slog.Error("SWOKEN env var is empty")
		os.Exit(1)
	}

	client := http.DefaultClient

	urlStr, err := url.JoinPath(*endpoint, "v1/logs")
	if err != nil {
		slog.Error("Could not parse endpoint", "error", err)
		os.Exit(1)
	}

	parsedUrl, err := url.Parse(urlStr)
	if err != nil {
		slog.Error("Could not parse endpoint", "error", err)
		os.Exit(1)
	}

	params := url.Values{}
	params.Add("requestPageSize", strconv.Itoa(*lines))

	parsedUrl.RawQuery = params.Encode()

	req, err := http.NewRequest("GET", parsedUrl.String(), nil)
	if err != nil {
		os.Exit(1)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("accept", "application/json")

	resp, err := client.Do(req)
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

	fmt.Println(string(content))

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
