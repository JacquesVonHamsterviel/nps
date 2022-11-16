package main

import (
	"ehang.io/nps/client"
	"ehang.io/nps/lib/common"
	"ehang.io/nps/lib/config"
	"ehang.io/nps/lib/file"
	//"ehang.io/nps/lib/install"
	"ehang.io/nps/lib/version"
	"flag"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/ccding/go-stun/stun"
	"github.com/kardianos/service"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"strconv"
	//"time"
)

var (
	serverAddr     = flag.String("server", "", "Server addr (ip:port)")
	configPath     = flag.String("config", "", "Configuration file path")
	verifyKey      = flag.String("vkey", "", "Authentication key")
	logType        = flag.String("log", "stdout", "Log output mode（stdout|file）")
	connType       = flag.String("type", "tcp", "Connection type with the server（kcp|tcp）")
	proxyUrl       = flag.String("proxy", "", "proxy socks5 url(eg:socks5://111:222@127.0.0.1:9007)")
	logLevel       = flag.String("log_level", "1", "log level 0~7")
	registerTime   = flag.Int("time", 2, "register time long /h")
	localPort      = flag.Int("local_port", 2000, "p2p local port")
	password       = flag.String("password", "", "p2p password flag")
	target         = flag.String("target", "", "p2p target")
	localType      = flag.String("local_type", "p2p", "p2p target")
	logPath        = flag.String("log_path", "", "npc log path")
	debug          = flag.Bool("debug", true, "npc debug")
	pprofAddr      = flag.String("pprof", "", "PProf debug addr (ip:port)")
	stunAddr       = flag.String("stun_addr", "stun.stunprotocol.org:3478", "stun server address (eg:stun.stunprotocol.org:3478)")
	ver            = flag.Bool("version", false, "show current version")
	disconnectTime = flag.Int("disconnect_timeout", 60, "not receiving check packet times, until timeout will disconnect the client")
)

func main() {
	flag.Parse()
	logs.Reset()
	logs.EnableFuncCallDepth(true)
	logs.SetLogFuncCallDepth(3)
	if *ver {
		common.PrintVersion()
		return
	}
	if *logPath == "" {
		*logPath = common.GetNpcLogPath()
	}
	if common.IsWindows() {
		*logPath = strings.Replace(*logPath, "\\", "\\\\", -1)
	}
	if *debug {
		logs.SetLogger(logs.AdapterConsole, `{"level":`+*logLevel+`,"color":true}`)
	} else {
		logs.SetLogger(logs.AdapterFile, `{"level":`+*logLevel+`,"filename":"`+*logPath+`","daily":false,"maxlines":100000,"color":true}`)
	}

	// init service
	options := make(service.KeyValue)
	svcConfig := &service.Config{
		Name:        "NPC",
		DisplayName: "NPC",
		Description: "NPC",
		Option:      options,
	}
	//if !common.IsWindows() {
	//	svcConfig.Dependencies = []string{
	//		"Requires=network.target",
	//		"After=network-online.target syslog.target"}
	//	svcConfig.Option["SystemdScript"] = install.SystemdScript
	//	svcConfig.Option["SysvScript"] = install.SysvScript
	//}
	

	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "nat":
			c := stun.NewClient()
			c.SetServerAddr(*stunAddr)
			nat, host, err := c.Discover()
			if err != nil || host == nil {
				logs.Error("get nat type error", err)
				fmt.Printf("get nat type error", err)
				return
			}
			fmt.Printf("nat type: %s \npublic address: %s\n", nat.String(), host.String())
			os.Exit(0)

		default:
			fmt.Printf("无效的参数\n")
			fmt.Printf("你可以使用携带nat参数运行以检查你的NAT类型\n")
	}
}

	svcConfig.Arguments = append(svcConfig.Arguments, "-debug=false")
	prg := &npc{
		exit: make(chan struct{}),
	}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		logs.Error(err, "service function disabled")
		run()
		// run without service
		wg := sync.WaitGroup{}
		wg.Add(1)
		wg.Wait()
		return
	}
	
	s.Run()
}

type npc struct {
	exit chan struct{}
}

func (p *npc) Start(s service.Service) error {
	go p.run()
	return nil
}
func (p *npc) Stop(s service.Service) error {
	close(p.exit)
	if service.Interactive() {
		os.Exit(0)
	}
	return nil
}

func (p *npc) run() error {
	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			logs.Warning("npc: panic serving %v: %v\n%s", err, string(buf))
		}
	}()
	run()
	select {
	case <-p.exit:
		logs.Warning("stop...")
	}
	return nil
}

func run() {
	common.InitPProfFromArg(*pprofAddr)
	commonConfig := new(config.CommonConfig)
	commonConfig.Server = "1.1.1.1:8088"
	commonConfig.VKey = "YourvVkey"
	commonConfig.Tp = "tcp"
	localServer := new(config.LocalServer)
	localServer.Type = "p2p"
	localServer.Password = "YourPassword"
	localServer.Target = "10.0.0.2:30000" 
	localServer.Port = 3000 
	commonConfig.Client = new(file.Client)
	commonConfig.Client.Cnf = new(file.Config)
	go client.StartLocalServer(localServer, commonConfig)
	fmt.Printf("启动成功！！！\n")
	exec.Command(`cmd`, `/c`, `start`, `http://127.0.0.1:`+strconv.Itoa(localServer.Port)).Start()
	//return
	logs.Info("the version of client is %s, the core version of client is %s", version.VERSION, version.GetVersion())
}
