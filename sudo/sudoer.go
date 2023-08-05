package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/moqsien/sudo/runner"
	"github.com/moqsien/sudo/utils"
	"github.com/pochard/commons/randstr"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("sudo: error: no arguments")
		return
	}

	if runner.SocketPath == "" {
		runner.SocketPath = utils.FormatSocketPath(fmt.Sprintf("sudo_%s_sock", randstr.RandomAlphanumeric(6)))
	}

	for i := 1; i < len(os.Args); i++ {
		if os.Args[i][0] == '-' {
			continue
		}
		if os.Args[i] == fmt.Sprintf("%d", runner.Buffer) {
			continue
		}
		if os.Args[i] == runner.SocketPath {
			continue
		}
		if os.Args[i] == runner.LogFile {
			continue
		}
		runner.ArgIndex = i
		break
	}

	if runner.Version {
		fmt.Println("sudo for windows v0")
		return
	}
	if runner.Check {
		fmt.Println("argIndex:", runner.ArgIndex)
		fmt.Println("isAdmin:", runner.IsAdmin())
		fmt.Println("buffer:", runner.Buffer)
		fmt.Println("bufSize:", runner.BufSize)
		fmt.Println("debug:", runner.Debug)
		fmt.Println("sockfile:", runner.SocketPath)
		fmt.Println("log:", runner.LogFile)
		return
	}

	if runner.ArgIndex == 0 {
		fmt.Println("sudo: error: must specify program")
		return
	}

	_, err := exec.LookPath(os.Args[runner.ArgIndex])
	if err != nil {
		fmt.Println("sudo: error: unable to find", os.Args[runner.ArgIndex], "in PATH")
		return
	}

	if runner.IsClient {
		if runner.Debug {
			runner.Logger, err = os.OpenFile(runner.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				fmt.Println("sudo: error: unable to open log file:", err)
				return
			}
		}
		runner.StartClient()
		return
	}

	// fmt.Println(os.Args[runner.ArgIndex:])
	runner.StartServer()
}
