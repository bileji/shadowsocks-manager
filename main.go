package main

import (
    "fmt"
    "time"
    "strconv"
    "strings"
    "encoding/json"
    "shadowsocks-manager/manager"
    "flag"
    "os"
    "shadowsocks-manager/service"
)

//type Limit struct {
//    AllowSize float64
//    Password  string
//}

var (
    MONGODB_HOST = "127.0.0.1:27017"
    MONGODB_DATABASE = "vpn"
    MONGODB_USERNAME = "shadowsocks"
    MONGODB_PASSWORD = "mlgR4evB"

    USER_COLLECTION = "users"
    FLOW_COLLECTION = "flows"

    HEARTBEAT_FREQUENCY = 30
)

func GetArgs() *manager.Options {
    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "Welcome to use %s ^_^____\r\nOptions:\n", os.Args[0])
        flag.PrintDefaults()
    }

    DBHost := flag.String("host", MONGODB_HOST, "mongodb host:port")
    DBName := flag.String("db_name", MONGODB_DATABASE, "db name")
    Username := flag.String("username", MONGODB_USERNAME, "db's username")
    Pwd := flag.String("password", MONGODB_PASSWORD, "db's password")
    Heartbeat := flag.Int("heartbeat", HEARTBEAT_FREQUENCY, "flow detection frequency(sec)")

    flag.Parse()
    return &manager.Options{
        DBHost: *DBHost,
        DBName: *DBName,
        DBUsername: *Username,
        DBPassword: *Pwd,
        HeartbeatFrequency: *Heartbeat,
    }
}

func main() {
    Args := GetArgs()

    err, Con := manager.ConnectToMgo(Args.DBHost, Args.DBName, Args.DBUsername, Args.DBPassword)
    if err != nil {
        panic(err)
    }
    defer Con.Session.Close()

    USock := manager.UnixSock{
        Net: "unixgram",
        LSock: "/var/run/manager.sock",
        RSock: "/var/run/shadowsocks-manager.sock",
        Con: Con,
        Collection: FLOW_COLLECTION,
        Args: Args,
        ListenPorts: manager.New(),
    }

    // 正在监听的端口
    USock.Listen()
    go USock.Ping()

    // 每30sec检查流量是否超标
    go USock.HeartBeat(Args.HeartbeatFrequency, USock.Monitor)

    // 监听各端口流量情况
    go USock.Rec(func(buffer []byte) {
        M := make(map[string]interface{})
        if Message := strings.TrimLeft(string(buffer), "stat: "); strings.Compare(Message, "pong") > 0 {

        } else {
            if err := json.NewDecoder(strings.NewReader(Message)).Decode(&M); err == nil {
                for k, v := range M {
                    switch Size := v.(type) {
                    case float64:
                        Port, _ := strconv.Atoi(k)
                        USock.SaveToDB(&manager.Flow{
                            Port: int32(Port),
                            Size: Size,
                            Created: time.Now().Format("2006-01-02 15:04:05"),
                            Modified: time.Now().Format("2006-01-02 15:04:05"),
                        })
                    default:
                        fmt.Printf("undefined message type: %T => %T\r\n", k, v)
                    }
                }
            }
        }
    })

    // web服务
    Web := service.Web{
        Addr: ":80",
        DBCon: Con,
        OnlinePort: USock.ListenPorts,
    }
    go Web.Run()

    select {}
}
