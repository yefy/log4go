package main

import (
	"errors"
	"fmt"
	"github.com/yefy/log4go/ee"
	"github.com/yefy/log4go/log4"
)

func err() error {
	return errors.New("1111")
}

func err1() error {
	//err := err()
	//return ee.New(err, "2222")
	return ee.New(nil, "2222")
}

func err2() error {
	err := err1()
	return ee.New(err, "")
}

func err3() error {
	err := err2()
	return ee.New(err, "4444")
}

func err4() error {
	return err3()
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			log4.Recover(r)
		}
	}()

	err := doMain()
	if err != nil {
		fmt.Printf("err:%v\n", err)
		log4.Error("err:%v", err)
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

	err = err4()
	if err != nil {
		fmt.Printf("err:%v\n", err)
		fmt.Printf("err:%+v\n", err)
		fmt.Printf("err:%+#v\n", err)

		log4.Error("err:%v", err)
		log4.Error("err:%+v", err)
		log4.Error("err:%+#v", err)
	}
	return nil
}
