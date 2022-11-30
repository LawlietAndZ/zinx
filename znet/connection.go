package znet

import (
	"errors"
	"fmt"
	"io"
	"net"
	"zinx/utils"
	"zinx/ziface"
)

type Connection struct {
	Conn     *net.TCPConn //当前链接的TCP嵌套字
	ConnID   uint32       //链接的ID
	isClosed bool         //当前链接的状态

	////该连接的处理方法router
	//Router ziface.IRouter

	//消息管理MsgId和对应处理方法的消息管理模块
	MsgHandler ziface.IMsgHandle

	ExitChan chan bool

	//无缓冲管道，用于读、写两个goroutine之间的消息通信
	msgChan chan []byte

	//当前Conn属于那个server
	TCPServer ziface.IServer
}

// SendMsg 直接将Message数据发送数据给远程的TCP客户端
func (c *Connection) SendMsg(msgId uint32, data []byte) error {
	if c.isClosed == true {
		return errors.New("connection closed when send msg")
	}
	//将data封包，并且发送
	dp := NewDataPack()
	msg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return errors.New("Pack error msg ")
	}

	//写回客户端
	c.msgChan <- msg

	return nil
}

func (c *Connection) GetTCPConnection() *net.TCPConn {
	//TODO implement me
	return c.Conn
}

// NewConntion 创建连接的方法
func NewConntion(server ziface.IServer, conn *net.TCPConn, connID uint32, msgHandler ziface.IMsgHandle) *Connection {
	c := &Connection{
		Conn:       conn,
		ConnID:     connID,
		isClosed:   false,
		MsgHandler: msgHandler,
		ExitChan:   make(chan bool, 1),
		msgChan:    make(chan []byte), //msgChan初始化
		TCPServer:  server,
	}
	//将conn加到connManager方法
	server.GetConnMgr().Add(c)

	return c
}

func (c *Connection) StartReader() {

	fmt.Println("开始读数据的协程")
	defer c.Stop()

	for {
		//创建解包对象
		dp := NewDataPack()

		//读取客户端的Msg,Head
		headDate := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(c.GetTCPConnection(), headDate); err != nil {
			fmt.Println("read msg head error ", err)
			c.ExitChan <- true
		}

		//拆包，得到msgId和dateLen存入msg中
		msg, err := dp.Unpack(headDate)
		if err != nil {
			fmt.Println("unpack error ", err)
			c.ExitChan <- true
			continue
		}

		//根据 dataLen 读取 data，放在msg.Data中
		var data []byte
		if msg.GetDataLen() > 0 {
			data = make([]byte, msg.GetDataLen())
			if _, err := io.ReadFull(c.GetTCPConnection(), data); err != nil {
				fmt.Println("read msg data error ", err)
				c.ExitChan <- true
				continue
			}
		}
		msg.SetData(data)

		//得到当前客户端请求的Request数据
		req := Request{
			conn: c,
			msg:  msg, //将之前的buf 改成 msg
		}
		//从路由Routers 中找到注册绑定Conn的对应Handle
		if utils.GlobalObject.WorkerPoolSize > 0 {
			//已经启动工作池机制，将消息交给Worker处理
			c.MsgHandler.SendMsgToTaskQueue(&req)
		} else {
			//从绑定好的消息和对应的处理方法中执行对应的Handle方法
			go c.MsgHandler.DoMsgHandler(&req)
		}
	}
}

// StartWriter 写消息Goroutine， 用户将数据发送给客户端
func (c *Connection) StartWriter() {

	fmt.Println("[Writer Goroutine is running]")
	defer fmt.Println(c.GetRemote().String(), "[conn Writer exit!]")

	for {
		select {
		case data := <-c.msgChan:
			//有数据要写给客户端
			if _, err := c.Conn.Write(data); err != nil {
				fmt.Println("Send Data error:, ", err, " Conn Writer exit")
				return
			}
		case <-c.ExitChan:
			//conn已经关闭
			return
		}
	}
}

func (c *Connection) Start() {

	fmt.Println("启动链接，链接id=", c.ConnID)
	//启动读数据的业务
	go c.StartReader()
	// 启动写
	go c.StartWriter()
}

func (c *Connection) Stop() {
	fmt.Println("关闭链接，链接id=", c.ConnID)
	if !c.isClosed {
		return
	}
	c.isClosed = true
	c.Conn.Close()

	c.ExitChan <- true

	c.TCPServer.GetConnMgr().Remove(c)
	//回收资源
	close(c.ExitChan)
	close(c.msgChan)

}

func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}

func (c *Connection) GetRemote() net.Addr {
	return c.Conn.RemoteAddr()
}
