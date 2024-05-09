package swo

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewOptions(t *testing.T) {
	fixedTime, err := time.Parse(time.DateTime, "2000-01-01 10:00:30")
	require.NoError(t, err)

	testCases := []struct {
		name          string
		flags         []string
		action        func()
		expected      Options
		expectedError error
	}{
		{
			name:  "default flag values",
			flags: []string{},
			expected: Options{
				args:       []string{},
				count:      defaultCount,
				configFile: defaultConfigFile,
				ApiUrl:     defaultApiUrl,
			},
		},
		{
			name:  "many flags",
			flags: []string{"--count", "5", "--group", "groupValue", "--system", "systemValue", "--color", "program"},
			expected: Options{
				args:       []string{},
				count:      5,
				group:      "groupValue",
				system:     "systemValue",
				color:      program,
				configFile: defaultConfigFile,
				ApiUrl:     defaultApiUrl,
			},
		},
		{
			name:  "many flags and args",
			flags: []string{"--count", "5", "--group", "groupValue", "one", "two", "three"},
			expected: Options{
				args:       []string{"one", "two", "three"},
				count:      5,
				group:      "groupValue",
				configFile: defaultConfigFile,
				ApiUrl:     defaultApiUrl,
			},
		},
		{
			name:          "invalid color value",
			flags:         []string{"--color", "yellow"},
			expected:      Options{},
			expectedError: errColorFlag,
		},
		{
			name:  "read full config file",
			flags: []string{"--configfile", filepath.Join(os.TempDir(), "config-file.yaml")},
			expected: Options{
				args:       []string{},
				count:      defaultCount,
				configFile: filepath.Join(os.TempDir(), "config-file.yaml"),
				ApiUrl:     "https://api.solarwinds.com",
				Token:      "123456",
			},
			action: func() {
				f, err := os.Create(filepath.Join(os.TempDir(), "config-file.yaml"))
				require.NoError(t, err, "creating a temporary file should not fail")

				yamlStr := `
token: 123456
api-url: https://api.solarwinds.com
`
				n, err := f.Write([]byte(yamlStr))
				require.Equal(t, n, len(yamlStr))
				require.NoError(t, err)
			},
		},
		{
			name:  "read token from config file",
			flags: []string{"--configfile", filepath.Join(os.TempDir(), "config-file.yaml")},
			expected: Options{
				args:       []string{},
				count:      defaultCount,
				configFile: filepath.Join(os.TempDir(), "config-file.yaml"),
				ApiUrl:     defaultApiUrl,
				Token:      "123456",
			},
			action: func() {
				f, err := os.Create(filepath.Join(os.TempDir(), "config-file.yaml"))
				require.NoError(t, err, "creating a temporary file should not fail")

				yamlStr := `
token: 123456
`
				n, err := f.Write([]byte(yamlStr))
				require.Equal(t, n, len(yamlStr))
				require.NoError(t, err)
			},
		},
		{
			name:  "read token from env var",
			flags: []string{},
			expected: Options{
				args:       []string{},
				count:      defaultCount,
				configFile: defaultConfigFile,
				ApiUrl:     defaultApiUrl,
				Token:      "tokenFromEnvVar",
			},
			action: func() {
				err := os.Setenv("SWOKEN", "tokenFromEnvVar")
				require.NoError(t, err)
			},
		},
		{
			name:  "parse min time",
			flags: []string{"--min-time", "5 seconds ago"},
			expected: Options{
				args:       []string{},
				count:      defaultCount,
				configFile: defaultConfigFile,
				ApiUrl:     defaultApiUrl,
				minTime:    "2000-01-01T10:00:25Z",
			},
			action: func() {
				now = fixedTime
			},
		},
		{
			name:  "parse max time",
			flags: []string{"--max-time", "in 5 seconds"},
			expected: Options{
				args:       []string{},
				count:      defaultCount,
				configFile: defaultConfigFile,
				ApiUrl:     defaultApiUrl,
				maxTime:    "2000-01-01T10:00:35Z",
			},
			action: func() {
				now = fixedTime
			},
		},
		{
			name:  "fail parsing min time",
			flags: []string{"--min-time", "what?"},
			action: func() {
				now = fixedTime
			},
			expectedError: errMinTimeFlag,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_ = os.Remove(filepath.Join(os.TempDir(), "config-file.yaml"))
			if tc.action != nil {
				tc.action()
			}

			opts, err := NewOptions(tc.flags)

			// erase fs so we can compare structs
			tc.expected.fs = nil
			if opts != nil {
				opts.fs = nil
			}

			if opts != nil {
				require.Equal(t, tc.expected, *opts)
			}

			require.True(t, errors.Is(err, tc.expectedError))
		})

		os.Setenv("SWOKEN", "")
	}
}
