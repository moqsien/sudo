package runner

import (
	"flag"
	"os"
)

var (
	MsgEOF *Msg = &Msg{Op: opEOF}

	ArgIndex = 0
	// Port       int64  = 0
	SocketPath string = ""
	BufSize    int64  = 65536
	Buffer     int64  = 65536
	Version    bool   = false
	Check      bool   = false
	Debug      bool   = false
	IsClient   bool   = false

	LogFile string
	Logger  *os.File
)

func init() {
	flag.Int64Var(&Buffer, "buffer", Buffer, "Set max size of all byte buffers")
	flag.BoolVar(&Version, "version", Version, "Set to display version info and exit")
	flag.BoolVar(&Check, "check", Check, "Set to check for admin privileges and exit")
	flag.BoolVar(&Debug, "debug", Debug, "Set to enable debug logging")
	flag.BoolVar(&IsClient, "client", IsClient, "Set to act as client")
	// flag.Int64Var(&Port, "port", Port, "Custom TCP port for session (0 for randomization)")
	flag.StringVar(&SocketPath, "socketfile", SocketPath, "Unix socket file for session(blank for randomization).")
	flag.StringVar(&LogFile, "log", LogFile, "File to log packets to when debugging")
	flag.Parse()

	BufSize = Buffer - int64(len(MsgEOF.Bytes())) //Ensure the size of a msg is accounted for
}
