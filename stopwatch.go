package pomo

import (
	"fmt"
	"math"
	"time"
)

func StopWatchFormat(dur time.Duration) string {
	var out string

	sec := dur.Seconds()
	if sec < 0 {
		out += "-"
	}
	sec = math.Abs(sec)

	if sec >= 3600 {
		hours := sec / 3600
		sec = math.Mod(sec, 3600)
		out += fmt.Sprintf("%v:", int(hours))
	}

	if sec >= 60 {
		var form string
		mins := sec / 60
		sec = math.Mod(sec, 60)
		if len(out) == 0 {
			form = `%v:`
		} else {
			form = `%02v:`
		}
		out += fmt.Sprintf(form, int(mins))
	}

	var form string
	if len(out) == 0 {
		form = `%02v`
	} else {
		form = `%02v`
	}
	out += fmt.Sprintf(form, int(sec))

	return out
}
