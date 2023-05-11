package ports

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"VKInternshipChatBot/internal/app"
	"VKInternshipChatBot/internal/entities/reminder"
	"VKInternshipChatBot/internal/ui"
)

type UserMode int

const (
	UserClicksMenuMode = iota
	UserEnterReminderMode
	UserDeletingReminderMode
	UserUpdatingReminderMode
)

type VKClient struct {
	BaseURL     string
	accessToken string
	groupID     string
	version     string
	client      http.Client
	app         app.App
	userModes   map[int]UserMode
}

func (vkc *VKClient) Run() error {
	go func() {
		for {
			vkc.notifyReminders()
			time.Sleep(10 * time.Second)
		}
	}()

	for {
		lps, err := vkc.getLongPollServer()
		if err != nil {
			return err
		}

		data, err := vkc.doRequest(
			http.MethodGet,
			lps.Server,
			map[string]string{
				"act":  "a_check",
				"key":  lps.Key,
				"ts":   lps.Ts,
				"wait": "25",
			},
			nil,
		)
		if err != nil {
			return err
		}

		var getUpdatesSuccessResponse GetUpdatesSuccessResponse
		if err := json.Unmarshal(data, &getUpdatesSuccessResponse); err != nil {
			return fmt.Errorf("failed to unmarshal body: %v", err)
		}

		vkc.handleUpdates(getUpdatesSuccessResponse.Updates)
	}
}

func (vkc *VKClient) notifyReminders() error {
	for k, _ := range vkc.userModes {
		reminders, _ := vkc.app.ListReminders(context.Background(), int64(k))
		for _, r := range *reminders {
			if r.Time.Before(time.Now()) {
				_, err := vkc.doRequest(
					http.MethodPost,
					vkc.BaseURL+"messages.send",
					map[string]string{
						"user_id":   strconv.Itoa(k),
						"random_id": strconv.Itoa(rand.Intn(math.MaxInt32)),
						"v":         vkc.version,
						"message":   ui.TimeIsLeft + "\n" + remindersToString([]reminder.Reminder{r}),
					},
					nil,
				)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (vkc *VKClient) doRequest(httpMethod string, URL string, query map[string]string, data []byte) ([]byte, error) {
	req, err := http.NewRequest(httpMethod, URL, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %v", err)
	}

	req.Header.Add("Authorization", "Bearer "+vkc.accessToken)

	q := req.URL.Query()
	for k, v := range query {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := vkc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %v", err)
	}

	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	return respBody, nil
}

func (vkc *VKClient) handleUpdates(updates []Update) {
	for _, update := range updates {
		switch update.Type {
		case "message_new":
			err := vkc.handleMessage(update.Object.Message)
			if err != nil {
				log.Printf("failed to handle message: %v\n", update.Object.Message)
			}
		default:
		}
	}
}

func (vkc *VKClient) handleMessage(message Message) error {
	if vkc.userModes[message.FromID] == UserEnterReminderMode {
		if message.Text == ui.BackToMainMenuUserMessage {
			vkc.userModes[message.FromID] = UserClicksMenuMode
			return vkc.showMenu(ui.MainMenu, message.FromID, ui.BackToMainMenuBotMessage)
		}
		parts := strings.Split(message.Text, ";")
		timeParsed, err := time.Parse(time.DateTime, parts[0])
		if err != nil {
			return err
		}
		_, err = vkc.app.AddReminder(context.Background(), parts[1], timeParsed, int64(message.FromID))
		if err != nil {
			return err
		}
		vkc.userModes[message.FromID] = UserClicksMenuMode
		return vkc.showMenu(ui.MainMenu, message.FromID, ui.ReminderSuccessfullyAddedBotMessage+ui.BackToMainMenuBotMessage)
	}

	if vkc.userModes[message.FromID] == UserDeletingReminderMode {
		if message.Text == ui.DeleteManyRemindersUserMessage {
			return vkc.showMenu(ui.OnlyBackMenu, message.FromID, ui.EnterReminderIDsBotMessage)
		}

		if message.Text == ui.BackToMainMenuUserMessage {
			vkc.userModes[message.FromID] = UserClicksMenuMode
			return vkc.showMenu(ui.MainMenu, message.FromID, ui.BackToMainMenuBotMessage)
		}
		parts := strings.Split(message.Text, ",")
		for _, p := range parts {
			idx, err := strconv.Atoi(p)
			if err != nil {
				return vkc.showMenu(ui.OnlyBackMenu, message.FromID, ui.WrongFormatBotMessage)
			}
			vkc.app.DeleteReminder(context.Background(), int64(idx))
		}
		vkc.userModes[message.FromID] = UserClicksMenuMode
		return vkc.showMenu(ui.MainMenu, message.FromID, ui.ReminderSuccessfullyDeletedBotMessage)
	}

	if vkc.userModes[message.FromID] == UserUpdatingReminderMode {
		parts := strings.Split(message.Text, ";")
		timeParsed, err := time.Parse(time.DateTime, parts[1])
		if err != nil {
			return vkc.showMenu(ui.OnlyBackMenu, message.FromID, ui.WrongFormatBotMessage)
		}
		idx, err := strconv.Atoi(parts[0])
		if err != nil {
			return vkc.showMenu(ui.OnlyBackMenu, message.FromID, ui.WrongFormatBotMessage)
		}

		vkc.app.UpdateReminder(context.Background(), int64(idx), timeParsed, int64(message.FromID))
		vkc.userModes[message.FromID] = UserClicksMenuMode
		return nil
	}

	var err error

	switch message.Text {
	case ui.StartUserMessage:
		vkc.userModes[message.FromID] = UserClicksMenuMode
		err = vkc.showMenu(ui.MainMenu, message.FromID, ui.StartBotMessage)
	case ui.CreateReminderUserMessage:
		err = vkc.showMenu(ui.CreateReminderMenu, message.FromID, ui.ChooseActionBotMessage)
	case ui.ListReminderUserMessage:
		reminders, _ := vkc.app.ListReminders(context.Background(), int64(message.FromID))
		if len(*reminders) == 0 {
			err = vkc.showMenu(ui.ListRemindersMenu, message.FromID, ui.DontHaveRemindersBotMessage)
		} else {
			err = vkc.showMenu(ui.ListRemindersMenu, message.FromID, remindersToString(*reminders))
		}
	case ui.SortRemindersByDateUserMessage:
		reminders, _ := vkc.app.ListSortedByDateReminders(context.Background(), int64(message.FromID))
		if len(*reminders) == 0 {
			err = vkc.showMenu(ui.ListRemindersMenu, message.FromID, ui.DontHaveRemindersBotMessage)
		} else {
			err = vkc.showMenu(ui.ListRemindersMenu, message.FromID, remindersToString(*reminders))
		}
	case ui.DeleteReminderUserMessage:
		vkc.userModes[message.FromID] = UserDeletingReminderMode
		reminders, _ := vkc.app.ListReminders(context.Background(), int64(message.FromID))
		if len(*reminders) == 0 {
			err = vkc.showMenu(ui.DeleteReminderMenu, message.FromID, ui.DontHaveRemindersBotMessage)
		} else {
			err = vkc.showMenu(ui.DeleteReminderMenu, message.FromID, remindersToIndexed(*reminders)+ui.EnterReminderIDBotMessage)
		}
	case ui.BackToMainMenuUserMessage:
		err = vkc.showMenu(ui.MainMenu, message.FromID, ui.BackToMainMenuBotMessage)
	case ui.EnterReminderDateAndTextUserMessage:
		vkc.userModes[message.FromID] = UserEnterReminderMode
		err = vkc.showMenu(ui.OnlyBackMenu, message.FromID, ui.EnterReminderDateAndTextBotMessage)
	case ui.DeleteManyRemindersUserMessage:
		vkc.userModes[message.FromID] = UserDeletingReminderMode
		err = vkc.showMenu(ui.OnlyBackMenu, message.FromID, ui.EnterReminderIDsBotMessage)
	case ui.UpdateReminderUserMessage:
		err = vkc.showMenu(ui.UpdateReminderMenu, message.FromID, ui.ChooseActionBotMessage)
	case ui.UpdateReminderDateUserMessage:
		vkc.userModes[message.FromID] = UserUpdatingReminderMode
		reminders, _ := vkc.app.ListReminders(context.Background(), int64(message.FromID))
		if len(*reminders) == 0 {
			err = vkc.showMenu(ui.OnlyBackMenu, message.FromID, ui.DontHaveRemindersBotMessage)
			vkc.userModes[message.FromID] = UserClicksMenuMode
		} else {
			err = vkc.showMenu(ui.OnlyBackMenu, message.FromID, remindersToIndexed(*reminders)+ui.EnterReminderIdUpdatingBotMessage)
		}
	}
	if err != nil {
		return err
	}
	return nil
}

func remindersToString(r []reminder.Reminder) string {
	s := ""
	for _, v := range r {
		s += fmt.Sprintf("Время:\n%s\nТекст:%s\n\n", v.Time.Format(time.DateTime), v.Text)
	}
	return s
}

func remindersToIndexed(r []reminder.Reminder) string {
	s := ""
	for i, v := range r {
		s += fmt.Sprintf("%d:\nВремя:\n%s\nТекст:%s\n\n", i, v.Time.Format(time.DateTime), v.Text)
	}
	return s
}

func (vkc *VKClient) showMenu(menuType ui.MenuType, userID int, message string) error {
	keyboard := ui.MenuToKeyboard[menuType]

	keyboardData, _ := json.Marshal(keyboard)

	_, err := vkc.doRequest(
		http.MethodPost,
		vkc.BaseURL+"messages.send",
		map[string]string{
			"user_id":   strconv.Itoa(userID),
			"random_id": strconv.Itoa(rand.Intn(math.MaxInt32)),
			"v":         vkc.version,
			"keyboard":  string(keyboardData),
			"message":   message,
		},
		nil,
	)

	if err != nil {
		return err
	}

	return nil
}

func (vkc *VKClient) getLongPollServer() (*LongPollServer, error) {
	data, err := vkc.doRequest(
		http.MethodGet,
		vkc.BaseURL+"groups.getLongPollServer",
		map[string]string{
			"access_token": vkc.accessToken,
			"group_id":     vkc.groupID,
			"v":            vkc.version,
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	var longPollServerResponse LongPollServerResponse

	if err := json.Unmarshal(data, &longPollServerResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal body: %v", err)
	}

	return &LongPollServer{
		Key:    longPollServerResponse.Data.Key,
		Server: longPollServerResponse.Data.Server,
		Ts:     longPollServerResponse.Data.Ts,
	}, nil
}

func NewClient(baseURL string, accessToken string, app app.App, groupID string, version string) *VKClient {
	return &VKClient{
		BaseURL:     baseURL,
		app:         app,
		accessToken: accessToken,
		client:      http.Client{},
		groupID:     groupID,
		version:     version,
		userModes:   map[int]UserMode{},
	}
}
