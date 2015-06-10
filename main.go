// sender project main.go
package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func checkerr(err error) {
	if err != nil {
		log.Fatal("[Red] ", err)
	}
}

type ShareFileInfo struct {
	Path string
	Name string
	Size int64
}

func shareDir(conn net.Conn, pathstr string) error {
	// 绝对路径的前缀长度
	preLen := len(pathstr)
	err := filepath.Walk(pathstr, func(pathstr string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}

		fmt.Println("[Red] handle: ", pathstr)

		fileInfo, err := os.Stat(pathstr)
		if err != nil {
			fmt.Println("[Red] read file stat err! ", err)
		}

		var shareFileInfo ShareFileInfo
		shareFileInfo.Path = filepath.Dir(pathstr)[preLen:]
		shareFileInfo.Name = fileInfo.Name()
		shareFileInfo.Size = fileInfo.Size()
		data, _ := json.Marshal(shareFileInfo)

		length_byte := make([]byte, 4)
		binary.BigEndian.PutUint32(length_byte, uint32(len(data)))
		_, err = conn.Write(length_byte) // 发送json的数据长度
		if err != nil {
			conn.Close()
			fmt.Println("[Red] socket close: ", err)
			return err
		}
		_, err = conn.Write(data) // 发送json内容
		if err != nil {
			conn.Close()
			fmt.Println("[Red] socket close: ", err)
			return err
		}

		// 发送文件
		file, err := os.Open(pathstr)
		defer file.Close()
		if err != nil {
			fmt.Println("[Red] file open err! ", err)
		}
		buf := make([]byte, 1024*1024)
		for {
			n, _ := file.Read(buf)
			if n == 0 {
				break
			}
			_, err = conn.Write(buf[:n])
			if err != nil {
				conn.Close()
				fmt.Println("[Red] socket close: ", err)
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}

func main() {
	dirpath := "."
	port := "7000"

	if len(os.Args) > 1 {
		dirpath = strings.TrimSpace(os.Args[1])
	}

	//命令行可以指定端口，防止端口被占用
	if len(os.Args) > 2 {
		port = os.Args[2]
	}

	addr, err := net.ResolveTCPAddr("tcp4", ":"+port)
	checkerr(err)
	listener, err := net.ListenTCP("tcp4", addr)
	checkerr(err)
	abspath, err := filepath.Abs(".")
	checkerr(err)
	fmt.Println("[Red] share... ", abspath)
	fmt.Println("[Red] start listen... ", addr)
	for {
		conn, err := listener.Accept()
		// accept发生错误时，保证服务器不受影响
		if err != nil {
			continue
		}
		go shareDir(conn, dirpath)
	}
}
