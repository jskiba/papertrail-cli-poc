package swo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPrepareRequest(t *testing.T) {
	configFile := filepath.Join(os.TempDir(), "config-file.yaml")
	f, err := os.Create(configFile)
	require.NoError(t, err, "creating a temporary file should not fail")
	defer os.Remove(configFile)

	token := "1234567"
	yamlStr := fmt.Sprintf("token: %s", token)
	n, err := f.Write([]byte(yamlStr))
	require.Equal(t, n, len(yamlStr))
	require.NoError(t, err)

	fixedTime, err := time.Parse(time.DateTime, "2000-01-01 10:00:30")
	require.NoError(t, err)
	now = fixedTime

	testCases := []struct {
		name           string
		flags          []string
		expectedValues map[string][]string
		expectedError  error
	}{
		{
			name:  "default request",
			flags: []string{"--configfile", configFile},
			expectedValues: map[string][]string{
				"pageSize": {"10"},
			},
		},
		{
			name:  "custom pageSize group startTime and endTime",
			flags: []string{"--configfile", configFile, "--count", "8", "--group", "groupValue", "--min-time", "10 seconds ago", "--max-time", "2 seconds ago"},
			expectedValues: map[string][]string{
				"pageSize":  {"8"},
				"group":     {"groupValue"},
				"startTime": {"2000-01-01T10:00:20Z"},
				"endTime":   {"2000-01-01T10:00:28Z"},
			},
		},
		{
			name:  "system flag",
			flags: []string{"--configfile", configFile, "--system", "systemValue"},
			expectedValues: map[string][]string{
				"pageSize": {"10"},
				"filter":   {"host:systemValue"},
			},
		},
		{
			name:  "system flag with filter",
			flags: []string{"--configfile", configFile, "--system", "systemValue", "--", "\"access denied\"", "1.2.3.4", "-sshd"},
			expectedValues: map[string][]string{
				"pageSize": {"10"},
				"filter": func() []string {
					escaped := url.PathEscape("filter=host:systemValue \"access denied\" 1.2.3.4 -sshd")
					values, err := url.ParseQuery(escaped)
					require.NoError(t, err)
					value, ok := values["filter"]
					require.True(t, ok)
					return value
				}(),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts, err := NewOptions(tc.flags)
			require.NoError(t, err)
			client, err := NewClient(opts)
			require.NoError(t, err)

			request, err := client.prepareRequest(context.Background())
			require.NoError(t, err)

			values := request.URL.Query()
			for k, v := range tc.expectedValues {
				require.ElementsMatch(t, v, values[k])
			}

			header := request.Header
			for k, v := range map[string][]string{
				"Authorization": {fmt.Sprintf("Bearer %s", token)},
				"Accept":        {"application/json"},
			} {
				require.ElementsMatch(t, v, header[k])
			}
		})
	}

}

func TestRun(t *testing.T) {
	configFile := filepath.Join(os.TempDir(), "config-file.yaml")
	f, err := os.Create(configFile)
	require.NoError(t, err, "creating a temporary file should not fail")
	defer os.Remove(configFile)

	logsData := LogsData{
		Logs: []Log{
			{
				Time:     "timeValueOne",
				Message:  "messageOne",
				Hostname: "hostnameOne",
				Severity: "severityOne",
				Program:  "programOne",
			},
			{
				Time:     "timeValueTwo",
				Message:  "messageTwo",
				Hostname: "hostnameTwo",
				Severity: "severityTwo",
				Program:  "programTwo",
			},
		},
		PageInfo: PageInfo{PrevPage: "prevPageValue"},
	}

	handler := func(responseWriter http.ResponseWriter, _ *http.Request) {
		data, err := json.Marshal(logsData)
		require.NoError(t, err)

		_, err = responseWriter.Write(data)
		require.NoError(t, err)
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	token := "1234567"
	yamlStr := fmt.Sprintf(`
token: %s
api-url: %s
`, token, server.URL)
	n, err := f.Write([]byte(yamlStr))
	require.Equal(t, n, len(yamlStr))
	require.NoError(t, err)

	originalStdout := os.Stdout
	defer func() {
		os.Stdout = originalStdout
	}()

	r, w, err := os.Pipe()
	require.NoError(t, err)

	opts, err := NewOptions([]string{"--configfile", configFile, "--json"})
	require.NoError(t, err)

	client, err := NewClient(opts)
	require.NoError(t, err)
	client.output = w

	go func() {
		output, err := io.ReadAll(r)
		require.NoError(t, err)
		require.NoError(t, err)
		require.Equal(t, logsData, output)
	}()

	err = client.Run(context.Background())
	require.NoError(t, err)
}

func TestPrintResult(t *testing.T) {

}
