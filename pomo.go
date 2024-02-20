package pomo

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

const (
	Duration    = "25m"
	Break       = "5m"
	LongBreak   = "15m"
	Warn        = "1m"
	Prefix      = "🍅"
	PrefixBreak = "🧘"
	PrefixWarn  = "💢"

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
		now := time.Now()
		endtime := now.Add(dur).Format(time.RFC3339)

		session := Session{}

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
					// Do nothing
					return nil
				}
			} else {
				// Save the current session with the end time at the moment
				if err := session.Stop(); err != nil {
					return err
				}
			}
		}

		if err := startSession(now, endtime, WorkSession); err != nil {
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

				now := time.Now()
				endtime := now.Add(dur).Format(time.RFC3339)

				session := Session{}

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
							// Do nothing
							return nil
						}
					} else {
						// Save the current session with the end time at the moment
						if err := session.Stop(); err != nil {
							return err
						}
					}
				}

				if err := startSession(now, endtime, BreakSession); err != nil {
					return err
				}

				return nil
			},
		},
		{
			Name:  "stop",
			Usage: "stop the pomodoro countdown",
			Action: func(_ *cli.Context) error {
				session := Session{}

				if err := session.Current(); err != nil {
					return err
				}

				if !session.isRunning() {
					return fmt.Errorf("no session is running")
				}

				if err := session.Stop(); err != nil {
					return err
				}

				if err := conf.Del("endtime"); err != nil {
					return err
				}

				return nil
			},
		},
		{
			Name:  "print",
			Usage: "print current to standard output",
			Action: func(_ *cli.Context) error {
				endtime := conf.Query("endtime")
				if endtime == "" {
					fmt.Print("No session...")
					return nil
				}

				endt, err := time.Parse(time.RFC3339, endtime)
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
		{
			Name:  "status",
			Usage: "Displays an summary of all the sessions that was acomplished today",
			Action: func(_ *cli.Context) error {
				sessions, err := ListSessions()
				if err != nil {
					return err
				}

				todaySessions, err := filterTodaySessions(sessions)
				if err != nil {
					return err
				}

				typeDurations, _ := summarizeSessions(todaySessions)

				workDuration := typeDurations[WorkSession]
				breakDuration := typeDurations[BreakSession]

				green := color.New(color.FgGreen).SprintFunc()
				blue := color.New(color.FgBlue).SprintFunc()
				yellow := color.New(color.FgYellow).SprintFunc()

				workGoalPercent := float64(workDuration) / float64(WorkGoal) * 100

				fmt.Print("\n=== Sessions Today ===\n\n")
				fmt.Printf("Work  -> %s\n", green(formatDurationHm(workDuration)))
				fmt.Printf("Break -> %s\n", blue(formatDurationHm(breakDuration)))

				fmt.Print("\n=== Goals ===\n\n")
				fmt.Printf("Work: %s  =>  %s%%\n", yellow(formatDurationHm(WorkGoal)), yellow(fmt.Sprintf("%.2f", workGoalPercent)))
				fmt.Printf("Break: %s\n\n", yellow(formatDurationHm(RestGoal)))

				return nil
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

func startSession(start time.Time, end string, mode SessionType) error {
	newSession := Session{
		StartTime: start.Format(time.RFC3339),
		Type:      mode,
	}

	if err := newSession.Start(); err != nil {
		return err
	}

	if err := conf.Set("endtime", end); err != nil {
		return err
	}

	return nil
}

func filterTodaySessions(sessions []Session) ([]Session, error) {
	var todaySessions []Session
	now := time.Now()
	today := now.Format("2006-01-02") // YYYY-MM-DD format

	for _, session := range sessions {
		startTime, err := time.Parse(time.RFC3339, session.StartTime)
		if err != nil {
			return nil, fmt.Errorf("parsing session start time: %v", err)
		}

		if startTime.Format("2006-01-02") == today {
			todaySessions = append(todaySessions, session)
		}
	}

	return todaySessions, nil
}

func summarizeSessions(sessions []Session) (map[SessionType]time.Duration, map[string]time.Duration) {
	typeDurations := make(map[SessionType]time.Duration)
	projectDurations := make(map[string]time.Duration)

	for _, session := range sessions {
		startTime, _ := time.Parse(time.RFC3339, session.StartTime)
		var endTime time.Time
		var duration time.Duration

		if session.EndTime != "" {
			endTime, _ = time.Parse(time.RFC3339, session.EndTime)
			duration = endTime.Sub(startTime)
		} else {
			// Assuming sessions without an end time are still running or error
			duration = time.Since(startTime)
		}

		typeDurations[session.Type] += duration
		projectDurations[session.Filepath] += duration
	}

	return typeDurations, projectDurations
}

func formatDurationHm(d time.Duration) string {
	hours := d / time.Hour
	minutes := (d % time.Hour) / time.Minute

	if hours == 0 {
		return fmt.Sprintf("%02dm", minutes)
	}

	return fmt.Sprintf("%dh:%02dm", hours, minutes)
}
