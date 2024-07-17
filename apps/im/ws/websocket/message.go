package websocket

import "time"

type FrameType uint8

const (
	FrameData      FrameType = 0x0
	FramePing      FrameType = 0x1
	FrameErr       FrameType = 0x2
	FrameAck       FrameType = 0x3
	FrameNoAck     FrameType = 0x4
	FrameTranspond FrameType = 0x5
)

type Message struct {
	Type         FrameType   `json:"frameType"`
	Id           string      `json:"Id"`
	Method       string      `json:"method"`
	FromId       string      `json:"fromId"`
	Data         interface{} `json:"data"`
	AckSeq       int         `json:"ackSeq"`
	ackTime      time.Time   `json:"ackTime"`
	errCount     int         `json:"errCount"`
	TranspondUid string      `json:"transpondUid"`
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
