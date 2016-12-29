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

        flow := make(map[string]interface{})
        json_str := strings.TrimLeft(string(buffer), "stat: ")
        fmt.Println("`" + json_str + "`")
        err = json.Unmarshal([]byte("`" + json_str + "`"), &flow)
        fmt.Println(err)
        if err != nil {
            for k, v := range flow {
                switch vv := v.(type) {
                case int, int8, int32, int64:
                    fmt.Println(k, "is int", vv)
                case string:
                    fmt.Println(k, "is string", vv)
                default:
                    fmt.Println(k, "====", vv)
                }
            }
        }
    })
    select {}
}
