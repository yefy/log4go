package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/yefy/log4go/ee"
	"github.com/yefy/log4go/log4"
	"os"
	"regexp"
	"strconv"
	"strings"
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

	var count int
	flag.IntVar(&count, "count", 100000, "count")
	flag.Parse()

	if true {
		go func() {
			for {
				time.Sleep(time.Second * 1)
				log4.Reopen()
			}
		}()
	}

	context := log4.NewWaitGroupContext()

	if true {
		context.Add(1)
		go func() {
			defer context.Done()
			for i := 0; i < count; i++ {
				log4.Info("i:%v", i)
			}
		}()
	}

	if true {
		context.Add(1)
		go func() {
			defer context.Done()
			for i := 0; i < count; i++ {
				log4.Target("main").Info("i:%v", i)
			}
		}()
	}

	context.Wait()

	check("root", "./logs/sniffer.log", count)
	check("main", "./logs/sniffer.log", count)
	check("main", "./logs/sniffer_main.log", count)

	return nil
}

func check(target string, path string, count int) {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	re := regexp.MustCompile(`i:(\d+)`)
	seen := make(map[int]bool)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, target) {
			continue
		}
		matches := re.FindStringSubmatch(line)
		if len(matches) == 2 {
			n, err := strconv.Atoi(matches[1])
			if err == nil {
				//fmt.Printf("n:%v\n", n)
				seen[n] = true
			}
		}
	}

	for i := 0; i < count; i++ {
		if !seen[i] {
			fmt.Printf("not find: i:%d, target:%v, path:%v\n", i, target, path)
		}
	}
}
