package reminderrepo

import (
	"context"
	"errors"

	"VKInternshipChatBot/internal/entities/reminder"
)

var (
	ReminderNotFoundErr = errors.New("reminder not found")
)

type RepositoryMap struct {
	data map[int64]reminder.Reminder
	size int64
}

func (rm *RepositoryMap) AddReminder(ctx context.Context, reminder reminder.Reminder) (int64, error) {
	id := rm.size
	reminder.ID = id
	rm.data[id] = reminder
	rm.size++
	return id, nil
}

func (rm *RepositoryMap) GetReminderById(ctx context.Context, ID int64) (*reminder.Reminder, error) {
	if v, ok := rm.data[ID]; ok {
		return &v, nil
	}
	return nil, ReminderNotFoundErr
}

func (rm *RepositoryMap) UpdateReminderById(ctx context.Context, ID int64, reminder reminder.Reminder) error {
	rm.data[ID] = reminder
	return nil
}

func (rm *RepositoryMap) DeleteReminderById(ctx context.Context, ID int64) error {
	delete(rm.data, ID)
	return nil
}

func (rm *RepositoryMap) GetAllReminders(ctx context.Context) (*[]reminder.Reminder, error) {
	reminders := make([]reminder.Reminder, 0)

	for _, v := range rm.data {
		reminders = append(reminders, v)
	}

	return &reminders, nil
}

func New() *RepositoryMap {
	return &RepositoryMap{data: make(map[int64]reminder.Reminder)}
}
