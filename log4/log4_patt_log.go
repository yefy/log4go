// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package log4

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

const (
	FORMAT_DEFAULT  = "[%D %T] [%L] (%S) %M"
	FORMAT_SHORT    = "[%t %d] [%L] %M"
	FORMAT_ABBREV   = "[%L] %M"
	FORMAT_TIME_UTC = "%U"
)

type formatCacheType struct {
	LastUpdateSeconds    int64
	shortTime, shortDate string
	longTime, longDate   string
}

// Known format codes:
// %T - Time (15:04:05 MST)
// %t - Time (15:04)
// %D - Date (2006-01-02)
// %d - Date (01-02-06)
// %L - Level (FINE, DEBG, TRAC, WARN, ERROR, CRIT)
// %S - Source
// %M - Message
// Ignores unknown formats
// Recommended: "[%D %T] [%L] (%S) %M"
// %U = utc
func FormatLogRecord(format string, isUtc bool, rec *Log4Record, formatCache *formatCacheType) string {
	if rec == nil {
		return ""
	}
	if len(format) == 0 {
		return ""
	}

	out := bytes.NewBuffer(make([]byte, 0, 64))
	Created := rec.GetCreateTime(isUtc)
	secs := Created.UnixNano() / 1e9

	cache := *formatCache
	if cache.LastUpdateSeconds != secs {
		month, day, year := Created.Month(), Created.Day(), Created.Year()
		hour, minute, second, millisecond := Created.Hour(), Created.Minute(), Created.Second(), Created.Nanosecond()/1000000

		zone, _ := Created.Zone()
		updated := &formatCacheType{
			LastUpdateSeconds: secs,
			shortTime:         fmt.Sprintf("%02d:%02d", hour, minute),
			shortDate:         fmt.Sprintf("%02d-%02d-%02d", day, month, year%100),
			longTime:          fmt.Sprintf("%02d:%02d:%02d.%03d %s", hour, minute, second, millisecond, zone),
			longDate:          fmt.Sprintf("%04d-%02d-%02d", year, month, day),
		}
		cache = *updated
		formatCache = updated

	}
	//custom format datetime pattern %D{2006-01-02T15:04:05}
	formatByte := changeDttmFormat(format, isUtc, rec)
	// Split the string into pieces by % signs
	pieces := bytes.Split(formatByte, []byte{'%'})

	// Iterate over the pieces, replacing known formats
	for i, piece := range pieces {
		if i > 0 && len(piece) > 0 {
			isFindUtc := false
			switch piece[0] {
			case 'T':
				out.WriteString(cache.longTime)
			case 't':
				out.WriteString(cache.shortTime)
			case 'D':
				out.WriteString(cache.longDate)
			case 'd':
				out.WriteString(cache.shortDate)
			case 'L':
				out.WriteString(rec.Level)
			case 'S':
				out.WriteString(rec.Source)
			case 'U':
				isFindUtc = true
			case 's':
				slice := strings.Split(rec.Source, "/")
				out.WriteString(slice[len(slice)-1])
			case 'M':
				out.WriteString(rec.Message)
			case 'C':
				out.WriteString(rec.Target)
			}
			if isFindUtc {
				if len(piece) > 1 {
					piece := piece[1:]
					shipSpaceCount := 0
					for i := 0; i < len(piece); i++ {
						if piece[i] == ' ' {
							shipSpaceCount += 1
						} else {
							break
						}
					}
					out.Write(piece[shipSpaceCount:])
				}
			} else {
				if len(piece) > 1 {
					out.Write(piece[1:])
				}
			}
		} else if len(piece) > 0 {
			out.Write(piece)
		}
	}
	out.WriteByte('\n')

	return out.String()
}

func changeDttmFormat(format string, isUtc bool, rec *Log4Record) []byte {
	Created := rec.GetCreateTime(isUtc)
	formatByte := []byte(format)
	r := regexp.MustCompile("\\%D\\{(.*?)\\}")
	i := 0
	formatByte = r.ReplaceAllFunc(formatByte, func(s []byte) []byte {
		if i < 2 {
			i++
			str := string(s)
			str = strings.Replace(str, "%D", "", -1)
			str = strings.Replace(str, "{", "", -1)
			str = strings.Replace(str, "}", "", -1)
			return []byte(Created.Format(str))
		}
		return s
	})
	return formatByte
}
