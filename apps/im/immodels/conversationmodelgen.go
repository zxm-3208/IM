// Code generated by goctl. DO NOT EDIT.
package immodels

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type conversationModel interface {
	Insert(ctx context.Context, data *Conversation) error
	FindOne(ctx context.Context, id string) (*Conversation, error)
	Update(ctx context.Context, data *Conversation) error
	Delete(ctx context.Context, id string) (int64, error)
	ListByConversationIds(ctx context.Context, ids []string) ([]*Conversation, error)
	UpdateMsg(ctx context.Context, chatLog *ChatLog) error
}

type defaultConversationModel struct {
	conn *mon.Model
}

func newDefaultConversationModel(conn *mon.Model) *defaultConversationModel {
	return &defaultConversationModel{conn: conn}
}

func (m *defaultConversationModel) Insert(ctx context.Context, data *Conversation) error {
	if data.ID.IsZero() {
		data.ID = primitive.NewObjectID()
		data.CreateAt = time.Now()
		data.UpdateAt = time.Now()
	}

	_, err := m.conn.InsertOne(ctx, data)
	return err
}

func (m *defaultConversationModel) FindOne(ctx context.Context, id string) (*Conversation, error) {
	var data Conversation
	err := m.conn.FindOne(ctx, &data, bson.M{"conversationid": id})
	switch err {
	case nil:
		return &data, nil
	case mon.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultConversationModel) Update(ctx context.Context, data *Conversation) error {
	data.UpdateAt = time.Now()

	_, err := m.conn.ReplaceOne(ctx, bson.M{"_id": data.ID}, data)
	return err
}

func (m *defaultConversationModel) Delete(ctx context.Context, id string) (int64, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return 0, ErrInvalidObjectId
	}

	res, err := m.conn.DeleteOne(ctx, bson.M{"_id": oid})
	return res, err
}

func (m *defaultConversationModel) ListByConversationIds(ctx context.Context, ids []string) ([]*Conversation, error) {
	var data []*Conversation
	err := m.conn.Find(ctx, &data, bson.M{
		"conversationId": bson.M{
			"$in": ids,
		},
	})
	switch err {
	case nil:
		return data, nil
	case mon.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultConversationModel) UpdateMsg(ctx context.Context, chatLog *ChatLog) error {
	_, err := m.conn.UpdateOne(ctx,
		bson.M{"conversationId": chatLog.ConversationId},
		bson.M{
			// 收到一条消息，更新会话的总消息数，以及最后的的消息
			"$inc": bson.M{"total": 1},
			"$set": bson.M{"msg": chatLog},
		},
	)
	return err
}