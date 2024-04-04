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
	"time"
)

type log struct {
	Time     time.Time `json:"time"`
	Message  string    `json:"message"`
	Hostname string    `json:"hostname"`
	Severity string    `json:"severity"`
	Program  string    `json:"program"`
}

type pageInfo struct {
	PrevPage string `json:"prevPage"`
	NextPage string `json:"nextPage"`
}

type response struct {
	Logs     []log    `json:"logs"`
	PageInfo pageInfo `json:"pageInfo"`
}

func main() {
	lines := flag.Int("n", 5, "number of log lines that should be fetched")
	wholeMessage := flag.Bool("json", false, "print whole response as a json")
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
	params.Add("pageSize", strconv.Itoa(*lines))

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

	var jsonContent response
	err = json.Unmarshal(content, &jsonContent)
	if err != nil {
		slog.Error("Could not unmarshal response body to json")
		os.Exit(1)
	}

	if *wholeMessage {
		j, err := json.Marshal(jsonContent)
		if err != nil {
			slog.Error("Could not marshal response", "error", err)
			os.Exit(1)
		}
	
		fmt.Println(string(j))
	} else {
		for _, e := range jsonContent.Logs {
			fmt.Println(e.Message)
		}
	}
}
