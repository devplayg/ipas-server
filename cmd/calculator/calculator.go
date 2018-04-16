package main

import (
	"github.com/devplayg/golibs/secureconfig"
	"github.com/devplayg/ipas-server"
	log "github.com/sirupsen/logrus"
	"os"
	"runtime"
)

const (
	AppName    = "IPAS Statistics Calculator"
	AppVersion = "1.0.1804.11501"
)

func main() {
	// CPU 설정
	runtime.GOMAXPROCS(2)

	// 옵션 설정
	var (
		version   = ipasserver.CmdFlags.Bool("version", false, "Version")
		debug     = ipasserver.CmdFlags.Bool("debug", false, "Debug")
		verbose   = ipasserver.CmdFlags.Bool("v", false, "Verbose")
		setConfig = ipasserver.CmdFlags.Bool("config", false, "Edit configurations")
		//worker    = ipasserver.CmdFlags.Int("worker", runtime.NumCPU(), "Worker count")
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

	// 데이터 분류 시작
	//statusCalculator := calculator.NewCalculator(engine, "status")
	//clf := classifier.NewClassifier(engine, *worker)
	//if err := statusCalculator.Start(); err != nil {
	//	log.Fatal(err)
	//}
	//defer clf.Stop()

	// 종료 시그널 대기
	ipasserver.WaitForSignals()
}