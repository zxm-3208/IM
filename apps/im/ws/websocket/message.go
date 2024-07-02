package websocket

type Message struct {
	Method string      `json:"method,omitempty"`
	UserId string      `json:"userId,omitempty"`
	FromId string      `json:"fromId,omitempty"`
	Data   interface{} `json:"data,omitempty"`
}

func NewMessage(fid string, data interface{}) *Message {
	return &Message{
		FromId: fid,
		Data:   data,
	}
}
