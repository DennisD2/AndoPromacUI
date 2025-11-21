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
	NormalInput        ConnState = 0
	CommandInput                 = 1
	ReceiveData                  = 3
	WaitForPassMessage           = 4
)
