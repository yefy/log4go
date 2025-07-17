package log4

import (
	"github.com/yefy/log4go/ee"
	"github.com/yefy/log4go/efile"
)

//go:generate gomodifytags -file log4_config.go -struct Log4Config -add-tags yaml -transform snakecase -w
type Log4Config struct {
	RefreshRate int64                         `yaml:"refresh_rate"`
	Appenders   map[string]Log4ConfigAppender `yaml:"appenders"`
	Root        Log4ConfigLogger              `yaml:"root"`
	Loggers     map[string]Log4ConfigLogger   `yaml:"loggers"`
}

func (log4Config *Log4Config) Check() error {
	appenders := make([]string, 0, len(log4Config.Appenders))
	for appender, v := range log4Config.Appenders {
		appenders = append(appenders, appender)
		if v.Kind != KindConsole && v.Kind != KindFile {
			return ee.New(nil, "not find kind:%v, use:%+v|%+v in appenders:%v|%+v", v.Kind, KindConsole, KindFile, appender, v)
		}

		if v.Kind == KindFile {
			if len(v.Path) <= 0 {
				return ee.New(nil, "open path nil in appenders:%v|%+v", v.Path, appender, v)
			}

			err := efile.EnsureLogDirExists(v.Path)
			if err != nil {
				return ee.New(err, "create path:%v in appenders:%v|%+v", v.Path, appender, v)
			}

			file, err := efile.OpenFileWithShareDelete(v.Path)
			if err != nil {
				return ee.New(err, "open path:%v in appenders:%v|%+v", v.Path, appender, v)
			}
			defer file.Close()
		}
	}

	{
		_, err := LevelNameToLevel(log4Config.Root.Level)
		if err != nil {
			return ee.New(err, "LevelNameToLevel in root:%+v", log4Config.Root)
		}

		for _, appender := range log4Config.Root.Appenders {
			_, ok := log4Config.Appenders[appender]
			if !ok {
				return ee.New(nil, "not find appender:%v, use:%+v in root:%+v", appender, appenders, log4Config.Root)
			}
		}
	}

	for k, v := range log4Config.Loggers {
		if k == defaultRootTarget {
			return ee.New(nil, "name == root in loggers:%v|%+v\"", k, v)
		}

		if k == defaultDiscardTarget {
			return ee.New(nil, "name == discard_root in loggers:%v|%+v\"", k, v)
		}
		_, err := LevelNameToLevel(v.Level)
		if err != nil {
			return ee.New(err, "LevelNameToLevel in loggers:%v|%+v\"", k, v)
		}

		for _, appender := range v.Appenders {
			_, ok := log4Config.Appenders[appender]
			if !ok {
				return ee.New(nil, "not find appender:%v, use:%+v in loggers:%v|%+v", appender, appenders, k, v)
			}
		}
	}
	return nil
}

const KindConsole = "console"
const KindFile = "file"

//go:generate gomodifytags -file log4_config.go -struct Log4ConfigAppender -add-tags yaml -transform snakecase -w
type Log4ConfigAppender struct {
	Kind    string `yaml:"kind"`
	Pattern string `yaml:"pattern"`
	Path    string `yaml:"path"`
}

//go:generate gomodifytags -file log4_config.go -struct Log4ConfigLogger -add-tags yaml -transform snakecase -w
type Log4ConfigLogger struct {
	Level     string   `yaml:"level"`
	Multiline bool     `yaml:"multiline"`
	Additive  bool     `yaml:"additive"`
	Appenders []string `yaml:"appenders"`
}
