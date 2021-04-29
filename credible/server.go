package credible

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"runtime"
	"time"
)

func (my *server) Run() error {
	go my.listen()
	return nil
}

func (my *server) Send(order string, data interface{}) Sender {
	for _, conn := range my.clients {
		if nil != conn {
			bs, err := prepareSendJson(order, data)
			if nil != err {
				Error(err)
			}
			cnt := 0
		signResend:
			cnt++
			if cnt > 3 {
				Error("resend over 3 times")
				continue
			}
			_, err = conn.Write(bs)
			if nil != err {
				goto signResend
			}
		}
	}
	return my
}

//RegisterHandler ...
/*
*	@date:2021/04/15
*	@version: V1.0.01
*	@debug: 实现对应指令的处理函数注册
 */
func (my *server) RegisterHandler(order string, handlerFunc func(interface{})) Register {
	my.Lock()
	defer my.Unlock()
	my.handler[order] = handlerFunc

	return my
}

func (my *server) CurrentConnected() int {
	my.RLock()
	my.RUnlock()
	return my.ccNum
}

func (my *server) SetMaxBuff(maxBuffLen int) {
	my.Lock()
	defer my.Unlock()

	my.maxBuff = maxBuffLen
}

func (my *server) GetMaxBuff() int {
	my.RLock()
	defer my.RUnlock()
	return my.maxBuff
}

func (my *server) Close() {
	for k, conn := range my.clients {
		conn.Close()
		delete(my.clients, k)
	}
	for k := range my.handler {
		delete(my.handler, k)
	}

	my = nil
}

func (my *server) listen() {
	defer my.Close()

	retryCnt := 0
relisten:
	retryCnt++
	if retryCnt > 3 {
		panic("re-listen over 3 times")
	}

	lsner, err := net.Listen("tcp", my.ip+":"+my.port)
	if nil != err {
		Error(err)
		goto relisten
	}
	defer lsner.Close()

	for true {
		select {
		case <-my.exitFlag:
			fmt.Println("exit")
			runtime.Goexit()
		default:
			conn, err := lsner.Accept()
			if nil != err {
				Error(err)
				continue
			}
			Debugf("%s connected", conn.RemoteAddr().String())
		reGenUid:
			clientId := genUid()
			_, ok := my.clients[clientId]
			if ok {
				goto reGenUid
			}
			//add client
			my.clients[clientId] = conn
			my.Lock()
			my.ccNum++
			my.Unlock()
			go my.read(clientId)
		}
	}
}

func (my *server) read(clientId string) error {
	var (
		data = make([]byte, my.maxBuff)
		done = make(chan bool, 1)
	)

	for true {
		var tmpData []byte
	readAgain:
		n, err := my.clients[clientId].Read(data)
		if nil != err {
			my.Lock()
			my.ccNum--
			my.Unlock()
			if err == io.EOF {
				Debugf("%s disconnected", my.clients[clientId].RemoteAddr().String())
				my.clients[clientId].Close()
				break
			}

			delete(my.clients, clientId)

			Debug(err)
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
					my.dealData(clientId, tmpData)
				}
				tmpData = []byte{}
				return
			}()

			goto readAgain
		} else {
			tmpData = append(tmpData, data[:n]...)
			my.dealData(clientId, tmpData)
		}

	}

	return nil
}

func checkJson(data []byte) ([]byte, []byte) {
	const (
		startToken = '{'
		endToken   = '}'
		deliToken  = '"'
	)

	var (
		startFlag = false
		deliCnt   = 0
		startLoc  = 0
		startCnt  = 0
		endCnt    = 0
	)

	for i, c := range data {
		if c == startToken {
			endCnt++
			startCnt++
			if !startFlag {
				startLoc = i
				startFlag = true
				//Debug(startLoc)
				continue
			}
		}
		if startFlag {
			if c == deliToken {
				deliCnt++
				//Debug(deliCnt)
			} else if deliCnt%2 == 0 && c == endToken {
				startCnt--
				endCnt++
				//Debugf("startCnt:%d, endCnt:%d", startCnt, endCnt)
				if startCnt == 0 && endCnt%2 == 0 {
					if len(data) == i+1 {
						return data[startLoc : i+1], []byte{}
					}
					return data[startLoc : i+1], data[i+1:]
				}
			}
		}
	}

	return []byte{}, []byte{}
}

func (my *server) dealException(cid string, data []byte) {
	real, rem := checkJson(data)
	if len(real) != 0 {
		var bodyData body
		err := json.Unmarshal(real, &bodyData)
		if nil != err {
			Debug(err)
			goto doAnother
		}
		my.Lock()
		go my.handler[bodyData.Order](bodyData.Data)
		my.Unlock()
		trycnt := 0
	signRetry:
		trycnt++
		err = respRaw(my.clients[cid], bodyData.Uuid)
		if nil != err {
			Error(err)
			if trycnt < 3 {
				goto signRetry
			}
		}
	}
doAnother:
	if len(rem) != 0 {
		my.dealData(cid, rem)
	}
}

func (my *server) dealData(clientId string, tmpData []byte) {
	if len(tmpData) < 8 {
		return
	}
	tf, df := checkHead(tmpData[:4]...)
	switch tf {
	case ocSendFlag:
		lenOfData, _ := convertB2I(tmpData[4:8])
		var bodyData body
		if ocJson == df {
			var err error
			if lenOfData == len(tmpData[8:]) {
				// Debug(lenOfData)
				err = json.Unmarshal(tmpData[8:], &bodyData)
			} else if lenOfData < len(tmpData[8:]) {
				// fmt.Println(string(tmpData[8 : lenOfData+8]))
				err = json.Unmarshal(tmpData[8:8+lenOfData], &bodyData)
				// fmt.Println(string(tmpData[8+lenOfData:]))
				go my.dealData(clientId, tmpData[8+lenOfData:])
			} else {
				// fmt.Println("------------------------------>")
				Error("error:there is something wrong")
				return
			}
			if nil != err {

				Error(map[string]interface{}{"err": err, "data": string(tmpData[8:])})

			} else {

				my.RLock()
				if _, ok := my.handler[bodyData.Order]; ok {
					go my.handler[bodyData.Order](bodyData.Data)
				}
				my.RUnlock()
				trycnt := 0
			signRetry:
				trycnt++
				err := respRaw(my.clients[clientId], bodyData.Uuid)
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
				go my.selfDealFunc(my.clients[clientId], tmpData[8:])
			}
		} else {
			Debugf("data fomat:%d", df)
		}
	//若为回复信息，则删除发送队列uid
	case ocRespFlag:
		uid := string(tmpData[4:])
		if len(uid) > 36 {
			uid = string(tmpData[4:40])
			go my.dealData("", tmpData[40:])
		}
		if msgQueue.checkExist(uid) {
			fmt.Println("remove:", uid)
			msgQueue.remove(uid)
		}
	case ocDebugFlag:
		fmt.Println("to be developed")
	case Undefined:
		fmt.Println("undefined format")
		if nil != my.selfDealFunc {
			go my.selfDealFunc(my.clients[clientId], tmpData)
		}
	default:
		if nil != my.selfDealFunc {
			go my.selfDealFunc(my.clients[clientId], tmpData[3:])
		}
	}
}

func (my *server) SetSelfDealFunc(selfDealFunc func(net.Conn, []byte) error) Server {
	my.selfDealFunc = selfDealFunc
	return my
}
