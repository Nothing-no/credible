package credible

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"
)

func (my *client) Connect() error {
	conn, err := my.doConnect()
	if nil != err {
		return err
	}
	my.Conn = conn

	go my.read()
	return nil
}

func (my *client) doConnect() (net.Conn, error) {
	return net.Dial("tcp", my.ip+":"+my.port)
}

func (my *client) Send(cmd string, data interface{}) Sender {
	if nil != my {
		bs, err := prepareSendJson(cmd, data)
		if nil != err {
			Error(err)
		}
		go func([]byte) {
			my.Lock()
		signResend:
			_, err = my.Write(bs)
			if nil != err {
				my.Connect()
				//my.Unlock()
				goto signResend
			}
			my.Unlock()
		}(bs)
	}

	return my
}

func (my *client) RegisterHandler(cmd string, handler func(interface{})) Register {
	my.Lock()
	defer my.Unlock()
	my.handler[cmd] = handler
	return my
}

func (my *client) read() error {
	defer my.Close()
	var (
		data = make([]byte, my.maxBuff)
		done = make(chan bool, 1)
	)

	for true {
		var tmpData []byte
	readAgain:
		n, err := my.Read(data)
		if nil != err {
			if err == io.EOF {
				Debugf("%s disconnected\n", my.RemoteAddr().String())
				break
			}

			Debugf("error:%v\n", err)
			return err
		}

		//判断数据是否一次未读完
		if len(tmpData) != 0 {
			done <- true
		}

		if n == my.maxBuff {
			tmpData = append(tmpData, data[:n]...)
			go func() {
				select {
				case <-done:
					return
				case <-time.After(30 * time.Millisecond):
					my.dealData("", tmpData)
				}
				tmpData = []byte{}
				return
			}()

			goto readAgain
		} else {
			tmpData = append(tmpData, data[:n]...)
			my.dealData("", tmpData)
		}

	}

	return nil
}

func (my *client) dealData(some string, tmpData []byte) {
	tf, df := checkHead(tmpData[:4]...)
	switch tf {
	case ocSendFlag:
		lenOfData, _ := convertB2I(tmpData[4:8])
		var bodyData body
		if ocJson == df {
			var err error
			if lenOfData == len(tmpData[8:]) {
				err = json.Unmarshal(tmpData[8:], &bodyData)
			} else if lenOfData < len(tmpData[8:]) {
				err = json.Unmarshal(tmpData[8:8+lenOfData], &bodyData)

				//处理剩下的数据
				go my.dealData("", tmpData[8+lenOfData:])
			} else {
				Error("error: there is something wrong")
				return
			}
			if nil != err {
				Error(err)
			} else {
				my.RLock()
				if _, ok := my.handler[bodyData.Order]; ok {
					go my.handler[bodyData.Order](bodyData.Data)
				}
				my.RUnlock()
				trycnt := 0
			signRetry:
				trycnt++
				err := respRaw(my.Conn, bodyData.Uuid)
				if nil != err {
					Error(err)
					if trycnt < 3 {
						goto signRetry
					}
				}
			}
		} else if ocRaw == df {
			//处理发送的数据不是JSON格式的
			fmt.Println("to be developed")
			if nil != my.selfDealFunc {
				go my.selfDealFunc(my.Conn, tmpData[8:])
			}
		} else {
			//有待设计
			Debug("to be developed")
		}
	//若为回复信息，则删除发送队列uid
	case ocRespFlag:
		uid := string(tmpData[4:])
		if len(uid) > 36 {
			uid = string(tmpData[4:40])
			go my.dealData("", tmpData[40:])
		}
		if msgQueue.checkExist(uid) {
			fmt.Println("remove:", uid, len(uid))
			msgQueue.remove(uid)
		}
	case ocDebugFlag:
		fmt.Println("to be developed")
	case Undefined:
		fmt.Println("undefined format")
		if nil != my.selfDealFunc {
			go my.selfDealFunc(my.Conn, tmpData)
		}
	default:
		fmt.Println("others")
		if nil != my.selfDealFunc {
			go my.selfDealFunc(my.Conn, tmpData)
		}
	}
}

func (my *client) SetSelfDealFunc(selfDealFunc func(net.Conn, []byte) error) Client {
	my.selfDealFunc = selfDealFunc
	return my
}
