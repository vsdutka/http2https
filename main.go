// main
package main

//go:generate C:\!Dev\GOPATH\src\github.com\vsdutka\gover\gover.exe
import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"

	"github.com/kardianos/service"
	_ "golang.org/x/tools/go/ssa"
)

var (
	logger              service.Logger
	loggerLock          sync.Mutex
	verFlag             *bool
	svcFlag             *string
	listenPortFlag      *int
	destHostFlag        *string
	destPortFlag        *int
	destCertFlag        *string
	destKeyFlag         *string
	destKeyPassFlag     *string
	confServiceName     string
	confServiceDispName string
	debugFlag           *bool
)

func logInfof(format string, a ...interface{}) error {
	loggerLock.Lock()
	defer loggerLock.Unlock()
	if logger != nil {
		return logger.Infof(format, a...)
	}
	return nil
}
func logError(v ...interface{}) error {
	loggerLock.Lock()
	defer loggerLock.Unlock()
	if logger != nil {
		return logger.Error(v)
	}
	return nil
}

// Program structures.
//  Define Start and Stop methods.
type program struct {
	exit chan struct{}
}

func (p *program) Start(s service.Service) error {
	if service.Interactive() {
		logInfof("Service \"%s\" is running in terminal.", confServiceDispName)
	} else {
		logInfof("Service \"%s\" is running under service manager.", confServiceDispName)
	}
	p.exit = make(chan struct{})

	// Start should not block. Do the actual work async.
	go p.run()
	return nil
}
func (p *program) run() {
	startServer()
	logInfof("Service \"%s\" is started.", confServiceDispName)
	for {
		select {
		case <-p.exit:
			return
		}
	}
}
func (p *program) Stop(s service.Service) error {
	// Any work in Stop should be quick, usually a few seconds at most.
	logInfof("Service \"%s\" is stopping.", confServiceDispName)
	stopServer()
	logInfof("Service \"%s\" is stopped.", confServiceDispName)
	close(p.exit)
	return nil
}

// Service setup.
//   Define service config.
//   Create the service.
//   Setup the logger.
//   Handle service controls (optional).
//   Run the service.
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Usage = usage
	verFlag = flag.Bool("version", false, "Show version")
	svcFlag = flag.String("service", "", fmt.Sprintf("Control the system service. Valid actions: %q\n", service.ControlAction))
	listenPortFlag = flag.Int("listen_port", 13777, "Listening port")
	destHostFlag = flag.String("dest_host", "", "Destination host name")
	destPortFlag = flag.Int("dest_port", 443, "Destination port")
	destCertFlag = flag.String("dest_cert", "", "Destination certificate file name")
	destKeyFlag = flag.String("dest_key", "", "Destination key file name")
	destKeyPassFlag = flag.String("dest_key_pass", "", "Destination key password")
	debugFlag = flag.Bool("debug", false, "Dump request/response")
	flag.Parse()

	if *verFlag == true {
		fmt.Println("Version: ", VERSION)
		fmt.Println("Build:   ", BUILD_DATE)
		os.Exit(0)
	}

	if *destHostFlag == "" {
		usage()
		os.Exit(2)
	}

	confServiceName = fmt.Sprintf("http2https_%v", *listenPortFlag)
	confServiceDispName = fmt.Sprintf("%s for \"%s:%v\"", serviceDisplayName, *destHostFlag, *destPortFlag)

	svcConfig := &service.Config{
		Name:        confServiceName,
		DisplayName: confServiceDispName,
		Description: confServiceDispName,
		Arguments: []string{
			fmt.Sprintf("-listen_port=%v", *listenPortFlag),
			fmt.Sprintf("-dest_host=%s", *destHostFlag),
			fmt.Sprintf("-dest_port=%v", *destPortFlag),
			fmt.Sprintf("-dest_cert=%s", *destCertFlag),
			fmt.Sprintf("-dest_key=%s", *destKeyFlag),
			fmt.Sprintf("-dest_key_pass=%s", *destKeyPassFlag),
		},
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	errs := make(chan error, 5)
	func() {
		loggerLock.Lock()
		defer loggerLock.Unlock()
		logger, err = s.Logger(errs)
		if err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		for {
			err := <-errs
			if err != nil {
				log.Print(err)
			}
		}
	}()

	if len(*svcFlag) != 0 {
		err := service.Control(s, *svcFlag)
		if err != nil {
			log.Printf("Valid actions: %q\n", service.ControlAction)
			log.Fatal(err)
		}
		return
	}
	err = s.Run()
	if err != nil {
		logError(err)
	}
}

const serviceDisplayName = `HTTP to HTTPS tunneling service`
const usageTemplate = `Http2HttpS is ` + serviceDisplayName + `

Usage: http2https commands

The commands are:
`

func usage() {
	fmt.Fprintln(os.Stderr, usageTemplate)
	flag.PrintDefaults()
}
