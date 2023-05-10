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
	mode        UserMode
}

func (vkc *VKClient) Run() error {
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

		fmt.Println(getUpdatesSuccessResponse)

		vkc.handleUpdates(getUpdatesSuccessResponse.Updates)
	}
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

	fmt.Println(string(respBody))

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
	if vkc.mode == UserEnterReminderMode {
		if message.Text == ui.BackToMainMenuUserMessage {
			vkc.mode = UserClicksMenuMode
			return vkc.showMenu(ui.MainMenu, int(message.FromID), ui.BackToMainMenuBotMessage)
		}
		parts := strings.Split(message.Text, ";")
		timeParsed, err := time.Parse(time.DateTime, parts[0])
		if err != nil {
			return err
		}
		_, err = vkc.app.AddReminder(context.Background(), parts[1], timeParsed)
		if err != nil {
			return err
		}
		vkc.mode = UserClicksMenuMode
		return vkc.showMenu(ui.MainMenu, int(message.FromID), ui.ReminderSuccessfullyAddedBotMessage+ui.BackToMainMenuBotMessage)
	}

	if vkc.mode == UserDeletingReminderMode {
		if message.Text == ui.DeleteManyRemindersUserMessage {
			return vkc.showMenu(ui.OnlyBackMenu, int(message.FromID), ui.EnterReminderIDsBotMessage)
		}

		if message.Text == ui.BackToMainMenuUserMessage {
			vkc.mode = UserClicksMenuMode
			return vkc.showMenu(ui.MainMenu, int(message.FromID), ui.BackToMainMenuBotMessage)
		}
		parts := strings.Split(message.Text, ",")
		for _, p := range parts {
			idx, err := strconv.Atoi(p)
			if err != nil {
				return vkc.showMenu(ui.OnlyBackMenu, int(message.FromID), ui.WrongFormat)
			}
			vkc.app.DeleteReminder(context.Background(), int64(idx))
		}
		vkc.mode = UserClicksMenuMode
		return vkc.showMenu(ui.MainMenu, int(message.FromID), ui.ReminderSuccessfullyDeletedBotMessage)
	}

	if vkc.mode == UserUpdatingReminderMode {
		parts := strings.Split(message.Text, ";")
		timeParsed, err := time.Parse(time.DateTime, parts[1])
		if err != nil {
			return vkc.showMenu(ui.OnlyBackMenu, int(message.FromID), ui.WrongFormat)
		}
		idx, err := strconv.Atoi(parts[0])
		if err != nil {
			return vkc.showMenu(ui.OnlyBackMenu, int(message.FromID), ui.WrongFormat)
		}

		vkc.app.UpdateReminder(context.Background(), int64(idx), timeParsed)
		vkc.mode = UserClicksMenuMode
		return nil
	}

	var err error

	switch message.Text {
	case ui.StartUserMessage:
		err = vkc.showMenu(ui.MainMenu, int(message.FromID), ui.StartBotMessage)
	case ui.CreateReminderUserMessage:
		err = vkc.showMenu(ui.CreateReminderMenu, int(message.FromID), ui.ChooseActionBotMessage)
	case ui.ListReminderUserMessage:
		reminders, _ := vkc.app.ListReminders(context.Background())
		if len(*reminders) == 0 {
			err = vkc.showMenu(ui.ListRemindersMenu, int(message.FromID), ui.DontHaveRemindersBotMessage)
		} else {
			err = vkc.showMenu(ui.ListRemindersMenu, int(message.FromID), remindersToString(*reminders))
		}
	case ui.SortByDateUserMessage:
		reminders, _ := vkc.app.ListSortedByDateReminders(context.Background())
		if len(*reminders) == 0 {
			err = vkc.showMenu(ui.ListRemindersMenu, int(message.FromID), ui.DontHaveRemindersBotMessage)
		} else {
			err = vkc.showMenu(ui.ListRemindersMenu, int(message.FromID), remindersToString(*reminders))
		}
	case ui.DeleteReminderUserMessage:
		vkc.mode = UserDeletingReminderMode
		reminders, _ := vkc.app.ListReminders(context.Background())
		if len(*reminders) == 0 {
			err = vkc.showMenu(ui.DeleteReminderMenu, int(message.FromID), ui.DontHaveRemindersBotMessage)
		} else {
			err = vkc.showMenu(ui.DeleteReminderMenu, int(message.FromID), remindersToIndexed(*reminders)+ui.InputReminderIDBotMessage)
		}
	case ui.BackToMainMenuUserMessage:
		err = vkc.showMenu(ui.MainMenu, int(message.FromID), ui.BackToMainMenuBotMessage)
	case ui.EnterReminderUserMessage:
		vkc.mode = UserEnterReminderMode
		err = vkc.showMenu(ui.OnlyBackMenu, int(message.FromID), ui.CreateReminderDateTextBotMessage)
	case ui.DeleteManyRemindersUserMessage:
		vkc.mode = UserDeletingReminderMode
		err = vkc.showMenu(ui.OnlyBackMenu, int(message.FromID), ui.EnterReminderIDsBotMessage)
	case ui.UpdateReminderUserMessage:
		err = vkc.showMenu(ui.UpdateReminderMenu, int(message.FromID), ui.ChooseActionBotMessage)
	case ui.UpdateReminderDateUserMessage:
		vkc.mode = UserUpdatingReminderMode
		reminders, _ := vkc.app.ListReminders(context.Background())
		if len(*reminders) == 0 {
			err = vkc.showMenu(ui.OnlyBackMenu, int(message.FromID), ui.DontHaveRemindersBotMessage)
			vkc.mode = UserClicksMenuMode
		} else {
			err = vkc.showMenu(ui.OnlyBackMenu, int(message.FromID), remindersToIndexed(*reminders)+ui.InputReminderIdUpdatingBotMessage)
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
	}
}
