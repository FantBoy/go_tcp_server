package libs

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net"
	"strings"
	"sync"
)

// 混合类型的struct
type ComplexData struct {
	N int
	S string
	M map[string]int
	P []byte
	C *ComplexData
}

/**
net.Conn 实现了io.Reader  io.Writer  io.Closer接口
Open 返回一个有超时的TCP链接缓冲readwrite
*/
func Open(addr string) (*bufio.ReadWriter, error) {
	// Dial the remote process.
	// Note that the local port is chosen on the fly. If the local port
	// must be a specific one, use DialTCP() instead.
	fmt.Println("Dial " + addr)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, errors.Wrap(err, "Dialing "+addr+" failed")
	}
	return bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)), nil
}

type HandleFunc func(*bufio.ReadWriter)

type TcpServer struct {
	listener net.Listener
	// handlefunc是一个处理传入命令的函数类型。 它接收打包在一个读写器界面中的开放连接。
	handler map[string]HandleFunc

	// map不是线程安全的，所以需要读写锁控制
	m sync.RWMutex
}

func NewTcpServer() *TcpServer {
	return &TcpServer{
		handler: map[string]HandleFunc{},
	}
}

// 添加数据类型处理方法
func (e *TcpServer) AddHandleFunc(name string, f HandleFunc) {
	e.m.Lock()
	e.handler[name] = f
	e.m.Unlock()
}

// 验证请求数据类型，并发送到对应处理函数
func (e *TcpServer) handleMessage(conn net.Conn) {
	rw := bufio.NewReadWriter(bufio.NewReader(conn),
		bufio.NewWriter(conn))
	defer conn.Close()
	for {
		cmd, err := rw.ReadString('\n')
		switch {
		case err == io.EOF:
			fmt.Println("读取完成.")
			return
		case err != nil:
			fmt.Println("读取出错")
			return
		}

		cmd = strings.Trim(cmd, "\n ")
		e.m.RLock()
		handleCmd, ok := e.handler[cmd]
		if !ok {
			fmt.Println("未注册的请求数据类型.")
			return
		}
		//具体处理链接数据
		handleCmd(rw)
	}
}

func (e *TcpServer) Listen(cfg map[string]string) error {
	//var port int
	var err error
	//port, err = strconv.Atoi(cfg["Port"])

	e.listener, err = net.Listen("tcp", cfg["port"])
	if err != nil {
		return errors.Wrap(err, "TCP服务无法监听在端口"+cfg["port"])
	}
	fmt.Println(" 服务监听成功：", e.listener.Addr().String())
	for {
		conn, err := e.listener.Accept()
		if err != nil {
			fmt.Println("新请求监听失败!")
			continue
		}
		// 开始处理新链接数据
		go e.handleMessage(conn)
	}

}

func HandleStrings(rw *bufio.ReadWriter) {
	s, err := rw.ReadString('\n')
	if err != nil {
		fmt.Println("链接无法读取.")
		return
	}

	s = strings.Trim(s, "\n ")
	// ....
	_, err = rw.WriteString("处理完成......\n")
	if err != nil {
		fmt.Println("链接写入响应失败")
		return
	}
	// 写入底层网络链接
	err = rw.Flush()
	if err != nil {
		fmt.Println("Flush写入失败")
		return
	}
}

func HandleGob(rw *bufio.ReadWriter) {
	var data ComplexData

	dec := gob.NewDecoder(rw)
	err := dec.Decode(&data)
	if err != nil {
		fmt.Println("解析的二进制数据出现异常.")
		return
	}
	fmt.Println("输出：", data, data.C)
}
