package immodels

import "github.com/zeromicro/go-zero/core/stores/mon"

var _ ChatLogModel = (*customChatLogModel)(nil)

type (
	// ChatLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customChatLogModel.
	ChatLogModel interface {
		chatLogModel
	}

	customChatLogModel struct {
		*defaultChatLogModel
	}
)

// NewChatLogModel returns a model for the mongo.
// 数据库连接URL、数据库名和集合（或表）名
func NewChatLogModel(url, db, collection string) ChatLogModel {
	conn := mon.MustNewModel(url, db, collection)
	return &customChatLogModel{
		defaultChatLogModel: newDefaultChatLogModel(conn),
	}
}

func MustChatLogModel(url, db string) ChatLogModel {
	return NewChatLogModel(url, db, "chat_log")
}
