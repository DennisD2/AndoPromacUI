package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

// parseASCIIHexFormat parses ASCII Hex transfer format data
func parseASCIIHexFormat(ando *AndoConnection, lineNumber *int, errors *int) {
	log.Printf("Parsing ASCII-Hex format\n\r")
	valid, dataStart := isRawHeaderASCIIHex(genericState.rawData)
	if !valid {
		log.Printf("Not a ASCII-Hex header!\n\r")
		*errors++
	}
	valid, dataEnd := isRawFooterASCIIHex(genericState.rawData)
	if !valid {
		log.Printf("Not a ASCII-Hex footer!\n\r")
		*errors++
	}
	log.Printf("%v bytes in range %v-%v\n\r", (dataEnd - dataStart), dataStart, dataEnd)

	var lineBytes []byte
	i := dataStart
	for i <= dataEnd {
		b := genericState.rawData[i]
		if b != 0xa && b != 0xd {
			lineBytes = append(lineBytes, b)
		}
		i++
		if b == 0xa {
			// We have a complete line, with address and all 16 data bytes
			newLine := new(LineInfo)
			newLine.lineNumber = *lineNumber
			valid = parseLine(lineBytes, *lineNumber, newLine, errors, &ando.checksum)
			if valid {
				ando.lineInfos = append(ando.lineInfos, *newLine)
				dumpLine(*newLine)
				*lineNumber++
			} else {
				log.Printf("Line read fail: '%v'\n\r", string(lineBytes))
			}
			lineBytes = lineBytes[:0]
		}
	}
}

// parseLine extracts all data from a line downloaded (i.e. address and byte values)
func parseLine(bytes []byte, lineNumber int, lineInfo *LineInfo, errors *int, checksum *uint32) bool {
	var line = string(bytes)
	if strings.HasPrefix(line, "[") {
		line = line[1:]
	}
	if strings.HasPrefix(line, "#") {
		line = line[1:]
	} else {
		return false
	}

	firstCommaPos := strings.Index(line, ",")
	if firstCommaPos == -1 {
		log.Printf("Line '%v' contains no ',' character. Line ignored", line)
		*errors++
		return false
	}
	addressPart := line[1:firstCommaPos]
	value, err := strconv.ParseUint(addressPart, 16, 32)
	if err != nil {
		log.Printf("Error converting address %v Line '%v'", addressPart, line)
		*errors++
		return false
	}
	lineInfo.address = uint32(value)

	valuesPart := line[firstCommaPos+1:]
	codes := strings.Split(valuesPart, ",")
	if len(codes) != 17 {
		log.Printf("Line contains %v codes (expected 17) at address %v Line %v", len(lineInfo.codes), lineInfo.address, lineInfo.lineNumber)
		*errors++
	}
	for i := 0; i < 16; i++ {
		value, err := strconv.ParseUint(codes[i], 16, 8)
		if err != nil {
			log.Printf("Error converting value %v, index %v, in Line %v", codes[i], i, lineInfo.lineNumber)
			*errors++
			return false
		}
		val := uint8(value)
		lineInfo.codes[i] = val
		*checksum += uint32(val)
	}

	return true
}

// uploadFileAsASCIIHex uploads local file to EPrommer's RAM buffer, transfer format being used is ASCII-Hex
func uploadFileAsASCIIHex(ando *AndoConnection, errors *int) {
	sb := new(strings.Builder)
	var checksum uint32 = 0

	bytes, error := loadFile(ando, errors)
	if error {
		return
	}

	// Write prefix char
	sb.WriteString("[")

	address := 0
	i := 0
	bytesInLine := 0
	for i < len(bytes) {
		if i%16 == 0 {
			str := fmt.Sprintf("#%08x,", address)
			sb.WriteString(strings.ToUpper(str))
			address += 16
			bytesInLine = 0
		}
		b := bytes[i]
		str := fmt.Sprintf("%02x,", b)
		sb.WriteString(strings.ToUpper(str))
		checksum += uint32(b)

		i++
		bytesInLine++
		if bytesInLine == 16 {
			sb.WriteString("\r")
		}
	}
	log.Printf("Upload data checksum: 0x%06x\n\r", checksum)
	log.Printf("Upload buffer has size %v bytes. Please wait for upload to complete...\n\r", sb.Len())

	// Send data collected
	bbuf := make([]byte, 3)
	bbuf[0] = 'U'
	bbuf[1] = '6'
	bbuf[2] = '\r'
	ando.serial.tty.Write(bbuf)
	// give some time to have command understood
	time.Sleep(100 * time.Millisecond)

	i = 0
	sendString := sb.String()
	b := make([]byte, 1)
	for i < len(sendString) {
		//fmt.Printf("%c\n\r", sendString[i])
		b[0] = byte(sendString[i])
		ando.serial.tty.Write(b)
		i++
	}

	// device will need some time to process all data
	// We need to wait for "[PASS]" answer
	// only then, the final RESET '@' we like to send will be handled by device.
	// If we do not wait, the Programmer stays in S-INPUT mode, and we have to enter RESET via device key "RESET"
	// or send it via "Ando/Promac EPROM Programmer Communication UI" by using the '@' key
	// So we go to new state and wait there for incoming "[PASS]" message
	ando.state = SendData
}

// firmware 21.7 and 21.9
// 52 and 99 have been analyzed from download data. The manual says always 100x zeroes (page 28)
const NUM_HEADER_ZEROES = 52
const NUM_FOOTER_ZEROES = 99

// isRawHeaderASCIIHex returns true if this is a correct ASCII Hex transfer data header
func isRawHeaderASCIIHex(data []byte) (bool, int) {
	num_zeros := 0
	/*fmt.Printf("\r\nAAAA: ")
	for i := 0; i < 256; i++ {
		fmt.Printf("%02x ", data[i])
		if i%16 == 0 {
			fmt.Printf("\r\n")
		}
	}
	fmt.Printf("\r\n")*/
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
	for ; i < NUM_HEADER_ZEROES+6; i++ {
		//fmt.Printf("AAAAAA %v %02x\n\r ", i, data[i])
		if data[i] != 0x0 {
			return false, 0
		}
		num_zeros++
	}
	log.Printf("ASCII-Hex header OK\r\n")

	// Overread remaining 0x0
	for ; data[i] == 0x0; i++ {
		num_zeros++
	}
	log.Printf("Number of header zero bytes read: %v\r\n", num_zeros)
	return true, i
}

// isRawFooterASCIIHex returns true if this is a correct ASCII Hex transfer data header
func isRawFooterASCIIHex(data []byte) (bool, int) {
	num_zeros := 0
	pos := 0
	var i int
	for i = len(data) - 1; i > 0; i-- {
		if data[i] == 0xa {
			break
		}
	}
	if i == 0 {
		log.Printf("No 0xa marker found in data")
		return false, 0
	}
	pos = i + 1
	if data[pos-2] != 0xd || data[pos-1] != 0xa {
		log.Printf("XXXX %02x %02x\n\r", data[pos-2], data[pos-1])
		log.Printf("\r\nZZZZ: ")
		for i := 0; i < 16; i++ {
			log.Printf("%02x ", data[pos-16+i])
		}
		log.Printf("\r\n")
		return false, 0
	}
	pos = pos - 3
	for i := pos; i > pos-NUM_FOOTER_ZEROES; i-- {
		//fmt.Printf("YYYY %02x\n\r", data[i])
		if data[i] != 0x0 {
			return false, 0
		}
	}
	// Overread remaining 0x0
	for ; data[i] == 0x0; i-- {
		num_zeros++
	}
	log.Printf("ASCII-Hex footer OK\r\n")
	log.Printf("Number of header zero bytes read: %v\r\n", num_zeros)
	return true, pos - NUM_FOOTER_ZEROES
}

// dumpLine pretty print a line received with address and hex codes
func dumpLine(line LineInfo) {
	fmt.Printf("%06d %08x", line.lineNumber, line.address)
	for _, info := range line.codes {
		fmt.Printf(" %02x", info)
	}
	fmt.Printf("\n\r")
}
