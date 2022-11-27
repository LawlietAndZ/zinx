package ziface

import "net"

type IConnection interface {
	// Start 启动连接
	Start()

	// Stop 停止连接
	Stop()

	// GetTCPConnection  获取当前连接绑定的socket
	GetTCPConnection() *net.TCPConn

	// GetConnID 获取当前连接模块的链接id
	GetConnID() uint32

	// GetRemote 获取远程客户端的 TCP状态 IP port
	GetRemote() net.Addr

	// SendMsg 直接将Message数据发送数据给远程的TCP客户端(无缓冲)
	SendMsg(msgId uint32, data []byte) error
}

// HandleFunc 定义处理业务的方法
type HandleFunc func(*net.TCPConn, []byte, int) error
