package swo

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/olebedev/when"
	"gopkg.in/yaml.v3"
)

const (
	program = "program"
	system  = "system"
	all     = "all"
	off     = "off"

	defaultCount      = 10
	defaultConfigFile = "~/.swo-cli.yaml"
	defaultApiUrl     = "https://api.na-01.cloud.solarwinds.com"
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
	forceColor bool
	version    bool

	ApiUrl string `yaml:"api-url"`
	Token  string `yaml:"token"`
}

var (
	now = time.Now()

	errColorFlag   = errors.New("unknown value of the color flag")
	errMinTimeFlag = errors.New("failed to parse --min-time flag")
	errMaxTimeFlag = errors.New("failed to parse --max-time flag")
)

func NewOptions(args []string) (*Options, error) {
	opts := &Options{
		fs: flag.NewFlagSet(os.Args[0], flag.ContinueOnError),
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
	opts.fs.BoolVar(&opts.forceColor, "force-color", false, "")
	opts.fs.BoolVar(&opts.version, "V", false, "")
	opts.fs.BoolVar(&opts.version, "version", false, "")

	opts.fs.Parse(args)

	opts.args = opts.fs.Args()

	if opts.color != "" {
		if !(opts.color == program || opts.color == system || opts.color == all || opts.color == off) {
			return nil, errColorFlag
		}
	}

	if minTime != "" {
		result, err := when.EN.Parse(minTime, now)
		if err != nil || result == nil {
			return nil, errors.Join(errMinTimeFlag, err)
		}

		opts.minTime = result.Time.Format(time.RFC3339)
	}

	if maxTime != "" {
		result, err := when.EN.Parse(maxTime, now)
		if err != nil || result == nil {
			return nil, errors.Join(errMaxTimeFlag, err)
		}

		opts.maxTime = result.Time.Format(time.RFC3339)
	}

	if content, err := os.ReadFile(opts.configFile); err == nil {
		err = yaml.Unmarshal(content, opts)
		if err != nil {
			return nil, fmt.Errorf("error while unmarshaling %s config file: %w", opts.configFile, err)
		}
	}

	if token := os.Getenv("SWOKEN"); token != "" {
		opts.Token = token
	}

	return opts, nil
}
