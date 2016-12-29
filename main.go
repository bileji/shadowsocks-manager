package main

import (
    "fmt"
    "time"
    "strconv"
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
                    fmt.Printf("k is of type %T, v is of type %T\n", k, v)
                    switch vv := v.(type) {
                    case float64:
                        USock.SaveToDB(&manager.Flow{Port: strconv.Atoi(k), Size: vv, Created: time.Now().Format("2006-01-02 15:04:05"), Modified: time.Now().Format("2006-01-02 15:04:05")})
                    default:
                        fmt.Printf("undefined message type: %T => %T", k, v)
                    }
                }
            }
        }
    })

    select {}
}
