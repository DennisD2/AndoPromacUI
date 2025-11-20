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

// Connection connection to Eprommer
type AndoConnection struct {
	continueLoop int
	state        ConnState
	dryMode      bool
	debug        int
	batch        bool
	uploadFile   string
	serial       *AndoSerialConnection
	lineInfos    []LineInfo
}

// LineInfo info for a line sent by Programmer Device
type LineInfo struct {
	lineNumber int      // number of line
	address    uint32   // end address of bytes in line
	codes      [16]byte // 16 bytes
	raw        string
}

func main() {
	fmt.Println("Ando/Promac EPROM Programmer Communication UI")
	devicePtr := flag.String("device", "/dev/ttyUSB0",
		"TTY device used to access Eprommer")
	dryRunPtr := flag.Bool("dry-run", false,
		"Dry run mode")
	debugPtr := flag.Int("debug", 0,
		"Debug level")
	baudratePtr := flag.Int("baudrate", 19200,
		"Baudrate")
	batchPtr := flag.Bool("batch", false,
		"Non-interactive (batch) mode")
	uploadPtr := flag.String("upload", "multiecho.deposit",
		"Deposit file to upload")
	flag.Parse()

	fmt.Printf("--device, TTY Device: %s\n", *devicePtr)
	fmt.Printf("--dry-run: %t\n", *dryRunPtr)
	fmt.Printf("--debug: %d\n", *debugPtr)
	fmt.Printf("--baudrate: %d\n", *baudratePtr)
	fmt.Printf("--batch: %t (batch mode not yet supported)\n", *batchPtr)
	fmt.Printf("--upload: %s (batch mode not yet supported)\n", *uploadPtr)

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
		&andoSerial,
		nil,
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
		uploadFile(&ando, *uploadPtr, ando.debug)
	}

	fmt.Println("\n\rQuitting Ando/Promac EPROM Programmer Communication UI\n\r")
	term.Restore(int(os.Stdin.Fd()), oldState)
	os.Exit(0)
}

func uploadFile(a *AndoConnection, s string, debug int) {
	fmt.Println("\n\rBATCH UPLOAD YET UNSUPPORTED\n\r")
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
			if ando.state == ReceiveData && strings.HasPrefix(str, "[PASS]") {
				// Data receive is complete
				ando.state = NormalInput
				// leave S-OUTPUT state
				bbuf := make([]byte, 1)
				bbuf[0] = '@'
				ando.serial.tty.Write(bbuf)
				fmt.Printf("Data receive completed. Read %v lines\n\r", lineNumber-1)
				if errors > 0 {
					fmt.Printf("There were %v errors\n\r", errors)
					errors = 0
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
				extractData(newLine, errors)
				if newLine.address > 0 {
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
func dumpLine(newLine LineInfo) {
	fmt.Printf("%06d %08x", newLine.lineNumber, newLine.address)
	for _, info := range newLine.codes {
		fmt.Printf(" %02x", info)
	}
	fmt.Printf("\n\r")
}

// extractData extracts a string representing a line from Programmer device into address and hex values
func extractData(l *LineInfo, errors *int) {
	if strings.HasPrefix(l.raw, "#") {
		firstCommaPos := strings.Index(l.raw, ",")
		if firstCommaPos == -1 {
			log.Printf("Line %v contains no ',' character. Line ignored", l.lineNumber)
			*errors++
			return
		}
		addressPart := l.raw[1:firstCommaPos]
		value, err := strconv.ParseUint(addressPart, 16, 32)
		if err != nil {
			log.Printf("Error converting address %v Line %v", addressPart, l.lineNumber)
			*errors++
			return
		}
		l.address = uint32(value)

		valuesPart := l.raw[firstCommaPos+1:]
		codes := strings.Split(valuesPart, ",")
		if len(codes) != 17 {
			log.Printf("Line contains %v codes (expected 16) at address %v Line %v", len(l.codes), l.address, l.lineNumber)
			*errors++
		}
		if strings.HasSuffix(codes[16], "\r") {
			codes[16] = strings.Replace(codes[16], "\r", "", -1)
		}
		for i := 0; i < 16; i++ {
			value, err := strconv.ParseUint(codes[i], 16, 8)
			if err != nil {
				log.Printf("Error converting value %v, index %v, in Line %v", codes[i], i, l.lineNumber)
				*errors++
				return
			}
			l.codes[i] = byte(value)
		}
	}
}

// localKeyboardReader handles all local keyboard input and interaction
func localKeyboardReader(ando *AndoConnection) {
	cbuf := make([]byte, 128)

	consoleReader := bufio.NewReader(os.Stdin)
	b := make([]byte, 1)
	fmt.Print("Commands:\n\r")
	fmt.Print(" @		- RESET\n\r")
	fmt.Print(" U 9 <CR>	- Quit REMOTE CONTROL\n\r")
	fmt.Print(" U 6 <CR>	- Send data to Eprommer\n\r")
	fmt.Print(" U 7 <CR>	- Receive Data from Eprommer\n\r")
	fmt.Print(" U 8 <CR>	- VERIFY\n\r")
	fmt.Print(" : q		- Quit Ando/Promac EPROM Programmer Communication UI\n\r")
	fmt.Print(" : d		- Download EPROM data (like U7)\n\r")
	fmt.Print("\n\r")
	for ando.continueLoop > 0 {
		fmt.Printf("Command > ")
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
					fmt.Println(" Back to ODT\n\r")
					continue
				}
				if cbuf[0] == 'q' {
					ando.continueLoop = 0
					ando.state = NormalInput
				}
				if cbuf[0] == 'd' {
					fmt.Println("\n\r")
					ando.state = ReceiveData
					bbuf := make([]byte, 8)
					bbuf[0] = 'U'
					bbuf[1] = '7'
					bbuf[2] = '\r'
					ando.serial.tty.Write(bbuf)
				}
				continue
			}

			if ando.state == NormalInput {
				if cbuf[0] == ':' {
					// If ':' is selected, check next char for command to execute
					// We switch state to CommandInput for that
					ando.state = CommandInput
					fmt.Print("Command (:qd):")
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
