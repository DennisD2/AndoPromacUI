package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// handleASCIIHexInput handles input (also download) coming in from Eprommer
func handleASCIIHexInput(ando *AndoConnection, num int, cbuf []byte, newLine *LineInfo, lineNumber *int, errors *int) {
	for i := 0; i < num; i++ {
		if cbuf[i] == '\n' {
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
		}
	}
}

// dumpLine pretty print a line received with address and hex codes
func dumpLine(line LineInfo) {
	/*if line.address > 100 && line.address < 0x3f00 {
		return
	}*/
	fmt.Printf("%06d %08x", line.lineNumber, line.address)
	//fmt.Printf("%v\n\r", line.raw)
	for _, info := range line.codes {
		fmt.Printf(" %02x", info)
	}
	fmt.Printf("\n\r")
}

// extractData extracts a string representing a line from Programmer device into address and hex values
func extractData(l *LineInfo, errors *int, checksum *uint32) bool {
	if strings.Contains(l.raw, "[#") {
		log.Printf("Start line\n\r")
		index := strings.Index(l.raw, "[#")
		l.raw = l.raw[index+1:]
	}
	if strings.HasPrefix(l.raw, "#") {
		firstCommaPos := strings.Index(l.raw, ",")
		if firstCommaPos == -1 {
			log.Printf("Line %v contains no ',' character. Line ignored", l.lineNumber)
			*errors++
			return false
		}
		addressPart := l.raw[1:firstCommaPos]
		value, err := strconv.ParseUint(addressPart, 16, 32)
		if err != nil {
			log.Printf("Error converting address %v Line %v", addressPart, l.lineNumber)
			*errors++
			return false
		}
		l.address = uint32(value)

		valuesPart := l.raw[firstCommaPos+1:]
		codes := strings.Split(valuesPart, ",")
		if len(codes) != 17 {
			log.Printf("Line contains %v codes (expected 17) at address %v Line %v", len(l.codes), l.address, l.lineNumber)
			*errors++
		}
		for i := 0; i < 16; i++ {
			value, err := strconv.ParseUint(codes[i], 16, 8)
			if err != nil {
				log.Printf("Error converting value %v, index %v, in Line %v", codes[i], i, l.lineNumber)
				*errors++
				return false
			}
			val := uint8(value)
			l.codes[i] = val
			*checksum += uint32(val)
		}
		return true
	}
	return false
}

// uploadFile uploads local file to EPrommer's RAM buffer
func uploadFile(ando *AndoConnection) {
	sb := new(strings.Builder)

	// Read in file
	bytes, err := os.ReadFile(ando.uploadFile)
	if err != nil {
		log.Printf("Error loading input file %s: %s\n\r", ando.uploadFile, err)
		return
	}
	log.Printf("Loaded input file %s, %v bytes\n", ando.uploadFile, len(bytes))

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

		i++
		bytesInLine++
		if bytesInLine == 16 {
			sb.WriteString("\r")
		}
	}
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
