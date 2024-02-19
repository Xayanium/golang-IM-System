package main

import (
	"fmt"
	"net"
	"strings"
)

type User struct { //User类
	Name   string
	Addr   string
	C      chan string //与当前用户绑定的channel
	conn   net.Conn    //客户端通信链接
	server *Server     //当前用户所属server
}

func NewUser(conn net.Conn, server *Server) *User { //User的构造函数
	userAddr := conn.RemoteAddr().String() //获得当前客户端地址

	user := &User{
		Name:   userAddr, //以当前客户端地址作为用户名
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	go user.ListenMessage() //启用监听当前user channel消息的go程
	return user
}

func (this *User) ListenMessage() {
	//监听当前的User channel,一旦有消息, 就直接发送给对接的客户端
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n")) //将消息发送给客户端(要转换成二进制数组形式才能发送)
	}
}

func (this *User) Online() {
	//用户上线, 将用户加入OnlineMap
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	//广播用户上线信息
	this.server.BroadCast(this, "log in")
} //用户上线的业务

func (this *User) Offline() {
	//用户下线, 将用户从OnlineMap删除
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	//广播用户下线信息
	this.server.BroadCast(this, "log out")
} //用户下线的业务

func (this *User) DoMessage(msg string) {
	fmt.Println(msg)

	if msg == "who" {
		//查询当前在线用户都有哪些
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMessage := fmt.Sprintf("%s is online\n", user.Name)
			//将查询到的消息传给发起查询的用户
			this.conn.Write([]byte(onlineMessage))
		}
		this.server.mapLock.Unlock()

	} else if len(msg) > 7 && msg[:7] == "rename|" { //更改用户名的方法
		//设定消息格式为 rename|xxxx
		newName := strings.Split(msg, "|")[1] //从字符串中通过某个字符截取, 将两部分放入不同数组中

		//判断name是否已经存在
		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.conn.Write([]byte("this username has been used\n"))
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name) //删掉map中之前的名字
			this.Name = newName                      //更新user实例中的名字
			this.server.OnlineMap[newName] = this    //更改为想要的名字
			this.server.mapLock.Unlock()
			this.conn.Write([]byte("You have successfully updated your name\n"))
		}

	} else if len(msg) > 4 && msg[:3] == "to|" { //私聊的方法
		//设定消息格式为 to|username|message
		//1. 获取对方的用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			this.conn.Write([]byte("format error, please use the correct format"))
			return
		}
		//2. 根据用户名得到user对象
		remoteUser, ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.conn.Write([]byte("username not exist"))
		}
		//3. 获取消息内容, 通过对方的User对象将消息发过去
		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.conn.Write([]byte("message is empty"))
			return
		}
		remoteUser.conn.Write([]byte(this.Name + " tell you " + content + "\n"))
	} else {
		this.server.BroadCast(this, msg)
	}
} //用户处理消息的业务
