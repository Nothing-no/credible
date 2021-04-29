package credible

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"

	uuid "github.com/satori/go.uuid"
)

func genUid() string {
	var err error
	uid := uuid.Must(uuid.NewV3(uuid.NewV4(), "NTE"), err).String()
	if nil != err {
		fmt.Println(err)
	}
	return uid
}

func NewServer(ip string, port string) Server {
	return &server{
		ip:           ip,
		port:         port,
		clients:      make(map[string]net.Conn),
		handler:      make(map[string]func(interface{})),
		selfDealFunc: nil,
		exitFlag:     make(chan bool),
		ccNum:        0,
		maxBuff:      DefaultMaxBuffLen,
		RWMutex:      &sync.RWMutex{},
	}
}

func NewClient(ip, port string) Client {
	return &client{
		ip:      ip,
		port:    port,
		Conn:    nil,
		maxBuff: DefaultMaxBuffLen,
		handler: make(map[string]func(interface{})),
		RWMutex: &sync.RWMutex{},
	}
}

func prepareSendJson(cmd string, data interface{}) ([]byte, error) {
	headData := []byte{'N', 'T', 'E', 0x01}
	uid := msgQueue.add(map[string]interface{}{
		"order": cmd,
		"data":  data,
	})

	bodyData := &body{
		Uuid:  uid,
		Order: cmd,
		Data:  data,
	}

	bs, err := json.Marshal(bodyData)
	if nil != err {
		return bs, err
	}

	lenByte := convertI2B(len(bs))
	headData = append(headData, lenByte[0], lenByte[1], lenByte[2], lenByte[3])

	return append(headData, bs...), nil
}

func convertI2B(inData int) (outData [4]byte) {
	for i := 0; i < 4; i++ {
		outData[i] = byte((inData >> ((3 - i) * 8) & 0xff))
	}

	return
}

func convertB2I(inData []byte) (outData int, err error) {
	if len(inData) != 4 {
		err = fmt.Errorf("%v is over 4B", inData)
		return
	}
	for i, v := range inData {
		outData |= (int(v) << ((3 - i) << 3))
	}

	return
}

func respRaw(conn net.Conn, uid string) error {
	dataHead := []byte{'N', 'T', 'E', 0x10}
	respData := append(dataHead, []byte(uid)...)
	_, err := conn.Write(respData)
	return err
}

func checkHead(head ...byte) (typeFlag int, dataFmt int) {
	if head[0] == 'N' && head[1] == 'T' && head[2] == 'E' {
		tf := int(head[3]) >> 4
		df := int(head[3]) & 0x000f
		return tf, df
	}
	return -1, -1
}
