package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	list_mutex  *sync.Mutex
	onlinecount int = 0
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("the argument is fail the formate is : dataforword.exe localport remoteaddr")
		fmt.Println("example: c:>dataforword.exe 8000 192.168.1.100:10000")
		os.Exit(1)
	}
	listenport := ":" + os.Args[1]
	localaddr, err := net.ResolveTCPAddr("tcp4", listenport)
	if err != nil {
		fmt.Println("Can't covert the ", os.Args[1], " to a correct ip address ,it is a short number")
		os.Exit(1)
	}
	remoteaddrstr := os.Args[2]
	remoteaddr, err := net.ResolveTCPAddr("tcp4", remoteaddrstr)
	if err != nil {
		fmt.Println("Can't covert the ", remoteaddrstr, " to a correct ip address", err)
		fmt.Println("Example c:>dataforword.exe 8000 192.168.2.100:10000")
		return
	}
	listener, err := net.ListenTCP("tcp", localaddr)
	if err != nil {
		fmt.Println("listen the port :", localaddr.Port, " error ,", err.Error())
		os.Exit(1)
	}
	fmt.Printf("listen on local port:%d forward data to :%s \n", localaddr.Port, remoteaddr.String())
	list_mutex = new(sync.Mutex)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("recv a new connection error,", err)
				continue
			}
			remoteconn, err := net.DialTimeout("tcp", remoteaddrstr, time.Second*10)
			if err != nil {
				fmt.Println("dial remote addr fail")
				conn.Close()
				continue
			}
			list_mutex.Lock()
			onlinecount++
			list_mutex.Unlock()
			fmt.Printf("new connection from %s online connection count:%d\n", conn.RemoteAddr().String(), onlinecount)
			go forward_data(conn, remoteconn)
			go forward_data(remoteconn, conn)
		}
	}()

	fmt.Print(">>")
	scaner := bufio.NewScanner(os.Stdin)
	for scaner.Scan() {
		cmd := scaner.Text()
		cmds := strings.Split(cmd, " ")
		switch cmds[0] {
		case "quit":
			return
		case "help":
			fmt.Println("quit        quit the programe")
			fmt.Println("list        show the online count")
		case "list":
			fmt.Println("online:", onlinecount)
		}
		fmt.Print(">>")
	}
}

func forward_data(src net.Conn, dest net.Conn) {
	defer func() {
		err := src.Close()
		dest.Close()
		if err == nil {
			fmt.Printf("the connection from :%s is closed\n", src.RemoteAddr().String())
			list_mutex.Lock()
			onlinecount--
			list_mutex.Unlock()
		}
	}()
	var buf [1024]byte
	for {
		n, err := src.Read(buf[0:])
		if err != nil {
			return
		}
		fmt.Printf("forward:% 02X\n", buf[0:n])
		_, err1 := dest.Write(buf[0:n])
		if err1 != nil {
			return
		}
	}
}
