package immodels

import (
	"IM/pkg/constants"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChatLog struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id, omitempty"` // omitempty表示该字段的值为“零值”，那么在生成时，该字段将被完全省略

	ConversationId string             `bson:"conversationId"`
	SendId         string             `bson:"sendId"`
	RecvId         string             `bson:"recvId"`
	MsgFrom        int                `bson:"msgFrom"`
	ChatType       constants.ChatType `bson:"chatType"`
	MsgType        constants.MType    `bson:"msgType"`
	MsgContent     string             `bson:"msgContent"`
	SendTime       int64              `bson:"sendTime"`
	Status         int                `bson:"status"`

	UpdateAt time.Time `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
	CreateAt time.Time `bson:"createAt,omitempty" json:"createAt,omitempty"`
}
