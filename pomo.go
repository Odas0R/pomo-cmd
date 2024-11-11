package pomo

import (
	"fmt"
	"log"
	"time"

	"github.com/urfave/cli/v2"
)

const (
	Duration    = "25m"
	Break       = "5m"
	LongBreak   = "15m"
	Warn        = "1m"
	WorkPrefix  = "🍅"
	BreakPrefix = "🧘"
	WarnPrefix  = "💢"

	// Goals
	WorkGoal = 8*time.Hour + 20*time.Minute
	RestGoal = 1*time.Hour + 40*time.Minute
)

var (
	conf = Conf{
		Id:   "pomo",
		Dir:  "/home/odas0r/github.com/odas0r/configs",
		File: "config.json",
	}
)

func init() {
	loc, err := time.LoadLocation("Europe/Lisbon")
	if err != nil {
		log.Fatalf("Failed to load Lisbon timezone: %v", err)
	}
	time.Local = loc
}

var App = &cli.App{
	Name:                 "pomo",
	Usage:                "A pomodoro command line interface 🍅",
	EnableBashCompletion: true,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "ui",
			Aliases: []string{"u"},
			Usage:   "display interactive terminal UI",
		},
	},
	Action: func(cCtx *cli.Context) error {
		var arg string
		if cCtx.Args().Present() {
			arg = cCtx.Args().First()
			_, err := time.ParseDuration(arg)
			if err != nil {
				return fmt.Errorf("error: the input must be like 1m, 1h, 1s, 1h30m, etc")
			}
		}

		var durationStr string
		if arg != "" {
			durationStr = arg
		} else {
			durationStr = conf.QueryString("duration")
		}
		if durationStr == "" {
			durationStr = Duration
		}

		duration, err := time.ParseDuration(durationStr)
		if err != nil {
			return err
		}

		var session Session
		if err := session.Current(); err != nil {
			return err
		}

		if session.isRunning() {
			if session.Type == WorkSession {
				ok := InputConfirm("[WARNING]: A session is already running, do you want to reset it?")
				if ok {
					if err := session.Remove(); err != nil {
						return err
					}
				} else {
					if cCtx.Bool("ui") {
						return StartUI()
					}
					return nil
				}
			} else {
				// Save the current session with the end time at the moment
				if err := session.Stop(); err != nil {
					return err
				}
			}
		}

		if err := session.Start(conf, duration, WorkSession); err != nil {
			return err
		}

		if err := conf.Set("prefix", WorkPrefix); err != nil {
			return err
		}

		if cCtx.Bool("ui") {
			return StartUI()
		}

		return nil
	},
	Commands: []*cli.Command{
		{
			Name:  "break",
			Usage: "initialize a break session",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "ui",
					Aliases: []string{"u"},
					Usage:   "display interactive terminal UI",
				},
			},
			Action: func(cCtx *cli.Context) error {
				var arg string
				if cCtx.Args().Present() {
					arg = cCtx.Args().First()
					_, err := time.ParseDuration(arg)
					if err != nil {
						return fmt.Errorf("error: the input must be like 1m, 1h, 1s, 1h30m, etc")
					}
				}

				var durationStr string
				if arg != "" {
					durationStr = arg
				} else {
					durationStr = conf.QueryString("break")
				}

				if durationStr == "" {
					durationStr = Break
				}

				duration, err := time.ParseDuration(durationStr)
				if err != nil {
					return err
				}

				var session Session
				if err := session.Current(); err != nil {
					return err
				}

				if session.isRunning() {
					if session.Type == BreakSession {
						ok := InputConfirm("[WARNING]: A session is already running, do you want to reset it?")
						if ok {
							if err := session.Remove(); err != nil {
								return err
							}
						} else {
							if cCtx.Bool("ui") {
								return StartUI()
							}
							return nil
						}
					} else {
						if err := session.Stop(); err != nil {
							return err
						}
					}
				}

				if err := session.Start(conf, duration, WorkSession); err != nil {
					return err
				}

				if err := conf.Set("prefix", BreakPrefix); err != nil {
					return err
				}

				if cCtx.Bool("ui") {
					return StartUI()
				}

				return nil
			},
		},
		{
			Name:  "stop",
			Usage: "stop the pomodoro countdown",
			Action: func(_ *cli.Context) error {
				var session Session
				if err := session.Current(); err != nil {
					return err
				}

				if !session.isRunning() {
					return fmt.Errorf("no session is running")
				}

				if err := session.Stop(); err != nil {
					return err
				}

				return nil
			},
		},
		{
			Name:  "print",
			Usage: "print current to standard output",
			Action: func(_ *cli.Context) error {
				var session Session
				if err := session.Current(); err != nil {
					return err
				}

				if !session.isRunning() {
					return nil
				}

				// Get elapsed time (remaining time)
				remaining := session.Elapsed()

				// Get configuration values
				prefix := conf.QueryString("prefix")          // 🍅
				prefixWarn := conf.QueryString("prefix_warn") // 💢
				warn := conf.QueryString("warn")              // 1m

				// Parse warning threshold
				warnTime, err := time.ParseDuration(warn)
				if err != nil {
					return err
				}

				var timeStr string
				if remaining < 0 {
					// For negative time, remove the minus and add a "-" prefix to the formatted time
					timeStr = "-" + StopWatchFormat(-remaining)
				} else {
					timeStr = StopWatchFormat(remaining)
					// Switch to warning prefix when less than 1 minute remains
					// and blink every 2 seconds
					if remaining < warnTime && remaining%(time.Second*2) == 0 {
						prefix = prefixWarn
					}
				}

				fmt.Printf("%v %v", prefix, timeStr)

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
				conf.Set("prefix", WorkPrefix)
				conf.Set("prefix_warn", WarnPrefix)

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
		{
			Name:  "status",
			Usage: "Displays an summary of all the sessions that was acomplished today",
			Action: func(_ *cli.Context) error {
				return ShowStatus()
			},
		},
		{
			Name: "sessions",
			Subcommands: []*cli.Command{
				{
					Name:  "edit",
					Usage: "Opens the editor on the current sessions file",
					Action: func(_ *cli.Context) error {
						path, err := sessionPath()
						if err != nil {
							return err
						}
						return Editor(path)
					},
				},
			},
		},
	},
}
