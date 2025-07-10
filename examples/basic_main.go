package main

import (
	"fmt"
	"log4go/ee"
	"log4go/log4"
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

		log4.Target("").Critical("discard Target Critical")
		log4.Target("").Error("discard Target Error")
		log4.Target("").Warn("discard Target Warn")
		log4.Target("").Info("discard Target Info")
		log4.Target("").Debug("discard Target Debug")
		log4.Target("").Trace("discard Target Trace")
		log4.Target("").Fine("discard Target Fine")

		log4.Target("main").Critical("main Target Critical")
		log4.Target("main").Error("main Target Error")
		log4.Target("main").Warn("main Target Warn")
		log4.Target("main").Info("main Target Info")
		log4.Target("main").Debug("main Target Debug")
		log4.Target("main").Trace("main Target Trace")
		log4.Target("main").Fine("main Target Fine")

		log4.Target("not_find").Critical("not_find discard Target Critical")
		log4.Target("not_find").Error("not_find discard Target Error")
		log4.Target("not_find").Warn("not_find discard Target Warn")
		log4.Target("not_find").Info("not_find discard Target Info")
		log4.Target("not_find").Debug("not_find discard Target Debug")
		log4.Target("not_find").Trace("not_find discard Target Trace")
		log4.Target("not_find").Fine("not_find discard Target Fine")

		if log4.GetLevel() >= log4.ERROR {
			log4.Critical("root Level Critical")
			log4.Error("root Level Error")
		}

		log4.Critical("=========================================== end")
		log4.Target("main").Critical("=========================================== end")

		time.Sleep(time.Second * 3)
	}

	return nil
}
