package main

import (
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

type Server struct {
	Ip        string
	Port      int
	OnlineMap map[string]*User
	mapLock   sync.RWMutex
	Message   chan string
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	return server
}

func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message

		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

func (this *Server) Broadcast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

func (this *Server) Handler(conn net.Conn) {
	// fmt.Println("连接建立成功")
	user := NewUser(conn)
	this.Online(user)

	isLive := make(chan bool)

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				this.Broadcast(user, "下线")
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}

			msg := string(buf[:n-1])
			this.DoMessage(user, msg)
			isLive <- true
		}
	}()
	for {
		select {
		case <-isLive:

		case <-time.After(time.Second * 100000):
			this.SendMsg(user, "你被踢了")
			time.Sleep(time.Second * 1)
			this.Offline(user)
			close(user.C)
			conn.Close()
			return
		}
	}
}

func (this *Server) Online(user *User) {
	this.mapLock.Lock()
	this.OnlineMap[user.Name] = user
	this.mapLock.Unlock()

	this.Broadcast(user, "已上线")
}

func (this *Server) Offline(user *User) {
	this.mapLock.Lock()
	delete(this.OnlineMap, user.Name)
	this.mapLock.Unlock()

	this.Broadcast(user, "下线")
}

func (this *Server) SendMsg(user *User, msg string) {
	user.C <- msg
}

func (this *Server) DoMessage(user *User, msg string) {
	if msg == "who" {
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			onlineMsg := cli.Addr + "]" + cli.Name + ":" + "在线..."
			user.C <- onlineMsg
		}
		this.mapLock.Unlock()
	} else if len(msg) > 4 && msg[:3] == "to|" {
		//获取对方姓名,内容
		remoteName, content := strings.Split(msg, "|")[1], strings.Split(msg, "|")[2]
		//校验
		remoteUser, ok := this.OnlineMap[remoteName]
		if remoteName == "" || !ok {
			this.SendMsg(user, "发送的对象不存在")
			return
		}
		if content == "" {
			this.SendMsg(user, "无消息内容，请重发")
			return
		}
		//发送信息
		this.SendMsg(remoteUser, user.Name+"对您说:"+content)
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		newName := msg[7:]
		_, ok := this.OnlineMap[newName]
		if ok {
			this.SendMsg(user, "当前用户名已被使用")
		} else {
			this.mapLock.Lock()
			delete(this.OnlineMap, user.Name)
			this.OnlineMap[newName] = user
			this.mapLock.Unlock()
			user.Name = newName
			this.SendMsg(user, "您已经更新用户名"+newName)
		}
	} else {
		this.Broadcast(user, msg)
	}
}

func (this *Server) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net Listen err:", err)
		return
	}
	defer listener.Close()

	go this.ListenMessager()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}

		go this.Handler(conn)
	}
}
