package runner

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
)

func IsAdmin() bool {
	if _, err := os.Open("C:\\Program Files\\WindowsApps"); err != nil {
		return false
	}
	return true
}

type Client struct {
	net.Conn
}

func (cli *Client) WriteMsg(msg *Msg) (n int, err error) {
	if len(msg.Data) > 0 {
		for i := int64(0); i < int64(len(msg.Data)); i += BufSize {
			var data []byte

			if i > int64(len(msg.Data))-BufSize {
				data = msg.Data[i:]
			} else {
				data = msg.Data[i : i+BufSize]
			}

			written, writeErr := cli.Write(NewMsg(msg.Op, data).Bytes())
			n += written
			if writeErr != nil {
				return n, writeErr
			}
		}
		//cli.Write(NewMsgDF("%d", n).Bytes())
		return n, nil
	}
	//cli.Write(NewMsgDF("%d", len(msg.Bytes())).Bytes())
	return cli.Write(msg.Bytes())
}
func (cli *Client) Close() {
	cli.WriteMsg(MsgEOF)
	cli.Conn.Close()
}

func StartClient() {
	// cli, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", Port))
	cli, err := net.Dial("unix", SocketPath)

	if err != nil {
		Logger.WriteString(fmt.Sprintf("sudo: error: unable to connect to server: %v\n", err))
		fmt.Println("sudo: error: unable to connect to server:", err)
		return
	}
	client := &Client{cli}
	defer client.Close()

	if !IsAdmin() {
		client.WriteMsg(NewMsgEF("no privileges"))
		return
	}

	client.WriteMsg(NewMsgD("client: Creating cmd..."))
	cmd := exec.Command(os.Args[ArgIndex])
	if len(os.Args) > ArgIndex+1 {
		cmd = exec.Command(os.Args[ArgIndex], os.Args[ArgIndex+1:]...)
	}

	cmd.Dir, err = os.Getwd()
	if err != nil {
		client.WriteMsg(NewMsgEF("cmdDir: %v", err))
		return
	}

	cmdIn, err := cmd.StdinPipe()
	if err != nil {
		client.WriteMsg(NewMsgEF("cmdIn: %v", err))
		return
	}
	cmdOut, err := cmd.StdoutPipe()
	if err != nil {
		client.WriteMsg(NewMsgEF("cmdOut: %v", err))
		return
	}
	cmdErr, err := cmd.StderrPipe()
	if err != nil {
		client.WriteMsg(NewMsgEF("cmdErr: %v", err))
		return
	}
	go cmdReadOutput(cmdOut, client)
	go cmdReadError(cmdErr, client)
	go cmdReadInput(cmdIn, client)

	client.WriteMsg(NewMsgD("client: Spawning target process..."))
	err = cmd.Start()
	if err != nil {
		client.WriteMsg(NewMsgEF("unable to spawn target process: %v", err))
		return
	}

	client.WriteMsg(NewMsgD("client: Waiting for process to exit..."))
	err = cmd.Wait()
	if err != nil {
		client.WriteMsg(NewMsgEF("cmd: %v", err))
		return
	}
}

func cmdReadOutput(cmdOut io.ReadCloser, client *Client) {
	defer cmdOut.Close()
	for {
		buf := make([]byte, BufSize)

		n, err := cmdOut.Read(buf)
		if err != nil {
			if err == io.EOF {
				Logger.WriteString("cmdReadOutput: EOF\n")
				client.WriteMsg(MsgEOF)
			} else {
				client.WriteMsg(NewMsgEF("cmdReadOutput: %v", err))
			}
			return
		}

		if n > 0 {
			msg := NewMsg(opStdout, buf[:n])
			client.WriteMsg(msg)
		}
	}
}

func cmdReadError(cmdErr io.ReadCloser, client *Client) {
	defer cmdErr.Close()
	for {
		buf := make([]byte, BufSize)

		n, err := cmdErr.Read(buf)
		if err != nil {
			if err == io.EOF {
				Logger.WriteString("cmdReadError: EOF\n")
				client.WriteMsg(MsgEOF)
			} else {
				client.WriteMsg(NewMsgEF("cmdReadError: %v", err))
			}
			return
		}

		if n > 0 {
			msg := NewMsg(opStderr, buf[:n])
			client.WriteMsg(msg)
		}
	}
}

func cmdReadInput(cmdIn io.WriteCloser, client *Client) {
	defer cmdIn.Close()
	for {
		buf := make([]byte, BufSize)

		n, err := client.Read(buf)
		if err != nil {
			client.WriteMsg(NewMsgEF("cmdReadInput: %v", err))
			return
		}

		if n > 0 {
			msgs, err := NewMsgIn(buf[:n])
			if err != nil {
				client.WriteMsg(NewMsgEF("NewMsgIn: %v", err))
				continue
			}

			for i := 0; i < len(msgs); i++ {
				msg := msgs[i]
				switch msg.Op {
				case opEOF:
					client.WriteMsg(NewMsgD("client: Goodbye!"))
					return
				case opStdin:
					client.WriteMsg(NewMsgD("client: Writing stdin..."))
					cmdIn.Write(msg.Data)
				default:
					client.WriteMsg(NewMsgEF("cmdReadInput: unknown op: %d", msg.Op))
				}
			}
		}
	}
}
