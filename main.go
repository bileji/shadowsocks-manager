package main

import (
    "os"
    "fmt"
    "flag"
    "shadowsocks-manager/service"
    "shadowsocks-manager/manager"
)

var (
    MONGODB_HOST = "127.0.0.1:27017"
    MONGODB_DATABASE = "vpn"
    MONGODB_USERNAME = "shadowsocks"
    MONGODB_PASSWORD = "mlgR4evB"

    WEB_ADDR = ":80"
    WEB_SECRET = "mlgR4evB"

    FLOW_COLLECTION = "flows"
    USER_COLLECTION = "users"

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

    WebAddr := flag.String("web_addr", WEB_ADDR, "web addr")
    WebSecret := flag.String("web_secret", WEB_SECRET, "admin secret")

    flag.Parse()
    return &manager.Options{
        DBHost: *DBHost,
        DBName: *DBName,
        DBUsername: *Username,
        DBPassword: *Pwd,
        HeartbeatFrequency: *Heartbeat,
        WebAddr: *WebAddr,
        WebSecret: *WebSecret,
    }
}

func main() {
    Args := GetArgs()

    USock := manager.UnixSock{
        Net: "unixgram",
        LSock: "/var/run/manager.sock",
        RSock: "/var/run/shadowsocks-manager.sock",
        Con: manager.ConnectToMgo(Args.DBHost, Args.DBName, Args.DBUsername, Args.DBPassword),
        Args: Args,
        FlowC: FLOW_COLLECTION,
        UserC: USER_COLLECTION,
        ListenPorts: manager.New(),
    }
    defer USock.Con.Session.Close()

    // 正在监听的端口
    USock.Listen()
    go USock.Ping()

    // 监听各端口流量情况
    go USock.Rec(USock.SaveToDB)

    // 每30sec检查流量是否超标
    USock.Monitor()
    go USock.HeartBeat(USock.Args.HeartbeatFrequency, USock.Monitor)

    // web服务
    go service.Web{
        Addr: USock.Args.WebAddr,
        DBCon: USock.Con,
        OnlinePort: USock.ListenPorts,
        Secret: USock.Args.WebSecret,
    }.Run()

    select {}
}
