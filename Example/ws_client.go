package main

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"golang.org/x/net/websocket"
	"io"
	"io/ioutil"
	"os"
)

type FlagType int
type ErrorCode int

type WSResponse struct {
	Code  ErrorCode `json:"status"`
	Msg   string    `json:"msg"`
	ReqId string    `json:"reqId"`
	Key   string    `json:"key"`
	Val   string    `json:"result"`
	Flag  FlagType  `json:"flag"`
}

func printUsage() {
	fmt.Printf("Usage:\n%s -m meta [/opt/data/meta.txt] -f filename[/opt/data/input.wav] -h host[localhost] -p port[8081] -e endpoint[/llcup/stream/upload]\n", os.Args[0])
}

var chunkSize = 4096
var maxResponseLen = 1024 * 1024

func receiveResponse(ws *websocket.Conn, waitc chan<- struct{}) {
	//receive response over
	//noity it is over
	defer close(waitc)

	for {
		//get response len
		header := make([]byte, 4)
		readLen, err := ws.Read(header)
		if err != nil {
			fmt.Printf("receive data slice from wsocket failed, err:%v\n", err)
			return
		}

		if readLen != 4 {
			msg := fmt.Sprintf("receive message header not equal 4 bytes, it is:%d\n", readLen)
			fmt.Printf(msg)
			return
		}

		//decode response len
		dataLen := int(binary.BigEndian.Uint32(header))
		if dataLen > maxResponseLen {
			fmt.Printf("receive message len:(%d) exceeds MAX_HEADER_LEN:(%d)\n", dataLen, maxResponseLen)
			return
		}

		//read response slice
		bytes := make([]byte, dataLen)
		unreadCount := dataLen
		for {
			if unreadCount == 0 {
				break
			}
			buf := make([]byte, unreadCount)
			bufLen, err := ws.Read(buf)
			if err != nil {
				fmt.Printf("receive data slice from wsocket failed, err:%v\n", err)
				return
			}
			//copy buf to bytes
			copy(bytes[dataLen-unreadCount:], buf)
			unreadCount = unreadCount - bufLen
		}

		//decode response to json
		var rsp WSResponse
		err = json.Unmarshal(bytes, &rsp)
		if err != nil {
			msg := fmt.Sprintf("%s", string(bytes))
			fmt.Printf("json unmarshal failed:%v,data:%s\n", err, msg)
			return
		}

		//check response status
		fmt.Printf("code:%d, msg:%s, key:%s, val:%s, flag:%d\n", rsp.Code, rsp.Msg, rsp.Key, rsp.Val, rsp.Flag)
		if rsp.Code != 0 {
			fmt.Printf("receive response failed, errcode:%d\n", rsp.Code)
			return
		}

		//check response state machine
		if rsp.Flag == 1 {
			fmt.Printf("receive response done, exiting..\n")
			return
		}
	}
}

func makeRequest(host string, port string, metaFilename string, fileName string, endpoint string) {
	//words file
	var header []byte
	var meta string
	if tmpMeta, err := ioutil.ReadFile(metaFilename); err != nil {
		fmt.Printf("read meta file:%s failed,err:%s\n", metaFilename, err)
		return
	} else {
		meta = string(tmpMeta)
	}

	waitc := make(chan struct{})

	base64Meta := base64.StdEncoding.EncodeToString([]byte(meta))
	dataLen := len(base64Meta)
	header = make([]byte, dataLen+4)
	binary.BigEndian.PutUint32(header, uint32(dataLen))
	copy(header[4:], base64Meta)

	f, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("open file:%s failed!,err:%s\n", fileName, err)
		return
	}
	defer f.Close()

	origin := "http://localhost/"
	url := fmt.Sprintf("ws://%s:%s%s", host, port, endpoint)
	//add nginx as front load-balance-proxy test case
	//url := "ws://127.0.0.1:80/ws/upload/noauth?qid=100001&quality=0"
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		fmt.Printf("failed when dial, err:%v\n", err)
		return
	}

	//send header
	if err := websocket.Message.Send(ws, header); err != nil {
		fmt.Printf("write header failed, err:%v\n", err)
		return
	}

	//begin receive response
	go receiveResponse(ws, waitc)

	//send stream
	for {
		var data = make([]byte, chunkSize)
		count, err := f.Read(data)
		if err == io.EOF {
			break
		}

		if err := websocket.Message.Send(ws, data[:count]); err != nil {
			//fmt.Fatal(err)
			fmt.Printf(" write err:%v\n", err)
			return
		}
	}

	//send eos package
	//"EOS"->"0x450x4f0x53"
	var eos = []byte{0x45, 0x4f, 0x53}
	if err := websocket.Message.Send(ws, eos); err != nil {
		//fmt.Fatal(err)
		fmt.Printf("failed when write last package, err:%v\n", err)
		return
	}

	fmt.Printf("send data over\n")
	<-waitc
}

func main() {
	hostPtr := flag.String("h", "localhost", "host to connect to acceptor")
	portPtr := flag.String("p", "8081", "port to connect to acceptor")
	metaFileNamePtr := flag.String("m", "", "meta file name to compute,for example[/opt/data/meta.txt]")
	fileNamePtr := flag.String("f", "", "file name to compute,for example[/opt/data/input.wav]")
	endpointPtr := flag.String("e", "/llcup/stream/upload", "endpoint for the request,for example[/llcup/stream/upload]")

	flag.Parse()

	if *fileNamePtr == "" || *metaFileNamePtr == "" {
		printUsage()
		return
	}

	makeRequest(*hostPtr, *portPtr, *metaFileNamePtr, *fileNamePtr, *endpointPtr)
}
