package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"time"
)

func dumpRequest(req *http.Request) string {
	s := ""
	if *debugFlag {
		dump, err := httputil.DumpRequest(req, true)
		s = "Request:\n>>>>>>>>>>>>>>>>>\n" + string(dump) + "\n"
		if err != nil {
			s = s + "\n" + "Request Dump Error: " + err.Error() + "\n"
		}
		s = s + "<<<<<<<<<<<<<<<<<\n"
	}
	return s
}
func dumpResponse(reqDump string, resp *http.Response) {
	if *debugFlag {
		s := reqDump
		dump, err := httputil.DumpResponse(resp, true)
		s = s + "\nResponse:\n>>>>>>>>>>>>>>>>>\n" + string(dump) + "\n"
		if err != nil {
			s = s + "\n" + "Response Dump Error: " + err.Error() + "\n"
		}
		s = s + "<<<<<<<<<<<<<<<<<\n"

		filename := fmt.Sprintf("%s\\log\\%s.dump", basePath, time.Now().Format("2006-01-02T15-04-05-999999999Z07-00"))
		_ = ioutil.WriteFile(filename, []byte(s), 0644)
	}
}
