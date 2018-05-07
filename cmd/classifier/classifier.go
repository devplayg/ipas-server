package main

import (
	"github.com/devplayg/golibs/secureconfig"
	"github.com/devplayg/ipas-server"
	"github.com/devplayg/ipas-server/classifier"
	log "github.com/sirupsen/logrus"
	"os"
	"runtime"
	"time"
)

const (
	AppName    = "IPAS Classifier"
	AppVersion = "2.0.1805.30201"
)

func main() {
	// CPU 설정
	runtime.GOMAXPROCS(runtime.NumCPU())

	// 옵션 설정
	var (
		version   = ipasserver.CmdFlags.Bool("version", false, "Version")
		debug     = ipasserver.CmdFlags.Bool("debug", false, "Debug")
		verbose   = ipasserver.CmdFlags.Bool("v", false, "Verbose")
		batchSize = ipasserver.CmdFlags.Uint("size", 3, "Batch size")
		interval = ipasserver.CmdFlags.Uint("interval", 2000, "Interval(ms)")
		setConfig = ipasserver.CmdFlags.Bool("config", false, "Edit configurations")
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
			"db.hostname, db.port, db.username, db.password, db.database",
			ipasserver.GetEncryptionKey(),
		)
		return
	}

	// 엔진 시작
	if err := engine.Start(); err != nil {
		log.Fatal(err)
	}
	log.Debug(engine.Config)

	// 데이터베이스 연결
	if err := engine.InitDatabase(); err != nil {
		log.Fatal(err)
	}
	defer engine.DB.Close()

	// 데이터 분류 시작
	clf := classifier.NewClassifier(engine, *batchSize)
	go func() {
		dur := time.Duration(*interval) * time.Millisecond
		for {
			clf.Run()
			time.Sleep(dur)
		}

	}()

	// 종료 시그널 대기
	ipasserver.WaitForSignals()
}