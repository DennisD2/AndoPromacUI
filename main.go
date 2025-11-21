package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"os"

	"golang.org/x/term"
)

func main() {
	fmt.Println("Ando/Promac EPROM Programmer Communication UI")
	devicePtr := flag.String("device", "/dev/ttyUSB0",
		"TTY device used to access EPrommer")
	dryRunPtr := flag.Bool("dry-run", false,
		"Dry run mode")
	debugPtr := flag.Int("debug", 0,
		"Debug level")
	baudratePtr := flag.Int("baudrate", 19200,
		"Baudrate")
	batchPtr := flag.Bool("batch", false,
		"Non-interactive (batch) mode")
	uploadPtr := flag.String("infile", "in.bin",
		"Input file for EPROM data to upload to EPrommer")
	downloadPtr := flag.String("outfile", "out",
		"Output file for EPROM data downloaded from EPrommer")
	flag.Parse()

	fmt.Printf("--device, TTY Device: %s\n", *devicePtr)
	fmt.Printf("--dry-run: %t\n", *dryRunPtr)
	fmt.Printf("--debug: %d\n", *debugPtr)
	fmt.Printf("--baudrate: %d\n", *baudratePtr)
	fmt.Printf("--outfile: %s-<checksum>.bin\n", *downloadPtr)
	fmt.Printf("--batch: %t (batch mode not yet supported)\n", *batchPtr)
	fmt.Printf("--infile: %s\n", *uploadPtr)

	// Create serial connection
	andoSerial := AndoSerialConnection{
		nil, //priv
		*devicePtr,
		*baudratePtr,
		0,
	}

	// Create Device structure
	ando := AndoConnection{
		1,           //priv
		NormalInput, //priv
		*dryRunPtr,
		*debugPtr,
		*batchPtr,
		*uploadPtr,
		*downloadPtr,
		&andoSerial,
		nil,
		0,
	}

	if !ando.dryMode {
		// open tty reader
		err := ando.serial.openTTY()
		if err != nil {
			fmt.Println(err)
			return
		}
		defer ando.serial.tty.Close()
	}

	var oldState *term.State
	var err error
	if !ando.batch {
		// switch stdin into 'raw' mode
		oldState, err = term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			fmt.Println(err)
			return
		}
		defer term.Restore(int(os.Stdin.Fd()), oldState)

		// Start local keyboard handler routine
		go localKeyboardReader(&ando)

		// Start tty routine
		go ttyReader(&ando)

		// stay in loop until end condition is met
		for ando.continueLoop > 0 {
			time.Sleep(25 * time.Millisecond)
		}
	} else {
		// just upload
		//uploadFile(&ando)
	}

	fmt.Println("\n\rQuitting Ando/Promac EPROM Programmer Communication UI\n\r")
	if !ando.batch {
		term.Restore(int(os.Stdin.Fd()), oldState)
	}
	os.Exit(0)
}

// ttyReader handle tty input from Programmer device
func ttyReader(ando *AndoConnection) {
	var newLine LineInfo
	var lineNumber = 1
	cbuf := make([]byte, 128)
	errors := 0
	for ando.continueLoop > 0 {
		if ando.dryMode {
			continue
		}
		// check Ando tty
		num, err := ando.serial.tty.Read(cbuf)
		if err != nil {
			fmt.Printf("Error in Read: %s\n", err)
			ando.continueLoop = 0
		} else {
			str := string(cbuf)
			if strings.HasPrefix(str, "[PASS]") {
				if ando.state == ReceiveData {
					log.Printf("Data receive completed. Read %v bytes in %v lines\n\r", (lineNumber-1)*16, lineNumber-1)
					log.Printf("Checksum calculated: %06x\n\r", ando.checksum)
					if errors > 0 {
						fmt.Printf("There were %v errors\n\r", errors)
						errors = 0
					}
					lineNumber = 1
				}
				if ando.state == SendData {
					log.Printf("\n\rUpload completed for all bytes from file %v\n\r", ando.uploadFile)
				}
				if ando.state == ReceiveData || ando.state == SendData {
					// Data receive/send is complete
					ando.state = NormalInput
					// leave S-OUTPUT or S-INPUT state, by sending RESET character
					bbuf := make([]byte, 1)
					bbuf[0] = '@'
					ando.serial.tty.Write(bbuf)
				}
			} else {
				handleTTYInput(ando, num, cbuf, &newLine, &lineNumber, &errors)
			}
		}
	}
}

// handleTTYInput handle the incoming byte sequences according to app state
func handleTTYInput(ando *AndoConnection, num int, cbuf []byte, newLine *LineInfo, lineNumber *int, errors *int) {
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

// localKeyboardReader handles all local keyboard input and interaction
func localKeyboardReader(ando *AndoConnection) {
	cbuf := make([]byte, 128)

	consoleReader := bufio.NewReader(os.Stdin)
	b := make([]byte, 1)
	helpText(ando)
	for ando.continueLoop > 0 {
		if ando.state != CommandInput {
			fmt.Printf("Command > ")
		}
		num, err := consoleReader.Read(cbuf)
		if err != nil {
			fmt.Println(err)
		} else {
			if num == 0 {
				continue
			}
			if num > 1 {
				// We currently cannot handle multiple chars at once
				fmt.Println("Multiple chars!")
			}

			if ando.state == CommandInput {
				fmt.Printf("%s", cbuf)
				// In command mode, execute command based on key input
				if cbuf[0] == ':' {
					ando.state = NormalInput
					fmt.Println(" Back to normal input handling\n\r")
					continue
				}
				if cbuf[0] == 'q' {
					ando.continueLoop = 0
					ando.state = NormalInput
				}
				if cbuf[0] == 'd' {
					ando.lineInfos = nil
					ando.checksum = 0
					fmt.Println("\n\r")
					ando.state = ReceiveData
					bbuf := make([]byte, 8)
					bbuf[0] = 'U'
					bbuf[1] = '7'
					bbuf[2] = '\r'
					ando.serial.tty.Write(bbuf)
				}
				if cbuf[0] == 'w' {
					ando.state = NormalInput
					writeDataToFile(ando)
				}
				if cbuf[0] == 'u' {
					ando.state = NormalInput
					uploadFile(ando)
				}
				continue
			}

			if ando.state == NormalInput {
				if cbuf[0] == ':' {
					// If ':' is selected, check next char for command to execute
					// We switch state to CommandInput for that
					ando.state = CommandInput
					fmt.Print(" [:qdw] >")
					continue
				}
			}
			// Normal input, forward it to tty
			b[0] = cbuf[0]
			if !ando.dryMode {
				if ando.debug > 0 {
					fmt.Printf("<%d:%s:%x>", num, b, b)
				} else {
					ando.serial.tty.Write(b)
				}
			}
		}
	}
}

// helpText print help text
func helpText(ando *AndoConnection) {
	fmt.Print("Commands:\n\r")
	fmt.Print(" @		- RESET\n\r")
	fmt.Print(" P A <CR>	- DEVICE-COPY\n\r")
	fmt.Print(" P C <CR>	- DEVICE-BLANK\n\r")
	fmt.Print(" P D <CR>	- DEVICE-PROGRAM\n\r")
	fmt.Print(" P E <CR>	- DEVICE-VERIFY\n\r")
	fmt.Print(" U 9 <CR>	- Quit REMOTE CONTROL\n\r")
	fmt.Print(" U 6 <CR>	- Send data to EPrommer\n\r")
	fmt.Print(" U 7 <CR>	- Receive Data from EPrommer\n\r")
	fmt.Print(" U 8 <CR>	- VERIFY\n\r")
	fmt.Print("Compound Commands:\n\r")
	fmt.Print(" : q		- Quit Ando/Promac EPROM Programmer Communication UI\n\r")
	fmt.Print(" : d		- Download EPROM data (like U7)\n\r")
	fmt.Printf(" : w		- Write EPROM data to file %v-<checksum>.bin\n\r", ando.downloadFile)
	fmt.Printf(" : u		- Upload EPROM data from file %v to EPrommer\n\r", ando.uploadFile)
	fmt.Print("\n\r")
}

// writeDataToFile writes data from AndoConnection.lineInfos to AndoConnection.downloadFile,
// data is written to EPrommer's RAM buffer
func writeDataToFile(ando *AndoConnection) {
	numBytes := 0
	sb := new(strings.Builder)
	// Convert codes to byte stream
	for _, line := range ando.lineInfos {
		for i, code := range line.codes {
			if i < 16 {
				sb.WriteByte(code)
				numBytes++
			}
		}
	}
	// Write file
	filename := createFileName(ando.downloadFile, ando.checksum)
	err := os.WriteFile(filename, []byte(sb.String()), 0644)
	if err != nil {
		log.Printf("Error Writing file %s\n\r", err)
		return
	}
	fmt.Printf("\n\rWrote %v bytes to file\n\r", numBytes)
}

func createFileName(file string, checksum uint32) string {
	fname := fmt.Sprintf("%v-%06x.bin", file, checksum)
	log.Printf("Created file name: %v", fname)
	return fname
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
