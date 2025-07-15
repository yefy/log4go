package log4

import "github.com/yefy/log4go/ee"

type Level int

const (
	FINE Level = iota
	TRACE
	DEBUG
	INFO
	WARNING
	ERROR
	CRITICAL
)

var (
	levelConfigStrings = [...]string{"fine", "trace", "debug", "info", "warn", "error", "crit"}
)

var levelConfigMap = map[string]Level{
	levelConfigStrings[FINE]:     FINE,
	levelConfigStrings[TRACE]:    TRACE,
	levelConfigStrings[DEBUG]:    DEBUG,
	levelConfigStrings[INFO]:     INFO,
	levelConfigStrings[WARNING]:  WARNING,
	levelConfigStrings[ERROR]:    ERROR,
	levelConfigStrings[CRITICAL]: CRITICAL,
}

var (
	levelFileStrings = [...]string{"FINE", "TRACE", "DEBUG", "INFO", "WARN", "ERROR", "CRIT"}
)

var levelFileMap = map[Level]string{
	FINE:     levelFileStrings[FINE],
	TRACE:    levelFileStrings[TRACE],
	DEBUG:    levelFileStrings[DEBUG],
	INFO:     levelFileStrings[INFO],
	WARNING:  levelFileStrings[WARNING],
	ERROR:    levelFileStrings[ERROR],
	CRITICAL: levelFileStrings[CRITICAL],
}

func LevelNameToLevel(name string) (Level, error) {
	level, ok := levelConfigMap[name]
	if !ok {
		return ERROR, ee.New(nil, "not find levelName:%v, use:%+v", name, levelConfigMap)
	}

	return level, nil
}

func LevelNameToLevelDef(name string) Level {
	level, err := LevelNameToLevel(name)
	if err != nil {
		return ERROR
	}

	return level
}

func LevelToLevelFileName(level Level) string {
	return levelFileMap[level]
}
