package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Server struct { //先创建一个Server类
	Ip   string
	Port int

	//在线用户的列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex //由于OnlineMap是全局的, 所以可以加一个读写锁

	//消息广播的channel
	Message chan string
}

func NewServer(ip string, port int) *Server { //作为一个Server类的构造器
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

func (this *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.listen err:　", err)
		return
	}

	//启动监听Message的goroutine
	go this.ListenMessage()

	for {
		//accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("net.listen err: ", err)
			continue
		}

		//do handler
		//为了不阻塞下一次accept, 要用go携程来处理当前链接的业务
		go this.Handler(conn)
	}

	//close listen socket
	listener.Close()
} //用于启动服务器的程序

func (this *Server) Handler(conn net.Conn) {
	user := NewUser(conn, this)

	//用户上线操作
	user.Online()

	//接受客户端发来的消息
	go func() {
		buf := make([]byte, 4096)
		//调用conn.Read方法从当前connection中读取数据到buf, 返回读取的字节数和错误
		for {
			n, err := conn.Read(buf)
			if n == 0 { //表示客户端关闭

				user.Offline()
				return
			}
			if err != nil && err != io.EOF { //每次读完都会有个EOF标志结尾, 如果条件成立那就一定是进行了一次非法操作了
				fmt.Println("Conn Read err: ", err)
			}
			//提取用户的消息(去掉'\n')
			msg := string(buf[:n-1])
			//用户针对msg消息进行处理
			user.DoMessage(msg)
		}
	}()
}

func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := fmt.Sprintf("[%s]%s: %s  \n", user.Addr, user.Name, msg)

	this.Message <- sendMsg //将消息放入msg channel中
} //处理要发送的消息, 存入msg channel

func (this *Server) ListenMessage() {
	for {
		msg := <-this.Message

		//将msg发送给存储在OlineMap中的User
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg //实现消息从message的channel到user的channel
		}
		this.mapLock.Unlock()
	}
} //监听message, 一旦有消息就发送给在线用户
