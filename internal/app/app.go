package app

import (
	"context"
	"sort"
	"time"

	"VKInternshipChatBot/internal/entities/reminder"
)

type App interface {
	AddReminder(ctx context.Context, text string, time time.Time) (*reminder.Reminder, error)
	UpdateReminder(ctx context.Context, ID int64, time time.Time) (*reminder.Reminder, error)
	DeleteReminder(ctx context.Context, ID int64) error
	ListReminders(ctx context.Context) (*[]reminder.Reminder, error)
	ListSortedByDateReminders(ctx context.Context) (*[]reminder.Reminder, error)
}

type ReminderApp struct {
	repository Repository
}

func (ra ReminderApp) AddReminder(ctx context.Context, text string, time time.Time) (*reminder.Reminder, error) {
	r := reminder.New(text, time)

	id, err := ra.repository.AddReminder(ctx, r)
	if err != nil {
		// todo
	}

	r.ID = id

	return &r, nil
}

func (ra ReminderApp) UpdateReminder(ctx context.Context, ID int64, time time.Time) (*reminder.Reminder, error) {
	r, err := ra.repository.GetReminderById(ctx, ID)
	if err != nil {
		return nil, err
	}

	r.Time = time

	_ = ra.repository.UpdateReminderById(ctx, ID, *r)

	return r, err
}

func (ra ReminderApp) ListReminders(ctx context.Context) (*[]reminder.Reminder, error) {
	reminders, err := ra.repository.GetAllReminders(ctx)
	if err != nil {
		return nil, err
	}
	return reminders, nil
}

func (ra ReminderApp) ListSortedByDateReminders(ctx context.Context) (*[]reminder.Reminder, error) {
	reminders, err := ra.repository.GetAllReminders(ctx)
	if err != nil {
		return nil, err
	}

	sortedReminders := make([]reminder.Reminder, len(*reminders))
	copy(sortedReminders, *reminders)

	sort.Slice(sortedReminders, func(i, j int) bool {
		return sortedReminders[i].Time.Before(sortedReminders[j].Time)
	})

	return &sortedReminders, nil
}

func (ra ReminderApp) DeleteReminder(ctx context.Context, ID int64) error {
	_ = ra.repository.DeleteReminderById(ctx, ID)
	return nil
}

type Repository interface {
	AddReminder(ctx context.Context, reminder reminder.Reminder) (int64, error)
	UpdateReminderById(ctx context.Context, ID int64, reminder reminder.Reminder) error
	DeleteReminderById(ctx context.Context, ID int64) error
	GetReminderById(ctx context.Context, ID int64) (*reminder.Reminder, error)
	GetAllReminders(ctx context.Context) (*[]reminder.Reminder, error)
}

func NewApp(repo Repository) App {
	return ReminderApp{repository: repo}
}
