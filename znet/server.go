package znet

import (
	"fmt"
	"net"
	"zinx/utils"
	"zinx/ziface"
)

type Server struct {
	//服务器名称
	Name string
	//服务器绑定的ip版本
	IPVersion string

	IP   string
	Port int

	////当前Server由用户绑定的回调router,也就是Server注册的链接对应的处理业务
	//Router ziface.IRouter
	//当前Server的消息管理模块，用来绑定MsgId和对应的处理方法
	MsgHandler     ziface.IMsgHandle
	WorkerPoolSize uint32

	//当前Server的链接管理器
	ConnMgr ziface.IConnManager
}

func (s *Server) AddRouter(msgId uint32, router ziface.IRouter) {
	s.MsgHandler.AddRouter(msgId, router)
	fmt.Println("Add router succ! msgId = ", msgId)
}

//func CallBackToClient(conn *net.TCPConn, buf []byte, cnt int) error {
//	fmt.Println("Conn callback to Client...")
//	if _, err := conn.Write(buf[:cnt]); err != nil {
//		fmt.Println("write back err:", err)
//		return errors.New("CallBackToClient error")
//	}
//	return nil
//}

func (s *Server) Start() {
	fmt.Println("开始监听")

	fmt.Printf("[START] Server name: %s,listenner at IP: %s, Port %d is starting\n", s.Name, s.IP, s.Port)
	fmt.Printf("[Zinx] Version: %s, MaxConn: %d,  MaxPacketSize: %d\n",
		utils.GlobalObject.Version,
		utils.GlobalObject.MaxConn,
		utils.GlobalObject.MaxPacketSize)
	go func() {
		//0 启动worker工作池机制
		s.MsgHandler.StartWorkerPool()
		//获取addr
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Println("获取tcp地址错误", err)
			return
		}

		//监听服务器地址
		listenner, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			fmt.Println("监听服务器出现错误", err)
			return
		}
		//阻塞等待客户端连接，处理连接业务

		for {
			conn, err := listenner.AcceptTCP()
			if err != nil {
				fmt.Println("连接失败：", err)
				continue
			}

			var cid uint32
			cid = 0

			if s.GetConnMgr().Len() >= utils.GlobalObject.MaxConn {
				conn.Close()
				continue
			}

			//将处理新业务的方法和conn进行绑定
			dealConn := NewConntion(s, conn, cid, s.MsgHandler)
			cid++

			go dealConn.Start()
			////已经建立连接
			//go func() {
			//	for {
			//		buf := make([]byte, 512)
			//		cnt, err := conn.Read(buf)
			//		if err != nil {
			//			fmt.Println("接收数据异常：", err)
			//
			//		}
			//		//回显
			//		if _, err := conn.Write(buf[:cnt]); err != nil {
			//			fmt.Println("写入数据失败 ", err)
			//			continue
			//		}
			//	}
			//}()

		}
	}()
}

func (s *Server) Stop() {
	fmt.Println("server stop")
	s.GetConnMgr().ClearConn()

}

func (s *Server) Server() {

	s.Start()

	//阻塞
	select {}
}

// GetConnMgr 得到链接管理
func (s *Server) GetConnMgr() ziface.IConnManager {
	return s.ConnMgr
}

// NewServer 初始化方法，读取conf/zinx.json 中的配置来初始化
func NewServer() ziface.IServer {
	utils.GlobalObject.Reload()

	s := &Server{
		Name:           utils.GlobalObject.Name,
		IPVersion:      "tcp4",
		IP:             utils.GlobalObject.Host,
		Port:           utils.GlobalObject.TcpPort,
		MsgHandler:     NewMsgHandle(),
		WorkerPoolSize: utils.GlobalObject.WorkerPoolSize,
		ConnMgr:        NewConnManager(), //创建ConnManager
	}
	return s
}
