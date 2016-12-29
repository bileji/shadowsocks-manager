package main

import (
    "fmt"
    "shadowsocks-manager/manager"
)

func main() {
    USock := manager.UnixSock{
        Net: "unixgram",
        LSock: "/var/run/manager.sock",
        RSock: "/var/run/shadowsocks-manager.sock",
    }

    USock.Listen()
    go USock.Ping()
    go USock.Rec(func(buffer []byte) {
        fmt.Println(string(buffer))
    })
    select {}
}
