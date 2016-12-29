package manager

import (
    "os"
    "net"
    "strconv"
)

type UnixSock struct {
    Net   string
    LSock string
    RSock string
    UConn *net.UnixConn
}

func (us *UnixSock) Listen() {
    os.Remove(us.LSock)
    rAddr, err := net.ResolveUnixAddr(us.Net, us.RSock)
    if err != nil {
        panic(err)
    }
    lAddr, err := net.ResolveUnixAddr(us.Net, us.LSock)
    if err != nil {
        panic(err)
    }
    us.UConn, err = net.DialUnix(us.Net, lAddr, rAddr)
    if err != nil {
        panic(err)
    }
}

func (us *UnixSock) Ping() (int int, err error) {
    int, err = us.UConn.Write([]byte("ping"))
    return
}

func (us *UnixSock) Add(port int32, pwd string) (int int, err error) {
    str := "add: {\"server_port\": " + strconv.FormatInt(int64(port), 10) + ", \"password\": \"" + pwd + "\"}"
    int, err = us.UConn.Write([]byte(str))
    return
}

func (us *UnixSock) Del(port int32) (int int, err error) {
    str := "remove: {\"server_port\": " + strconv.FormatInt(int64(port), 10) + "}"
    int, err = us.UConn.Write([]byte(str))
    return
}

func (us *UnixSock) Rec(fn func(res []byte)) {
    for {
        buffer := make([]byte, 128)
        _, err := us.UConn.Read(buffer)
        if err == nil {
            fn(buffer)
        }
    }
}