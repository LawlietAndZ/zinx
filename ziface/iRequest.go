package ziface

// IRequest 把客户端请求的数据包装到request中
type IRequest interface {
	GetConnection() IConnection //获取请求连接信息
	GetData() []byte            //获取请求消息的数据
	GetMsgID() uint32           //获取请求的消息ID
}
