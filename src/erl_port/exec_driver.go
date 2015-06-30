package port

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"sync"
)

var logger *log.Logger

func Run() {
	wg := &sync.WaitGroup{}
	wwg := &sync.WaitGroup{}
	file, err := os.OpenFile("driver.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0640)
	fatal_if(err)
	logger = log.New(file, "[exec]: ", log.Lmicroseconds|log.Lshortfile)
	flag.Parse()
	args := flag.Args()
	logger.Println(args)
	logger.Println(args[1:])
	cmd := exec.Command(args[0], args[1:]...)

	inputPipe, err := cmd.StdinPipe()
	if err != nil {
		logger.Println(err)
		fatal_if(err)
	}

	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		logger.Println(err)
		fatal_if(err)
	}

	errPipe, err := cmd.StderrPipe()
	if err != nil {
		logger.Println(err)
		fatal_if(err)
	}

	out := make(chan *command)
	go outWriter(wwg, out)

	err = cmd.Start()

	if err != nil {
		logger.Println(err)
		fatal_if(err)
	}

	// between Start() and Wait()
	wg.Add(2)
	go outForward(wg, MSG_TYPE_OUT, out, outPipe)
	go outForward(wg, MSG_TYPE_ERR, out, errPipe)
	go commandsInput(inputPipe)
	wg.Wait()
	close(out)
	wwg.Wait()

	err = cmd.Wait()

	if err != nil {
		logger.Println(err)
		os.Exit(exitStatus(err))
	}
}

func fatal_if(err error) {
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(-1)
	}
}
