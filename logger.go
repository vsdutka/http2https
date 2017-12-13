// logger
package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kardianos/osext"
)

type statusWriter struct {
	http.ResponseWriter
	status int
	length int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	w.length = len(b)
	return w.ResponseWriter.Write(b)
}

var logChan = make(chan string, 10000)
var basePath string

func init() {
	go func() {
		const fmtFileName = "${app_path}\\log\\ex${date}.log"
		var (
			lastLogging = time.Time{}
			logFile     *os.File
			err         error
			str         string
		)
		defer func() {
			if logFile != nil {
				logFile.Close()
			}
		}()

		basePath = ""
		exeName, err := osext.Executable()

		if err == nil {
			exeName, err = filepath.Abs(exeName)
			if err == nil {
				basePath = filepath.Dir(exeName)
			}
		}

		for {
			select {
			case str = <-logChan:
				{
					if lastLogging.Format("2006_01_02") != time.Now().Format("2006_01_02") {
						if logFile != nil {
							logFile.Close()
						}
						fileName := os.Expand(fmtFileName, func(key string) string {
							switch strings.ToUpper(key) {
							case "APP_PATH":
								return basePath
							case "DATE":
								return time.Now().Format("2006_01_02")
							default:
								return ""
							}
						})
						dir, _ := filepath.Split(fileName)
						os.MkdirAll(dir, os.ModeDir)

						logFile, err = os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
						if err != nil {
							log.Fatalln(err)
						}
					}
					lastLogging = time.Now()
					logFile.WriteString(str)
				}
			}
		}
	}()
}
func writeToLog(msg string) {
	logChan <- msg
}
