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

    defer connector.Close()

    USock := manager.UnixSock{
        Net: "unixgram",
        LSock: "/var/run/manager.sock",
        RSock: "/var/run/shadowsocks-manager.sock",
        Collection: connector.DB("vpn").C("flows"),
    }

    USock.Listen()
    go USock.Ping()

    // todo每1分钟检查流量是否超标
    go USock.HeartBeat(60, func() {
        fmt.Println("this is task to check users")
        return nil
    })

    // 监听各端口流量情况
    go USock.Rec(func(buffer []byte) {
        m := make(map[string]interface{})
        if message := strings.TrimLeft(string(buffer), "stat: "); strings.EqualFold(message, "pong") {
            fmt.Println("start the program: shadowsocks-manager")
        } else {
            if err := json.NewDecoder(strings.NewReader(message)).Decode(&m); err == nil {
                for k, v := range m {
                    switch vv := v.(type) {
                    case float64:
                        port, _ := strconv.Atoi(k)
                        USock.SaveToDB(&manager.Flow{Port: int32(port), Size: vv, Created: time.Now().Format("2006-01-02 15:04:05"), Modified: time.Now().Format("2006-01-02 15:04:05")})
                    default:
                        fmt.Printf("undefined message type: %T => %T", k, v)
                    }
                }
            }
        }
    })

    select {}
}
