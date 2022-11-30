package ziface

type IServer interface {
	// Start 启动服务器
	Start()
	// Stop 停止服务器
	Stop()
	// Server 运行服务器
	Server()

	// AddRouter 路由功能：给当前服务注册一个路由业务方法，供客户端链接处理使用
	AddRouter(msgId uint32, router IRouter)
	GetConnMgr() IConnManager
}
