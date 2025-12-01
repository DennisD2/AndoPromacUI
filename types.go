package main

import (
	"os"
	"time"
)

// AndoSerialConnection TTY connection to Programmer
type AndoSerialConnection struct {
	tty      *os.File
	device   string
	baudrate int
	timeout  time.Duration
	//dryMode      bool
	//debug        int
	//batch        bool
}

// ConnState State of Connection
type ConnState int

const (
	NormalInput  ConnState = 0
	CommandInput           = 1
	ReceiveData            = 3
	SendData               = 4
)

type TransferFormat int

const (
	ASCIIHex   TransferFormat = 0
	HP64000ABS                = 1
)

// Connection connection to Eprommer
type AndoConnection struct {
	continueLoop     int       // true as long as command loop runs
	state            ConnState // state of app
	dryMode          bool      // dry mode means do not really invoke EPrommer device
	debug            int       // debug level
	batch            bool      // batch mode
	uploadFile       string    // file to upload to EPrommer device
	downloadFile     string    // file to download from EPrommer device
	transferFormat   TransferFormat
	serial           *AndoSerialConnection // Serial onnection structure used
	lineInfos        []LineInfo            // internal representation of EPROM data during download
	checksum         uint32                // checksum value
	transferPosition int
	hp64k            *HP64KInfo

	//subState         SubState
}

// LineInfo info for a line sent by Programmer Device
type LineInfo struct {
	lineNumber int      // number of line
	address    uint32   // end address of bytes in line
	codes      [16]byte // 16 bytes
	raw        string   // raw chars received from EPrommer device
}
