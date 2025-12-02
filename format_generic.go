package main

import (
	"fmt"
	"log"
)

type GenericState int

const (
	GENERIC_START  GenericState = 0
	GENERIC_DATA                = 1
	GENERIC_END_CR              = 2
	GENERIC_END_LF              = 3
	GENERIC_END                 = 4
)

type GenericData struct {
	cr_count        uint32
	lf_count        uint32
	zero_count      uint32
	lastCharWasZero bool
	state           GenericState
	byteCount       uint32
}

var genericState *GenericData = nil

func initGenericFormat() {
	genericState = new(GenericData)
	genericState.zero_count = 0
	genericState.cr_count = 0
	genericState.lf_count = 0
	genericState.state = GENERIC_START
	genericState.lastCharWasZero = false
	genericState.byteCount = 0
}

func handleGenericInput(ando *AndoConnection, num int, cbuf []byte, line *LineInfo, number *int, errors *int) {
	if genericState == nil {
		initGenericFormat()
	}

	for i := 0; i < num; i++ {
		b := cbuf[i]
		fmt.Printf("%02x ", b)
		if genericState.state == GENERIC_START {
			if b == 0x0 {
				genericState.zero_count++
			} else if b == 0xa {
				genericState.cr_count++
			} else if b == 0xd {
				genericState.lf_count++
			}
			if genericState.zero_count == 100 && genericState.cr_count == 3 && genericState.lf_count == 3 {
				fmt.Printf("Header end!\n\r")
				genericState.state = GENERIC_DATA
				genericState.cr_count = 0
				genericState.lf_count = 0
				genericState.zero_count = 0
			}
			/*else {
				in_start = false
				fmt.Printf(" zeros=%v other=%v\n\r", zero_count, other_count)
			}*/
		}
		// wait for 100 x '0x0'
		if genericState.state == GENERIC_DATA {
			genericState.byteCount++
			if b == 0x0 {
				if genericState.lastCharWasZero {
					genericState.zero_count++
				} else {
					genericState.lastCharWasZero = true
					genericState.zero_count = 1
				}
			} else {
				genericState.lastCharWasZero = false
			}
			if genericState.zero_count == 100 {
				genericState.state = GENERIC_END_CR
				fmt.Printf("100 zeros received\n\r")
				continue
			}
		}
		// wait for 0xd,0xa
		if genericState.state == GENERIC_END_CR {
			if b == 0xd {
				genericState.state = GENERIC_END_LF
				fmt.Printf("0xd received\n\r")
				continue
			} else {
				genericState.state = GENERIC_DATA
			}
		}
		// wait for 0xd,0xa
		if genericState.state == GENERIC_END_LF {
			if b == 0xa {
				genericState.state = GENERIC_END
				log.Printf("0xa received\n\r")
				fmt.Printf("End Block received\n\r")
				fmt.Printf("Bytes in data: %v\n\r", genericState.byteCount-100-2)
			} else {
				genericState.state = GENERIC_DATA
			}

		}
		/*if cbuf[i] == '\n' {
			if ando.state == ReceiveData {
				newLine.lineNumber = *lineNumber
				valid := extractData(newLine, errors, &ando.checksum)
				if valid {
					ando.lineInfos = append(ando.lineInfos, *newLine)
					dumpLine(*newLine)
					*lineNumber++
				}
				newLine.raw = ""
			} else {
				fmt.Printf("\n\r")
			}
		} else {
			if ando.state == ReceiveData {
				newLine.raw = newLine.raw + string(cbuf[i])
			} else {
				fmt.Printf("%c", cbuf[i])
			}
		}*/
	}
	fmt.Printf("\n\r")
}
