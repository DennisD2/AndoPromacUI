package main

import "fmt"

type GenericState int

const (
	GENERIC_START GenericState = 0
	GENERIC_DATA               = 1
	GENERIC_END                = 2
)

type GenericData struct {
	cr_count   uint32
	lf_count   uint32
	zero_count uint32
	state      GenericState
}

var genericState *GenericData = nil

func initGenericFormat() {
	genericState = new(GenericData)
	genericState.zero_count = 0
	genericState.cr_count = 0
	genericState.lf_count = 0
	genericState.state = GENERIC_START
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
		if genericState.state == GENERIC_DATA {

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
