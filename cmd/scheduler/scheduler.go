package main

import (
	"github.com/devplayg/golibs/secureconfig"
	"github.com/devplayg/ipas-server"
	"github.com/devplayg/ipas-server/sysmonitor"
	"github.com/jasonlvhit/gocron"
	log "github.com/sirupsen/logrus"
	"os"
	"runtime"
)

const (
	AppName    = "System resource monitor"
	AppVersion = "2.0.1804.32201"
)

func main() {

	runtime.GOMAXPROCS(1)

	// 옵션 설정
	var (
		version = ipasserver.CmdFlags.Bool("version", false, "Version")
		//interval = ipasserver.CmdFlags.Duration("interval", 10*time.Second, "Interval(sec)")
		debug     = ipasserver.CmdFlags.Bool("debug", true, "Debug")
		verbose   = ipasserver.CmdFlags.Bool("v", false, "Verbose")
		setConfig = ipasserver.CmdFlags.Bool("config", false, "Edit configurations")
		disk      = ipasserver.CmdFlags.String("disk", "/home", "Disk to monitor")
	)
	ipasserver.CmdFlags.Usage = ipasserver.PrintHelp
	ipasserver.CmdFlags.Parse(os.Args[1:])

	// 버전 출력
	if *version {
		ipasserver.DisplayVersion(AppName, AppVersion)
		return
	}

	// 엔진 설정
	engine := ipasserver.NewEngine(AppName, *debug, *verbose)
	if *setConfig {
		secureconfig.SetConfig(
			engine.ConfigPath,
			"db.hostname, db.port, db.username, db.password, db.database, timezone",
			ipasserver.GetEncryptionKey(),
		)
		return
	}

	// 엔진 시작
	if err := engine.Start(); err != nil {
		log.Fatal(err)
	}

	// 데이터베이스 연결
	if err := engine.InitDatabase(1, 1); err != nil {
		log.Fatal(err)
	}
	defer engine.DB.Close()

	sysmonitor.UpdateResource(engine, *disk)
	gocron.Every(1).Minute().Do(sysmonitor.UpdateResource, engine, *disk)
	go func() {
		<-gocron.Start()
	}()

	// 종료 시그널 대기
	ipasserver.WaitForSignals()
}
