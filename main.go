package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
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
		F_ASCIIHex, //F_HP64000ABS,F_ASCIIHex, F_GENERIC
		&andoSerial,
		nil,
		0,
		0,
		nil,
		time.Now(),
		time.Now(),
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

var endCriteriaTest = 0

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
			endCriteriaReached := endCriteriaCheck(cbuf)
			if endCriteriaReached {
				if ando.state == ReceiveData {
					// End of download data
					ando.stopTime = time.Now()
					log.Printf("Time spent [s]: %v\n\r", ando.stopTime.Sub(ando.startTime).Seconds())
					if errors > 0 {
						fmt.Printf("There were %v errors on data download\n\r", errors)
						errors = 0
					} else {
						parseFormat(ando, errors, lineNumber)
						if errors > 0 {
							fmt.Printf("There were %v errors during parsing\n\r", errors)
							errors = 0
						} else {
							log.Printf("Data receive completed. Read %v bytes in %v lines/records\n\r", (lineNumber-1)*16, lineNumber-1)
							log.Printf("Checksum calculated: %06x\n\r", ando.checksum)
						}
					}
					lineNumber = 1
				}
				if ando.state == SendData {
					// Device signals that upload was processed complete and without errors
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
				if ando.state == ReceiveData {
					// incoming data during download
					handleGenericInput(ando, num, cbuf, &newLine, &lineNumber, &errors)
				} else {
					// human-readable output, we just print it out
					fmt.Printf("%s", cbuf)
				}
			}
		}
	}
}

// 5b             [
// 50 41 53 53     P A S S
// 5d              ]
func endCriteriaCheck(bytes []byte) bool {
	//Next line: this one lines works with firmware 21.7
	//return strings.HasPrefix(string(str), "[PASS]")
	str := string(bytes)

	switch endCriteriaTest {
	case 0:
		if strings.HasPrefix(str, "[") {
			endCriteriaTest = 1
			if strings.HasPrefix(str, "[P") {
				endCriteriaTest = 2
			}
			if strings.HasPrefix(str, "[PA") {
				endCriteriaTest = 3
			}
			if strings.HasPrefix(str, "[PAS") {
				endCriteriaTest = 4
			}
			if strings.HasPrefix(str, "[PASS") {
				endCriteriaTest = 5
			}
			if strings.HasPrefix(str, "[PASS]") {
				endCriteriaTest = 6
			}
		}
		break
	case 1:
		if strings.HasPrefix(str, "P") {
			endCriteriaTest = 2
		}
		if strings.HasPrefix(str, "PA") {
			endCriteriaTest = 3
		}
		if strings.HasPrefix(str, "PAS") {
			endCriteriaTest = 4
		}
		if strings.HasPrefix(str, "PASS") {
			endCriteriaTest = 5
		}
		if strings.HasPrefix(str, "PASS]") {
			endCriteriaTest = 6
		}
		break
	case 2:
		if strings.HasPrefix(str, "A") {
			endCriteriaTest = 3
		}
		if strings.HasPrefix(str, "AS") {
			endCriteriaTest = 4
		}
		if strings.HasPrefix(str, "ASS") {
			endCriteriaTest = 5
		}
		if strings.HasPrefix(str, "ASS]") {
			endCriteriaTest = 6
		}
		break
	case 3:
		if strings.HasPrefix(str, "S") {
			endCriteriaTest = 4
		}
		if strings.HasPrefix(str, "SS") {
			endCriteriaTest = 5
		}
		if strings.HasPrefix(str, "SS]") {
			endCriteriaTest = 6
		}
		break
	case 4:
		if strings.HasPrefix(str, "S") {
			endCriteriaTest = 5
		}
		if strings.HasPrefix(str, "S]") {
			endCriteriaTest = 6
		}
	default:
		log.Printf("C: UNKNOWN CASE in byte stream (%v)\n\r")
		break
	}

	switch endCriteriaTest {
	case 1:
		log.Printf("C: Found '[' string in byte stream\n\r")
		break
	case 2:
		log.Printf("C: Found '[P' string in byte stream\n\r")
		break
	case 3:
		log.Printf("C: Found '[PA' string in byte stream\n\r")
		break
	case 4:
		log.Printf("C: Found '[PAS' string in byte stream\n\r")
		break
	case 5:
		log.Printf("C: Found '[PASS' string in byte stream\n\r")
		break
	case 6:
		log.Printf("C: Found '[PASS]' string in byte stream\n\r")
		endCriteriaTest = 0
		return true
	}

	return false
}

// parseFormat calls function depending on transfer format
func parseFormat(ando *AndoConnection, errors int, lineNumber int) {
	if ando.transferFormat == F_GENERIC {
		parseGeneric(ando, &errors)
	}
	if ando.transferFormat == F_HP64000ABS {
		initHp64KFormat(ando)
		parseHp64KFormat(ando, &lineNumber, &errors)
	}
	if ando.transferFormat == F_ASCIIHex {
		parseASCIIHexFormat(ando, &lineNumber, &errors)
	}
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
					ando.startTime = time.Now()
					ando.lineInfos = nil
					ando.checksum = 0
					initGenericFormat(ando)
					endCriteriaTest = 0

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
				if cbuf[0] == 'f' {
					if ando.transferFormat == F_ASCIIHex {
						ando.transferFormat = F_HP64000ABS
						setTransferFormat(ando, "HP64000ABS")
						fmt.Println(" File format is now: HP64000ABS\n\r")
					} else if ando.transferFormat == F_HP64000ABS {
						ando.transferFormat = F_GENERIC
						fmt.Println(" File format is now: Generic\n\r")
					} else if ando.transferFormat == F_GENERIC {
						ando.transferFormat = F_ASCIIHex
						setTransferFormat(ando, "ASCII Hex")
						fmt.Println(" File format is now: ASCII-Hex\n\r")
					}
					ando.state = NormalInput
				}
				continue
			}

			if ando.state == NormalInput {
				if cbuf[0] == ':' {
					// If ':' is selected, check next char for command to execute
					// We switch state to CommandInput for that
					ando.state = CommandInput
					fmt.Print(" [:qdwuf] >")
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

// parseFormat calls function depending on transfer format
func uploadFile(ando *AndoConnection) {
	errors := 0
	if ando.transferFormat == F_GENERIC {
		//TBD
	}
	if ando.transferFormat == F_HP64000ABS {
		//TBD
	}
	if ando.transferFormat == F_ASCIIHex {
		uploadFileAsASCIIHex(ando, &errors)
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

	fmt.Print(" R <SPACE>	- outputs selected ROM-TYPE\n\r")
	fmt.Print(" U 5 <SPACE> <CR> - outputs currently selected Data Format\n\r")
	fmt.Print(" U 5 <HEXDIGIT> <CR> - Selected Data Format (Examples: 5=ASCII-Hex, A=HP64000ABS)\n\r")

	fmt.Print("Compound Commands:\n\r")
	fmt.Print(" : q		- Quit Ando/Promac EPROM Programmer Communication UI\n\r")
	fmt.Print(" : d		- Download EPROM data (like U7)\n\r")
	fmt.Printf(" : w		- Write EPROM data to file %v-<checksum>.bin\n\r", ando.downloadFile)
	fmt.Printf(" : u		- Upload EPROM data from file %v to EPrommer\n\r", ando.uploadFile)
	fmt.Printf(" : f		- Change file transfer format (ASCII-Hex, HP64000ABS, GENERIC). Current is: ")
	switch ando.transferFormat {
	case F_GENERIC:
		fmt.Println(" Generic\n\r")
		break
	case F_HP64000ABS:
		fmt.Println("HP64000ABS\n\r")
		break
	case F_ASCIIHex:
		fmt.Println("ASCII-Hex\n\r")
		break
	}
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

// loadFile loads a file from local filesystem
// Returns a byte array and true on error, false if loading was successful
func loadFile(ando *AndoConnection, errors *int) ([]byte, bool) {
	// Read in file
	bytes, err := os.ReadFile(ando.uploadFile)
	if err != nil {
		log.Printf("Error loading input file %s: %s\n\r", ando.uploadFile, err)
		*errors++
		return nil, true
	}
	log.Printf("Loaded input file %s, %v bytes\n", ando.uploadFile, len(bytes))
	return bytes, false
}
