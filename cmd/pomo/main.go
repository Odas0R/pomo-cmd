package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"time"

	c "github.com/odas0r/pomo-cmd/pkg/config"
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
	conf = c.Conf{
		Id:   "pomo",
		Dir:  "/home/odas0r/github.com/odas0r/configs",
		File: "config.json",
	}
)

func main() {
	app := &cli.App{
		Name:                 "pomo",
		Usage:                "A pomodoro command line interface 🍅",
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name:    "start",
				Aliases: []string{"s"},
				Usage:   "start the pomodoro countdown 🕒",
				Action: func(cCtx *cli.Context) error {
					if cCtx.Args().Present() {
						duration := cCtx.Args().First()

						_, err := time.ParseDuration(duration)
						if err != nil {
							return fmt.Errorf("error: the input must be like 1m, 1h, 1s, 1h30m, etc")
						}

						if err := conf.Set("duration", duration); err != nil {
							return err
						}
					}

					s := conf.Query("duration")
					if s == "" {
						s = Duration
					}
					dur, err := time.ParseDuration(s)
					if err != nil {
						return err
					}
					started := time.Now().Add(dur).Format(time.RFC3339)

					if err := conf.Set("started", started); err != nil {
						return err
					}
					if err := conf.Set("prefix", Prefix); err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:    "break",
				Aliases: []string{"b"},
				Usage:   "initialize a break session",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "long"},
				},
				Action: func(cCtx *cli.Context) error {
					if cCtx.Args().Present() {
						duration := cCtx.Args().First()

						_, err := time.ParseDuration(duration)
						if err != nil {
							return fmt.Errorf("error: the input must be like 1m, 1h, 1s, 1h30m, etc")
						}

						if err := conf.Set("break", duration); err != nil {
							return err
						}
					}

					var s string
					if cCtx.Bool("long") {
						s = conf.Query("break_long")
					} else {
						s = conf.Query("break")
					}

					if s == "" {
						s = Break
					}

					dur, err := time.ParseDuration(s)
					if err != nil {
						return err
					}
					started := time.Now().Add(dur).Format(time.RFC3339)

					conf.Set("prefix", PrefixBreak)

					if err := conf.Set("started", started); err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:    "stop",
				Aliases: []string{"so"},
				Usage:   "stop the pomodoro countdown",
				Action: func(_ *cli.Context) error {
					if err := conf.Del("started"); err != nil {
						return err
					}

					return nil
				},
			},

			{
				Name:    "print",
				Aliases: []string{"so"},
				Usage:   "print current to standard output",
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
				Name:  "init",
				Usage: "Initializes the pomodoro config with the default values",
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
				Name:    "set",
				Aliases: []string{"s"},
				Usage:   "Update the current config",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "duration"},
					&cli.StringFlag{Name: "break"},
					&cli.StringFlag{Name: "long-break"},
					&cli.StringFlag{Name: "prefix"},
					&cli.StringFlag{Name: "prefix-warn"},
					&cli.StringFlag{Name: "warn"},
				},
				Action: func(cCtx *cli.Context) error {
					duration := cCtx.String("duration")
					shortBreak := cCtx.String("break")
					longBreak := cCtx.String("long-break")
					prefix := cCtx.String("prefix")
					prefixWarn := cCtx.String("prefix-warn")
					warn := cCtx.String("warn")

					if duration != "" {
						conf.Set("duration", duration)
					} else if shortBreak != "" {
						conf.Set("break", shortBreak)
					} else if longBreak != "" {
						conf.Set("long_break", longBreak)
					} else if prefix != "" {
						conf.Set("prefix", prefix)
					} else if prefixWarn != "" {
						conf.Set("prefix_warn", prefixWarn)
					} else if warn != "" {
						conf.Set("warn", warn)
					}

					return nil
				},
			},
			{
				Name: "config",
				Action: func(_ *cli.Context) error {
					return conf.Print()
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

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
