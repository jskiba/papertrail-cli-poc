package swo

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/user"
	"time"
	"strings"
	"path/filepath"

	"github.com/olebedev/when"
	"gopkg.in/yaml.v3"
)

const (
	program = "program"
	system  = "system"
	all     = "all"
	off     = "off"

	defaultCount      = 100
	defaultConfigFile = "~/.swo-cli.yaml"
	defaultApiUrl     = "https://api.na-01.cloud.solarwinds.com"
)

var (
	now = time.Now()

	errColorFlag   = errors.New("unknown value of the color flag")
	errMinTimeFlag = errors.New("failed to parse --min-time flag")
	errMaxTimeFlag = errors.New("failed to parse --max-time flag")
	errMissingToken = errors.New("failed to find token")

	timeLayouts = []string{
	time.Layout,
	time.ANSIC,
	time.UnixDate,
	time.RubyDate,
	time.RFC822,
	time.RFC822Z,
	time.RFC850,
	time.RFC1123,
	time.RFC1123Z,
	time.RFC3339,
	time.RFC3339Nano,
	time.Kitchen,
	time.Stamp,
	time.StampMilli,
	time.StampNano,
	time.DateTime,
	time.DateOnly,
	time.TimeOnly,
}
)

type Options struct {
	fs         *flag.FlagSet

	args       []string
	count      int
	configFile string
	group      string
	system     string
	maxTime    string
	minTime    string
	color      string
	json       bool
	version    bool

	ApiUrl string `yaml:"api-url"`
	Token  string `yaml:"token"`
}


func NewOptions(args []string) (*Options, error) {
	opts := &Options{
		fs: flag.NewFlagSet(os.Args[0], flag.ContinueOnError),
	}

	opts.fs.Usage = func() {
		fmt.Printf("%36s\n", "swo-cli - command-line search for SolarWinds Observability log management service")
		fmt.Printf("    %2s, %16s %70s\n", "-h", "--help", "Show usage")
		fmt.Printf("    %2s  %16s %70s\n", "", "--count NUMBER", "Number of log entries to search (100)")
		fmt.Printf("    %2s  %16s %70s\n", "", "--min-time MIN", "Earliest time to search from")
		fmt.Printf("    %2s  %16s %70s\n", "", "--max-time MAX", "Latest time to search from")
		fmt.Printf("    %2s, %16s %70s\n", "-c", "--configfile", "Path to config (~/.swo-cli.yaml)")
		fmt.Printf("    %2s, %16s %70s\n", "-g", "--group GROUP_ID", "Group ID to search")
		fmt.Printf("    %2s, %16s %70s\n", "-s", "--system SYSTEM", "System to search")
		fmt.Printf("    %2s, %16s %70s\n", "-j", "--json", "Output raw JSON data (off)")
		fmt.Printf("    %2s  %16s %70s\n", "", "--color [program|system|all|off]", "")
		fmt.Printf("    %2s, %16s %70s\n", "-V", "--version", "Display the version and exit")

		fmt.Println()

		fmt.Println("  Usage:")
		fmt.Println("    swo-cli [--min-time time] [--max-time time] [-g group-id] [-s system]")
		fmt.Println("      [-c swo-cli.yml] [-j] [--color attributes] [--] [query]")

		fmt.Println()

		fmt.Println("  Examples:")
		fmt.Println("  swo-cli something")
		fmt.Println("  swo-cli 1.2.3 Failure")
		fmt.Println(`  swo-cli -s ns1 "connection refused"`)
		fmt.Println(`  swo-cli "(www OR db) (nginx OR pgsql) -accepted"`)
		fmt.Println(`  swo-cli -g <SWO_GROUP_ID> --color all "(nginx OR pgsql) -accepted"`)
		fmt.Println(`  swo-cli --min-time 'yesterday at noon' --max-time 'today at 4am' -g <SWO_GROUP_ID>`)
		fmt.Println("  swo-cli -- -redis")
	}

	var minTime, maxTime string
	opts.fs.IntVar(&opts.count, "count", defaultCount, "")
	opts.fs.StringVar(&opts.configFile, "c", "", "")
	opts.fs.StringVar(&opts.configFile, "configfile", defaultConfigFile, "")
	opts.fs.StringVar(&opts.group, "g", "", "")
	opts.fs.StringVar(&opts.group, "group", "", "")
	opts.fs.StringVar(&opts.system, "s", "", "")
	opts.fs.StringVar(&opts.system, "system", "", "")
	opts.fs.StringVar(&opts.color, "color", "", "")
	opts.fs.StringVar(&opts.ApiUrl, "api-url", defaultApiUrl, "")
	opts.fs.StringVar(&minTime, "min-time", "", "")
	opts.fs.StringVar(&maxTime, "max-time", "", "")
	opts.fs.BoolVar(&opts.json, "j", false, "")
	opts.fs.BoolVar(&opts.json, "json", false, "")
	opts.fs.BoolVar(&opts.version, "V", false, "")
	opts.fs.BoolVar(&opts.version, "version", false, "")

	err := opts.fs.Parse(args)
	if err != nil {
		return nil, err
	}

	opts.args = opts.fs.Args()

	if opts.color != "" {
		if !(opts.color == program || opts.color == system || opts.color == all || opts.color == off) {
			return nil, errColorFlag
		}
	}

	if minTime != "" {
		result, err := parseTime(minTime)
		if err != nil {
			return nil, errors.Join(errMinTimeFlag, err)
		}

		opts.minTime = result
	}

	if maxTime != "" {
		result, err := parseTime(maxTime)
		if err != nil {
			return nil, errors.Join(errMaxTimeFlag, err)
		}

		opts.maxTime = result
	}

	configPath := opts.configFile
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	localConfig := filepath.Join(cwd, ".swo-cli.yaml")
	if _, err := os.Stat(localConfig); err == nil {
		configPath = localConfig
	} else if strings.HasPrefix(opts.configFile, "~/") {
		usr, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("error while resolving current user to read configuration file: %w", err)
		}

		configPath = filepath.Join(usr.HomeDir, opts.configFile[2:])
	}

	if content, err := os.ReadFile(configPath); err == nil {
		err = yaml.Unmarshal(content, opts)
		if err != nil {
			return nil, fmt.Errorf("error while unmarshaling %s config file: %w", configPath, err)
		}
	}

	if token := os.Getenv("SWOKEN"); token != "" {
		opts.Token = token
	}

	if opts.Token == "" && !opts.version {
		return nil, errMissingToken
	}

	return opts, nil
}

func parseTime(input string) (string, error) {
	location := time.Local
	if strings.HasSuffix(input, "UTC") {
		l, err := time.LoadLocation("UTC")
		if err != nil {
			return "", err
		}

		location = l

		input = strings.ReplaceAll(input, "UTC", "")
	}

	for _, layout := range timeLayouts {
		result, err := time.Parse(layout, input)
		if err == nil {
			result = result.In(location)
			return result.Format(time.RFC3339), nil
		}
	}

	result, err := when.EN.Parse(input, now)
	if err != nil {
		return "", err
	}
	if result == nil {
		return "", errors.New("failed to parse time")
	}

	return result.Time.In(location).Format(time.RFC3339), nil
}