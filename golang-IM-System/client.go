package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int //保存当前client的menu选择
}

func NewClient(serverIp string, serverPort int) *Client {
	//创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       114514,
	}
	//链接服务器
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort)) //传入网络类型和ip地址, 返回连接对象和错误
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}
	client.conn = conn
	//返回创建对象
	return client
}

func (this *Client) Menu() bool {
	var flag int
	time.Sleep(500)
	fmt.Println("input 1 : public chat")
	fmt.Println("input 2 : private chat")
	fmt.Println("input 3 : change username")
	fmt.Println("input 0 : exit")

	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		this.flag = flag
		return true
	} else {
		fmt.Println("Please input the correct number")
		return false
	}
}

func (this *Client) Run() {
	for this.flag != 0 {

		for this.Menu() != true {
		}

		switch this.flag {
		case 1: //公聊模式
			this.PublicChat()
			break
		case 2: //私聊模式
			this.PrivateChat()
			break
		case 3: //更新用户名
			this.UpdateName()
			break
		}
	}
}

func (this *Client) UpdateName() bool {
	fmt.Println(">>>Please enter Your username:>>>")
	fmt.Scanln(&this.Name)

	sendMsg := "rename|" + this.Name + "\n"
	_, err := this.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}
	return true
}

func (this *Client) DealResponse() {
	//可以永久阻塞监听
	io.Copy(os.Stdout, this.conn) //不断等待服务端传回的数据, 一旦服务端传回数据就立刻输出到终端
}

func (this *Client) PublicChat() {
	var chatMsg string
	fmt.Println(">>>Please input your chat message('exit' to quit):")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" { //不断监听输入的消息

		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n" //根据公聊的协议
			_, err := this.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn.Write err:", err)
				break
			}
		}

		time.Sleep(1500)
		chatMsg = ""
		fmt.Println(">>>Please input your chat message('exit' to quit):")
		fmt.Scanln(&chatMsg)
	}
}

func (this *Client) SelectUser() {
	fmt.Println("users online:")
	sendMsg := "who\n"
	_, err := this.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn Write error:", err)
		return
	}
}

func (this *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	this.SelectUser()
	time.Sleep(1000)
	fmt.Println(">>>Please input the username('exit' to quit):")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" { //对消息发送的用户进行选择
		fmt.Println("Please enter context('exit' to quit)")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" { //不断监听输入的消息

			if len(chatMsg) != 0 {
				//根据私聊的协议, 因为我们写的服务端会截断最后一个换行, 所以需要两个'\n'
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n\n"
				_, err := this.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn.Write err:", err)
					break
				}
			}

			chatMsg = ""
			fmt.Println(">>>Please input your chat message('exit' to quit):")
			fmt.Scanln(&chatMsg)
		}

		this.SelectUser()
		fmt.Println(">>>Please input the username('exit' to quit):")
		fmt.Scanln(&remoteName)
	}
}

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "set IP address")
	flag.IntVar(&serverPort, "port", 8000, "set Port")
}

func main() {
	//命令行解析
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println("connection error......")
		return
	}

	go client.DealResponse()

	time.Sleep(500)
	fmt.Println("connection start successfully!")

	//启动客户端的业务
	client.Run()
}
