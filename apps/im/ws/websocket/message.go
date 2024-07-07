package websocket

type FrameType uint8

const (
	FrameData FrameType = 0x0
	FramePing FrameType = 0x1
	FrameErr  FrameType = 0x2
)

type Message struct {
	Type   FrameType   `json:"frameType"`
	Method string      `json:"method,omitempty"`
	UserId string      `json:"userId,omitempty"`
	FromId string      `json:"fromId,omitempty"`
	Data   interface{} `json:"data,omitempty"`
}

func NewMessage(fid string, data interface{}) *Message {
	return &Message{
		Type:   FrameData,
		FromId: fid,
		Data:   data,
	}
}

func NewErrorMessage(err error) *Message {
	return &Message{
		Type: FrameErr,
		Data: err.Error(),
	}
}
