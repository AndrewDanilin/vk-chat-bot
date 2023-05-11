package ports

type Update struct {
	Type    string `json:"type"`
	EventID string `json:"event_id"`
	Version string `json:"v"`
	Object  Object `json:"object"`
}

type Object struct {
	Message Message
}

type Message struct {
	ID     int64  `json:"id"`
	Date   int64  `json:"date"`
	PeerID int64  `json:"peer_id"`
	FromID int    `json:"from_id"`
	Text   string `json:"text"`
}

type GetUpdatesSuccessResponse struct {
	Ts      string   `json:"ts"`
	Updates []Update `json:"updates"`
	GroupID int64    `json:"group_id"`
}

type LongPollServerResponse struct {
	Data LongPollServerData `json:"response"`
}

type LongPollServerData struct {
	Key    string
	Server string
	Ts     string
}
