package reminder

import "time"

type Reminder struct {
	ID   int64
	Text string
	Time time.Time
}

func New(text string, time time.Time) Reminder {
	return Reminder{Text: text, Time: time}
}
