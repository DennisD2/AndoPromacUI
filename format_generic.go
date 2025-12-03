package main

import (
	"fmt"
	"strings"
)

type GenericData struct {
	rawCount uint32
	rawData  []byte
}

var genericState *GenericData = new(GenericData)

func initGenericFormat(ando *AndoConnection) {
	genericState = new(GenericData)
}

func handleGenericInput(ando *AndoConnection, num int, cbuf []byte, line *LineInfo, number *int, errors *int) {
	for i := 0; i < num; i++ {
		b := cbuf[i]
		fmt.Printf("%02x ", b)
		genericState.rawCount++
		genericState.rawData = append(genericState.rawData, b)
	}
	fmt.Printf("\n\r")
}

func parseGeneric(ando *AndoConnection) {
	fmt.Printf("Read %v raw bytes\n\r", len(genericState.rawData))

	valid, dataStart := isRawHeader(genericState.rawData)
	if !valid {
		fmt.Printf("Not a raw header!\n\r")
	}
	valid, dataEnd := isRawFooter(genericState.rawData)
	if !valid {
		fmt.Printf("Not a raw footer!\n\r")
	}
	fmt.Printf("%v bytes in range %v-%v\n\r", (dataEnd - dataStart), dataStart+1, dataEnd)

	sb := new(strings.Builder)
	sb.WriteString("\n\r")
	address := 0
	i := dataStart + 1
	bytesInLine := 0
	for i <= dataEnd {
		if (i-dataStart-1)%16 == 0 {
			str := fmt.Sprintf("%08x ", address)
			sb.WriteString(str)
			address += 16
			bytesInLine = 0
		}
		b := genericState.rawData[i]
		str := fmt.Sprintf("%02x ", b)
		sb.WriteString(str)

		i++
		bytesInLine++
		if bytesInLine == 16 {
			sb.WriteString("\r\n")
		}
	}
	fmt.Printf("%v\n\r", sb.String())
}

func isRawHeader(data []byte) (bool, int) {
	if data[0] != 0xd || data[1] != 0xa {
		return false, 0
	}
	if data[2] != 0xd || data[3] != 0xa {
		return false, 0
	}
	if data[4] != 0xd || data[5] != 0xa {
		return false, 0
	}
	var i = 6
	for ; i < 106; i++ {
		if data[i] != 0x0 {
			return false, 0
		}
	}
	return true, i
}

func isRawFooter(data []byte) (bool, int) {
	pos := len(data)
	if data[pos-2] != 0xd || data[pos-1] != 0xa {
		return false, 0
	}
	pos = pos - 3
	for i := pos; i > pos-100; i-- {
		if data[i] != 0x0 {
			return false, 0
		}
	}
	return true, pos - 100
}
