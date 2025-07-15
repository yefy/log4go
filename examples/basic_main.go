package main

import (
	"fmt"
	"github.com/yefy/log4go/ee"
	"github.com/yefy/log4go/log4"
	"time"
)

func main() {
	err := doMain()
	if err != nil {
		fmt.Printf("err:%v\n", err)
	}
}

func doMain() error {
	err := log4.InitFile("./conf/log4.yaml")
	if err != nil {
		return ee.New(err, "log4.InitFile")
	}

	defer func() {
		log4.Close(true)
	}()

	for {
		log4.Critical("=========================================== start")
		log4.Target("main").Critical("=========================================== start")

		log4.Critical("root Critical")
		log4.Error("rootError")
		log4.Warn("rootWarn")
		log4.Info("rootInfo")
		log4.Debug("rootDebug")
		log4.Trace("rootTrace")
		log4.Fine("rootFine")

		log4.Target("").Critical("nil Target Critical")
		log4.Target("").Error("nil Target Error")
		log4.Target("").Warn("nil Target Warn")
		log4.Target("").Info("nil Target Info")
		log4.Target("").Debug("nil Target Debug")
		log4.Target("").Trace("nil Target Trace")
		log4.Target("").Fine("nil Target Fine")

		log4.Target("main").Critical("main Target Critical")
		log4.Target("main").Error("main Target Error")
		log4.Target("main").Warn("main Target Warn")
		log4.Target("main").Info("main Target Info")
		log4.Target("main").Debug("main Target Debug")
		log4.Target("main").Trace("main Target Trace")
		log4.Target("main").Fine("main Target Fine")

		log4.Target("not_find").Critical("not_find Target Critical")
		log4.Target("not_find").Error("not_find Target Error")
		log4.Target("not_find").Warn("not_find Target Warn")
		log4.Target("not_find").Info("not_find Target Info")
		log4.Target("not_find").Debug("not_find Target Debug")
		log4.Target("not_find").Trace("not_find Target Trace")
		log4.Target("not_find").Fine("not_find Target Fine")

		log4.Target("test").Critical("test Target Critical")
		log4.Target("test").Error("test Target Error")
		log4.Target("test").Warn("test Target Warn")
		log4.Target("test").Info("test Target Info")
		log4.Target("test").Debug("test Target Debug")
		log4.Target("test").Trace("test Target Trace")
		log4.Target("test").Fine("test Target Fine")

		if log4.GetLevel() >= log4.INFO {
			log4.Critical("root Level Critical")
			log4.Error("root Level Error")
			log4.Warn("root Level Warn")
			log4.Info("root Level Info")
		}

		log4.Critical("=========================================== end")
		log4.Target("main").Critical("=========================================== end")

		time.Sleep(time.Second * 3)
	}

	return nil
}
