package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func PostFormFile(url string, files ...string) ([]byte, error) {
	var (
		fileWriter  io.Writer
		fileHandler *os.File
		err         error
		resp        *http.Response
		contType    string
		result      []byte
	)
	if len(files) == 0 {
		return []byte{}, errors.New("lack of file")
	}
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	// bodyWriter.CreatePart(make(textproto.MIMEHeader))
	fileWriter, err = bodyWriter.CreateFormFile("files", filepath.Base(files[0]))
	if nil != err {
		goto errReturn
	}

	fileHandler, err = os.Open(files[0])
	if nil != err {
		goto errReturn
	}
	defer fileHandler.Close()

	_, err = io.Copy(fileWriter, fileHandler)
	if nil != err {
		goto errReturn
	}

	contType = bodyWriter.FormDataContentType()
	bodyWriter.Close()
	fmt.Println(contType)
	resp, err = http.Post(url, contType, bodyBuf)
	if nil != err {
		goto errReturn
	}
	defer resp.Body.Close()

	result, err = ioutil.ReadAll(resp.Body)
	if nil != err {
		goto errReturn
	}

	return result, nil

errReturn:
	return []byte{}, err
}

var (
	fileUpload = flag.String("f", "./go.mod", "--")
	url        = flag.String("url", "http://127.0.0.1:19999/upload", "0-")
)

func main() {
	flag.Parse()
	bs, err := PostFormFile(*url, *fileUpload)
	if nil != err {
		fmt.Println(err)
		return
	}

	fmt.Println(string(bs))
	// router := gin.Default()
	// router.POST("/upload", handler)
	// router.Run(":19999")
}

func handler(c *gin.Context) {
	mulForm, err := c.MultipartForm()
	if nil != err {
		fmt.Println(err)
		c.JSON(200, gin.H{
			"err": err,
		})
		return
	}
	files := mulForm.File["files"]
	for _, f := range files {
		err := c.SaveUploadedFile(f, "./log/"+f.Filename)
		if nil != err {
			fmt.Println(err)
			c.JSON(200, gin.H{
				"err": err,
			})
			return
		}
	}
	c.String(200, "Ok")
}

// //credible server exmaple
// func main() {
// 	// credible.RedirectLevel(credible.LEVEL_ERROR, credible.NewFile("error.log"))
// 	// credible.RedirectLevel(credible.LEVEL_DEBUG, credible.NewFile("debug.log"))

// 	// srv := credible.NewServer("0.0.0.0", "16666")

// 	// srv.RegisterHandler("hello", hello).RegisterHandler("greet", greet)
// 	// srv.SetSelfDealFunc(selfDeal)

// 	// srv.Run()
// 	// cnt := 0
// 	// for {
// 	// 	cnt++
// 	// 	time.Sleep(3 * time.Second)
// 	// 	fmt.Println(srv.CurrentConnected())
// 	// 	if cnt < 10 {
// 	// 		srv.Send("test", "")
// 	// 	}
// 	// 	fmt.Println("msgqueue:", credible.LenOfMsgQueue())
// 	// }
// 	cli := credible.NewClient("127.0.0.1", "16666")
// 	err := cli.Connect()
// 	if nil != err {
// 		fmt.Println(err)
// 		return
// 	}
// 	cnt := 0
// 	for {
// 		cnt++
// 		time.Sleep(3 * time.Second)
// 		if cnt < 5 {
// 			go func(c int) {
// 				if c%2 == 0 {
// 					for i := 0; i < 50; i++ {
// 						// time.Sleep(500 * time.Millisecond)
// 						if (i*cnt)%2 != 0 {

// 							go cli.Send("hello", nil)
// 						} else {
// 							go cli.Send("greet", map[string]interface{}{
// 								"hello":  "jack",
// 								"hi":     "Mary",
// 								"Nothin": 1,
// 								"Peter":  1.23,
// 							})

// 						}
// 					}
// 				} else {
// 					for i := 0; i < 50; i++ {
// 						// time.Sleep(500 * time.Millisecond)
// 						if i%2 == 0 {

// 							go cli.Send("greet", "nothing")
// 						} else {

// 							go cli.Send("hello", "nothing")
// 						}
// 					}
// 				}
// 			}(cnt)
// 		}
// 		fmt.Println(credible.LenOfMsgQueue())
// 	}
// }

// func hello(data interface{}) {
// 	fmt.Println("hello world")
// }

// func greet(data interface{}) {
// 	fmt.Println("hello", data)
// }

// func selfDeal(conn net.Conn, data []byte) error {

// 	fmt.Println("self deal func, recv:", string(data))
// 	conn.Write(data)
// 	return nil
// }
