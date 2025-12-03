package main

import "fmt"

type HP64State int

const (
	HP64K_SOF         HP64State = 0
	HP64K_Data_Header           = 1
	HP64K_Data                  = 2
	HP64K_Checksum              = 3
	HP64K_EOF                   = 4
)

type StartOfFileRecord struct {
	wordCount       uint8
	dataBusWidth    uint16
	dataWidthBase   uint16
	transferAddress uint32
	checksum        uint8
}

type DataRecord struct {
	wordCount     uint8
	byteCount     uint16
	targetAddress uint32
	bytes         []byte
	bytePos       uint16
	checksum      uint8
}

type HP64KInfo struct {
	state HP64State
	sof   *StartOfFileRecord
	data  *DataRecord
}

// initHp64KFormat initializes required structures
func initHp64KFormat(ando *AndoConnection) {
	var sofRecord = new(StartOfFileRecord)
	var dataRecord = new(DataRecord)
	var hp64k = new(HP64KInfo)
	hp64k.sof = sofRecord
	hp64k.data = dataRecord
	hp64k.state = HP64K_SOF
	ando.hp64k = hp64k
}

func parseHp64KFormat(ando *AndoConnection) {

}

// handleHP64KABSInput handles HP64000 format
func handleHP64KABSInput(ando *AndoConnection, num int, cbuf []byte, newLine *LineInfo, lineNumber *int, errors *int) {
	for i := 0; i < num; i++ {
		b := uint8(cbuf[i])
		if ando.debug > 1 {
			fmt.Printf("%02x ", b)
		}
		if ando.hp64k.state == HP64K_SOF {
			handleSOFRecord(ando, b, errors)

		} else if ando.hp64k.state == HP64K_Data_Header || ando.hp64k.state == HP64K_Data || ando.hp64k.state == HP64K_Checksum {
			handleRecordData(ando, b, lineNumber, errors)
		}

	}
	if ando.debug > 1 {
		fmt.Printf("\n\r")
	}
}

// handleRecordData handle Data record
func handleRecordData(ando *AndoConnection, b uint8, lineNumber *int, errors *int) {
	if ando.hp64k.state == HP64K_Data_Header {
		if ando.recordPosition == 0 {
			ando.hp64k.data.wordCount = b
			ando.hp64k.data.checksum = 0
			ando.hp64k.data.bytes = nil

			if b == 0x0 {
				ando.hp64k.state = HP64K_EOF
				fmt.Printf("End-Of-File record received\n\r")
			}
		}
		if ando.recordPosition == 1 {
			ando.hp64k.data.byteCount = uint16(b << 8)
			ando.hp64k.data.checksum += b
		}
		if ando.recordPosition == 2 {
			ando.hp64k.data.byteCount = uint16(b)
			ando.hp64k.data.checksum += b
		}

		// "Target address"
		if ando.recordPosition == 3 {
			ando.hp64k.data.targetAddress = uint32(b) << 8
			ando.hp64k.data.checksum += b
		}
		if ando.recordPosition == 4 {
			ando.hp64k.data.targetAddress += uint32(b)
			ando.hp64k.data.checksum += b
		}
		if ando.recordPosition == 5 {
			ando.hp64k.data.targetAddress += uint32(b) << 24
			ando.hp64k.data.checksum += b
		}
		if ando.recordPosition == 6 {
			ando.hp64k.data.targetAddress += uint32(b) << 16
			ando.hp64k.data.checksum += b

			ando.hp64k.data.bytePos = 0

			// next state
			ando.hp64k.state = HP64K_Data
			ando.recordPosition = 0
		}

		if ando.hp64k.state == HP64K_Data_Header {
			ando.recordPosition++
		}
	} else if ando.hp64k.state == HP64K_Data {
		ando.hp64k.data.bytes = append(ando.hp64k.data.bytes, b)
		ando.hp64k.data.checksum += b
		ando.hp64k.data.bytePos++
		if ando.hp64k.data.bytePos == ando.hp64k.data.byteCount {
			ando.hp64k.state = HP64K_Checksum
			ando.recordPosition++
		}
	} else if ando.hp64k.state == HP64K_Checksum {
		checksum := b
		if checksum != ando.hp64k.data.checksum {
			fmt.Printf("data.checksum mismatch 0x%02x!=0x%02xd!\n\r", checksum, ando.hp64k.data.checksum)
			*errors++
		} else {
			if ando.debug > 2 {
				fmt.Printf("Data record checksum ok!\n\r")
			}
		}

		dumpDataRecord(ando, ando.hp64k.data)

		newLine := LineInfo{
			address: ando.hp64k.data.targetAddress,
		}
		for i, b := range ando.hp64k.data.bytes {
			newLine.codes[i] = b
			ando.checksum += uint32(b)
		}
		ando.lineInfos = append(ando.lineInfos, newLine)
		*lineNumber++

		ando.recordPosition = 0
		ando.hp64k.state = HP64K_Data_Header
	}
}

// dumpDataRecord dump a Data record
func dumpDataRecord(ando *AndoConnection, record *DataRecord) {
	if ando.debug > 1 {
		fmt.Printf("data.wordCount=%d\n\r", record.wordCount)
		fmt.Printf("data.byteCount=%d\n\r", record.byteCount)
	}
	fmt.Printf("0x%08x: ", record.targetAddress)
	for _, b := range record.bytes {
		fmt.Printf("%02x ", b)
	}
	fmt.Printf("\n\r")
	if ando.debug > 1 {
		fmt.Printf("data.checksum=0x%02x\n\r", record.checksum)
	}
}

// handleSOFRecord handle Start-Of-File record
func handleSOFRecord(ando *AndoConnection, b uint8, errors *int) {
	if ando.recordPosition == 0 {
		ando.hp64k.sof.wordCount = b
		ando.hp64k.sof.checksum = 0
	}
	if ando.recordPosition == 1 {
		ando.hp64k.sof.dataBusWidth = uint16(b) << 8
		ando.hp64k.sof.checksum += b
	}
	if ando.recordPosition == 2 {
		ando.hp64k.sof.dataBusWidth += uint16(b)
		ando.hp64k.sof.checksum += b
	}
	if ando.recordPosition == 3 {
		ando.hp64k.sof.dataWidthBase = uint16(b) << 8
		ando.hp64k.sof.checksum += b
	}
	if ando.recordPosition == 4 {
		ando.hp64k.sof.dataWidthBase += uint16(b)
		ando.hp64k.sof.checksum += b
	}

	// "Transfer address"
	if ando.recordPosition == 5 {
		ando.hp64k.sof.transferAddress = uint32(b) << 8
		ando.hp64k.sof.checksum += b
	}
	if ando.recordPosition == 6 {
		ando.hp64k.sof.transferAddress += uint32(b)
		ando.hp64k.sof.checksum += b
	}
	if ando.recordPosition == 7 {
		ando.hp64k.sof.transferAddress += uint32(b) << 24
		ando.hp64k.sof.checksum += b
	}
	if ando.recordPosition == 8 {
		ando.hp64k.sof.transferAddress += uint32(b) << 16
		ando.hp64k.sof.checksum += b
	}

	if ando.recordPosition == 9 {
		checksum := b
		if checksum != ando.hp64k.sof.checksum {
			fmt.Printf("sof.checksum mismatch 0x%02x!=0x%02xd!\n\r", checksum, ando.hp64k.sof.checksum)
			*errors++
		} else {
			if ando.debug >= 1 {
				fmt.Printf("Start-Of-File record checksum ok!\n\r")
			}
		}

		dumpSOFRecord(ando, ando.hp64k.sof)

		// set up vars for next record
		ando.hp64k.state = HP64K_Data_Header
		ando.recordPosition = 0
	}
	if ando.hp64k.state == HP64K_SOF {
		// move pointer forward
		ando.recordPosition++
	}

}

// dumpSOFRecord dumps a Start-Of-File record
func dumpSOFRecord(ando *AndoConnection, record *StartOfFileRecord) {
	if ando.debug > 1 {
		fmt.Printf("sof.wordCount=%d\n\r", record.wordCount)
		fmt.Printf("sof.dataBusWidth=%d\n\r", record.dataBusWidth)
		fmt.Printf("sof.dataWidthBase=%d\n\r", record.dataWidthBase)
		fmt.Printf("sof.transferAddress=0x%04x\n\r", record.transferAddress)
		fmt.Printf("sof.checksum=0x%02x\n\r", record.checksum)
	}
}
