package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))

	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}

	client.conn = conn
	return client
}

func (client *Client) SelectUsers() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("client conn write err:", err)
		return
	}
}

func (client *Client) PrivateChat() {
	var remoteName string
	var chatMsg string
	for remoteName != "exit" {
		client.SelectUsers()
		fmt.Println(">>>>请输入聊天对象[用户名], exit退出")
		fmt.Scanln(&remoteName)
		if remoteName == "exit" {
			break
		}
		for chatMsg != "exit" {
			fmt.Println(">>>>请输入聊天内容, exit退出")
			fmt.Scanln(&chatMsg)
			if chatMsg == "exit" {
				break
			}
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Write err:", err)
					break
				}
			}
			chatMsg = ""
		}
	}
}

func (client *Client) PubliChat() {
	var publicChatMsg string
	for publicChatMsg != "exit" {
		fmt.Println(">>>>请输入公聊消息，exit退出.")
		fmt.Scanln(&publicChatMsg)
		if publicChatMsg == "exit" {
			break
		}
		if len(publicChatMsg) != 0 {
			publicChatMsg := publicChatMsg + "\n"
			_, err := client.conn.Write([]byte(publicChatMsg))
			if err != nil {
				fmt.Println("conn Write err:", err)
				break
			}
		}
		publicChatMsg = ""
	}
}

func (client *Client) UpdatedName() bool {
	fmt.Println(">>>>请输入用户名:")
	fmt.Scanln(&client.Name)
	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}
	return true
}

func (client *Client) DealResponse() {
	io.Copy(os.Stdout, client.conn)
}

func (client *Client) menu() bool {
	var flag int
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")
	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println(">>>>请输入合法范围内的数字>>>>")
		return false
	}
}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {

		}

		switch client.flag {
		case 1:
			fmt.Println("公聊模式选择...")
			client.PubliChat()
		case 2:
			fmt.Println("私聊模式选择...")
			client.PrivateChat()
		case 3:
			fmt.Println("更新用户名选择...")
			client.UpdatedName()
		}
	}
}

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置默认服务器IP地址(默认127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "设置默认服务器端口(默认8888)")
}

func main() {
	flag.Parse()
	client := NewClient(serverIp, serverPort)

	if client == nil {
		fmt.Println(">>>>>服务器连接失败>>>>>")
		return
	}

	go client.DealResponse()

	fmt.Println(">>>>服务器连接成功>>>>>")

	client.Run()
}
