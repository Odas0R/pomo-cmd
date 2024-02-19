package pomo

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/urfave/cli/v2"
)

const (
	Duration    = "25m"
	Break       = "5m"
	LongBreak   = "10m"
	Warn        = "1m"
	Prefix      = "🍅"
	PrefixBreak = "🧘"
	PrefixWarn  = "💢"
)

var (
	conf = Conf{
		Id:   "pomo",
		Dir:  "/home/odas0r/github.com/odas0r/configs",
		File: "config.json",
	}
)

var App = &cli.App{
	Name:                 "pomo",
	Usage:                "A pomodoro command line interface 🍅",
	EnableBashCompletion: true,
	Action: func(cCtx *cli.Context) error {
		var arg string
		if cCtx.Args().Present() {
			arg = cCtx.Args().First()
			_, err := time.ParseDuration(arg)
			if err != nil {
				return fmt.Errorf("error: the input must be like 1m, 1h, 1s, 1h30m, etc")
			}
		}

		var duration string
		if arg != "" {
			duration = arg
		} else {
			duration = conf.Query("duration")
		}

		if duration == "" {
			duration = Duration
		}

		dur, err := time.ParseDuration(duration)
		if err != nil {
			return err
		}
		started := time.Now().Add(dur).Format(time.RFC3339)

		if err := conf.Set("started", started); err != nil {
			return err
		}

		return nil
	},
	Commands: []*cli.Command{
		{
			Name:  "break",
			Usage: "initialize a break session",
			Action: func(cCtx *cli.Context) error {
				var arg string
				if cCtx.Args().Present() {
					arg = cCtx.Args().First()
					_, err := time.ParseDuration(arg)
					if err != nil {
						return fmt.Errorf("error: the input must be like 1m, 1h, 1s, 1h30m, etc")
					}
				}

				var duration string
				if arg != "" {
					duration = arg
				} else {
					duration = conf.Query("break")
				}

				if duration == "" {
					duration = Break
				}

				dur, err := time.ParseDuration(duration)
				if err != nil {
					return err
				}
				started := time.Now().Add(dur).Format(time.RFC3339)

				if err := conf.Set("started", started); err != nil {
					return err
				}

				return nil
			},
		},
		{
			Name:  "longbreak",
			Usage: "initialize a long break session",
			Action: func(cCtx *cli.Context) error {
				var arg string
				if cCtx.Args().Present() {
					arg = cCtx.Args().First()
					_, err := time.ParseDuration(arg)
					if err != nil {
						return fmt.Errorf("error: the input must be like 1m, 1h, 1s, 1h30m, etc")
					}
				}

				var duration string
				if arg != "" {
					duration = arg
				} else {
					duration = conf.Query("long_break")
				}

				if duration == "" {
					duration = LongBreak
				}

				dur, err := time.ParseDuration(duration)
				if err != nil {
					return err
				}
				started := time.Now().Add(dur).Format(time.RFC3339)

				if err := conf.Set("started", started); err != nil {
					return err
				}

				return nil
			},
		},
		{
			Name:  "stop",
			Usage: "stop the pomodoro countdown",
			Action: func(_ *cli.Context) error {
				if err := conf.Del("started"); err != nil {
					return err
				}

				return nil
			},
		},
		{
			Name:  "print",
			Usage: "print current to standard output",
			Action: func(_ *cli.Context) error {
				started := conf.Query("started")
				if started == "" {
					fmt.Print("No session...")
					return nil
				}

				endt, err := time.Parse(time.RFC3339, started)
				if err != nil {
					return err
				}

				var subt time.Duration
				subc := conf.Query("interval")

				if subc != "" {
					subt, err = time.ParseDuration(subc)
					if err != nil {
						return err
					}
				}

				prefix := conf.Query("prefix")
				prefixWarn := conf.Query("prefix_warn")
				warn := conf.Query("warn")

				warnt, err := time.ParseDuration(warn)
				if err != nil {
					return err
				}

				sec := time.Second
				left := time.Until(endt).Round(sec)

				var sub float64
				if subc != "" {
					sub = math.Abs(math.Mod(left.Seconds(), subt.Seconds()))
				}

				if left < warnt && left%(sec*2) == 0 {
					// alternate the prefix
					prefix = prefixWarn
				}

				if subc != "" {
					fmt.Printf("%v %v(%02v)", prefix, StopWatchFormat(left), sub)
				} else {
					fmt.Printf("%v %v", prefix, StopWatchFormat(left))
				}

				return nil
			},
		},
		{
			Name: "init",
			Action: func(_ *cli.Context) error {
				err := conf.Init()
				if err != nil {
					log.Fatal(err)
				}

				conf.Set("duration", Duration)
				conf.Set("break", Break)
				conf.Set("long_break", LongBreak)
				conf.Set("warn", Warn)
				conf.Set("prefix", Prefix)
				conf.Set("prefix_warn", PrefixWarn)

				return nil
			},
		},
		{
			Name: "config",
			Action: func(_ *cli.Context) error {
				if err := conf.Print(); err != nil {
					return err
				}
				return nil

			},
			Subcommands: []*cli.Command{
				{
					Name:  "edit",
					Usage: "Opens the editor on the current config",
					Action: func(_ *cli.Context) error {
						return conf.Edit()
					},
				},
			},
		},
	},
}
