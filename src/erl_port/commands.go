package port

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

const (
	MSG_TYPE_OUT = byte(0)
	MSG_TYPE_ERR = byte(1)
	CMD_DATA     = byte(0)
	CMD_STOP     = byte(1)
)

type command struct {
	cmd  byte
	data []byte
}

func commandsInput(outPipe io.WriteCloser) {
loop:
	for {
		cmd := readCommand()
		logger.Println("cmd", cmd)

		switch cmd.cmd {
		case CMD_STOP:
			break loop
		case CMD_DATA:
			n, err := outPipe.Write(cmd.data)

			if err != nil {
				logger.Println(err)
				fatal_if(err)
			}

			l := len(cmd.data)

			if n != l {
				fatal_if(fmt.Errorf("forward input: expected to write %d bytes length, got %d", l, n))
			}
		}
	}

	outPipe.Close()
}

func outWriter(wg *sync.WaitGroup, input chan *command) {
	defer wg.Done()

	for msg := range input {
		logger.Println("get msg", msg)
		write(msg.cmd, msg.data)
	}
}

func outForward(wg *sync.WaitGroup, msgType byte, out chan *command, input io.ReadCloser) {
	defer wg.Done()
	data := make([]byte, 4096*2-3)

	for {
		time.Sleep(100 * time.Millisecond)

		n, err := input.Read(data)

		logger.Println("read from", msgType, n, err)

		if n > 0 {
			out <- &command{msgType, data[0:n]}
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			logger.Println(msgType, err)
			fatal_if(err)
		}
	}
}

func readCommand() *command {
	len := readLength()
	logger.Println("input cmd: len", len)

	result := &command{}

	if len > 0 {
		var buf []byte = make([]byte, len)
		_, err := io.ReadFull(os.Stdin, buf)

		if err != nil {
			logger.Println("Unexpected read data", err)
			fatal_if(err)
		}

		result.cmd = buf[0]
		result.data = buf[1:]
	} else {
		result.cmd = CMD_STOP
	}

	return result
}

func readLength() uint16 {
	var b = []byte{0, 0}
	_, err := io.ReadFull(os.Stdin, b)

	if err == io.EOF {
		return 0
	}

	logger.Println("read length bytes: ", b)

	if err != nil {
		logger.Println("Unexpected read length", err)
		fatal_if(err)
	}

	return uint16(b[0])<<8 | uint16(b[1])
}

func write(msgType byte, data []byte) {
	len := len(data) + 1
	var b = []byte{0, 0}
	b[0] = byte(len >> 8 & 0xff)
	b[1] = byte(len & 0xff)
	logger.Println("write", msgType, "len", b, "data", data)
	writeExact(b)
	writeExact([]byte{msgType})
	writeExact(data)
}

func writeExact(data []byte) {
	l := len(data)
	n, err := os.Stdout.Write(data)

	if err != nil {
		logger.Println(err)
		fatal_if(err)
	}

	if n != l {
		fatal_if(fmt.Errorf("write: could not write %d bytes", l))
	}
}
