package credible

import (
	"net"
	"sync"
)

type (
	//Server interface
	Server interface {
		Run() error
		//send to all clients
		Sender
		Register
		//CurrentConnected 获取当前连接数
		CurrentConnected() int

		//SetMaxBuff 设置buffer最大长度,默认是1024B
		SetMaxBuff(int)

		GetMaxBuff() int

		SetSelfDealFunc(func(net.Conn, []byte) error) Server
	}

	//Client interface
	Client interface {
		Connect() error
		//send to server
		Sender
		Register
		SetSelfDealFunc(func(net.Conn, []byte) error) Client
	}

	//Sender
	Sender interface {
		Send(string, interface{}) Sender
	}

	Register interface {
		RegisterHandler(string, func(interface{})) Register
	}

	//real server structure
	server struct {
		ip           string
		port         string
		clients      map[string]net.Conn
		handler      map[string]func(interface{})
		selfDealFunc func(net.Conn, []byte) error
		exitFlag     chan bool
		ccNum        int //current connections
		maxBuff      int
		*sync.RWMutex
	}

	client struct {
		net.Conn
		*sync.RWMutex
		ip           string
		port         string
		maxBuff      int
		selfDealFunc func(net.Conn, []byte) error
		handler      map[string]func(interface{})
	}

	msgQ struct {
		msg map[string]interface{}
		*sync.RWMutex
	}

	body struct {
		Uuid  string      `json:"uuid"`
		Order string      `json:"order"`
		Data  interface{} `json:"data"`
	}
)

const (
	ocPro             = "NTE"
	Undefined         = -1
	DefaultMaxBuffLen = 1024
)

//flag
const (
	ocSendFlag = iota
	ocRespFlag
	ocDebugFlag
)

//format
const (
	ocRaw = iota
	ocJson
)

var (
	msgQueue = &msgQ{
		msg:     make(map[string]interface{}),
		RWMutex: &sync.RWMutex{},
	}
)

var (
	_ Server   = &server{}
	_ Sender   = &server{}
	_ Register = &server{}
	_ Client   = &client{}
	_ Sender   = &client{}
	_ Register = &client{}
)

func (my *msgQ) checkExist(uid string) bool {
	my.RLock()
	defer my.RUnlock()
	_, ok := my.msg[uid]

	return ok
}

func (my *msgQ) remove(uid string) {
	my.Lock()
	defer my.Unlock()
	delete(my.msg, uid)
}

func (my *msgQ) add(d interface{}) string {
	uid := genUid()
	my.Lock()
	defer my.Unlock()
	my.msg[uid] = d

	return uid
}

func IterMsgQueue(dealFunc func(id string, data interface{})) {
	msgQueue.RLock()
	tmp := msgQueue
	msgQueue.RUnlock()
	for key, data := range tmp.msg {
		go dealFunc(key, data)
	}
}

func LenOfMsgQueue() int {
	return len(msgQueue.msg)
}
