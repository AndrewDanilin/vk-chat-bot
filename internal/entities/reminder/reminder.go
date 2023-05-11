package reminder

import "time"

type Reminder struct {
	ID     int64
	UserID int64
	Text   string
	Time   time.Time
}

func New(text string, time time.Time, userID int64) Reminder {
	return Reminder{Text: text, Time: time, UserID: userID}
}
