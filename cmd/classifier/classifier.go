package main

import (
	"github.com/devplayg/golibs/secureconfig"
	"github.com/devplayg/ipas-server"
	log "github.com/sirupsen/logrus"
	"os"
	"runtime"
	"github.com/devplayg/ipas-server/classifier"
)

const (
	AppName    = "IPAS Sorter"
	AppVersion = "1.0.1803.10801"
)

func main() {
	// CPU 설정
	runtime.GOMAXPROCS(runtime.NumCPU())

	// 옵션 설정
	var (
		version         = ipasserver.CmdFlags.Bool("version", false, "Version")
		debug           = ipasserver.CmdFlags.Bool("debug", false, "Debug")
		verbose         = ipasserver.CmdFlags.Bool("v", false, "Verbose")
		setConfig       = ipasserver.CmdFlags.Bool("config", false, "Edit configurations")
		//batchSize       = ipasserver.CmdFlags.Int("batchsize", 4, "Batch size")
		//batchTimeout    = ipasserver.CmdFlags.Int("batchtime", 5000, "Batch timeout, in milliseconds")
		//batchMaxPending = ipasserver.CmdFlags.Int("maxpending", 4, "Maximum pending events")

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
		log.Error(err)
		return
	}
	log.Debug(engine.Config)

	// Sorter 시작
	clf := classifier.NewClassifier(engine)
	if err := clf.Start(); err != nil {
		log.Error(err)
		return
	}




	//
	//// 데이터 수신기 시작
	//timeout := time.Duration(*batchTimeout) * time.Millisecond
	//dispatcher := receiver.NewDispatcher(*batchSize, timeout, *batchMaxPending, engine)
	//errChan := make(chan error)
	//if err := dispatcher.Start(errChan); err != nil {
	//	log.Fatalf("failed to start indexing batcher: %s", err.Error())
	//}
	//log.Debugf("batching configured with size %d, timeout %s, max pending %d",
	//	*batchSize, timeout, *batchMaxPending)

	//go drainLog("error batch", errChan)

	//// HTTP 라우터 시작
	//router := httprouter.New()
	//if err := startRouters(router, dispatcher); err != nil {
	//	log.Fatalf("failed to start routers: %s")
	//}
	//http.ListenAndServe(":8080", router) // 웹서버 시작

	// 종료 시그널 대기
	ipasserver.WaitForSignals()

	if err := clf.Stop(); err != nil {
		log.Error(err)
		return
	}


}
//
//func drainLog(msg string, errChan <-chan error) {
//	for {
//		select {
//		case err := <-errChan:
//			if err != nil {
//				log.Errorf("%s: %s", msg, err.Error())
//			}
//		}
//	}
//}
