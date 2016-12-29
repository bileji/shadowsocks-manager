package main

import (
    "fmt"
    "strings"
    "encoding/json"
    "shadowsocks-manager/manager"
)

func main() {
    err, connector := manager.ConnectToMgo("127.0.0.1", "vpn", "shadowsocks", "mlgR4evB")

    if err != nil {
        panic(err)
    }

    USock := manager.UnixSock{
        Net: "unixgram",
        LSock: "/var/run/manager.sock",
        RSock: "/var/run/shadowsocks-manager.sock",
        Collection: connector.C("flows"),
    }

    USock.Listen()
    go USock.Ping()

    // todo每2分钟检查流量是否超标

    go USock.Rec(func(buffer []byte) {
        m := make(map[string]interface{})
        if message := strings.TrimLeft(string(buffer), "stat: "); message == "pong" {
            fmt.Println("start to listen shadowsocks flow...")
        } else {
            if err := json.NewDecoder(strings.NewReader(message)).Decode(&m); err == nil {
                for k, v := range m {
                    switch vv := v.(type) {
                    case int, int8, int32, int64:
                        fmt.Println(k, "is int", vv)
                    case uint, uint8, uint16, uint32, uint64, uintptr:
                        fmt.Println(k, "is uint", vv)
                    case string:
                        fmt.Println(k, "is string", vv)
                    default:
                        fmt.Println(k, "====", vv)
                    }
                }
            }
        }
    })

    select {}
}
