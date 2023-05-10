package ui

type MenuType int

const (
	MainMenu = iota
	CreateReminderMenu
	ListRemindersMenu
	DeleteReminderMenu
	OnlyBackMenu
	UpdateReminderMenu
)

var MenuToKeyboard = map[MenuType]Keyboard{
	MainMenu: {
		OneTime: true,
		Buttons: [][]Button{
			{
				{Action: Action{
					Type:  "text",
					Label: CreateReminderUserMessage,
				}},
				{Action: Action{
					Type:  "text",
					Label: ListReminderUserMessage,
				}},
			},
			{
				{Action: Action{
					Type:  "text",
					Label: DeleteReminderUserMessage,
				}},
				{Action: Action{
					Type:  "text",
					Label: UpdateReminderUserMessage,
				}},
			},
		},
	},
	ListRemindersMenu: {
		OneTime: true,
		Buttons: [][]Button{
			{
				{Action: Action{
					Type:  "text",
					Label: SortByDateUserMessage,
				}},
				{Action: Action{
					Type:  "text",
					Label: BackToMainMenuUserMessage,
				}},
			},
		},
	},
	CreateReminderMenu: {
		OneTime: true,
		Buttons: [][]Button{
			{
				{Action: Action{
					Type:  "text",
					Label: EnterReminderUserMessage,
				}},
				{Action: Action{
					Type:  "text",
					Label: BackToMainMenuUserMessage,
				}},
			},
		},
	},
	DeleteReminderMenu: {
		OneTime: true,
		Buttons: [][]Button{
			{
				{Action: Action{
					Type:  "text",
					Label: DeleteManyRemindersUserMessage,
				}},
			},
			{
				{Action: Action{
					Type:  "text",
					Label: BackToMainMenuUserMessage,
				}},
			},
		},
	},
	OnlyBackMenu: {
		OneTime: true,
		Buttons: [][]Button{
			{
				{Action: Action{
					Type:  "text",
					Label: BackToMainMenuUserMessage,
				}},
			},
		},
	},
	UpdateReminderMenu: {
		OneTime: true,
		Buttons: [][]Button{
			{
				{Action: Action{
					Type:  "text",
					Label: UpdateReminderDateUserMessage,
				}},
			},
			{
				{Action: Action{
					Type:  "text",
					Label: BackToMainMenuUserMessage,
				}},
			},
		},
	},
}

type Button struct {
	Action Action `json:"action"`
}

type Action struct {
	Type  string `json:"type"`
	Label string `json:"label"`
}

type Keyboard struct {
	OneTime bool       `json:"one_time"`
	Buttons [][]Button `json:"buttons"`
}
