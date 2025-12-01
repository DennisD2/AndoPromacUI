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

func handleHP64KABSInput(ando *AndoConnection, num int, cbuf []byte, newLine *LineInfo, lineNumber *int, errors *int) {
	for i := 0; i < num; i++ {
		b := uint8(cbuf[i])
		fmt.Printf("%02x ", b)

		if ando.hp64k.state == HP64K_SOF {
			handleSOFRecord(ando, b)

		} else if ando.hp64k.state == HP64K_Data_Header || ando.hp64k.state == HP64K_Data || ando.hp64k.state == HP64K_Checksum {
			handleRecordData(ando, b)
		}

	}
	fmt.Printf("\n\r")
}

func handleRecordData(ando *AndoConnection, b uint8) {
	if ando.hp64k.state == HP64K_Data_Header {
		if ando.transferPosition == 0 {
			ando.hp64k.data.wordCount = b
			ando.hp64k.data.checksum = 0
			ando.hp64k.data.bytes = nil
		}
		if ando.transferPosition == 1 {
			ando.hp64k.data.byteCount = uint16(b << 8)
			ando.hp64k.data.checksum += b
		}
		if ando.transferPosition == 2 {
			ando.hp64k.data.byteCount = uint16(b)
			ando.hp64k.data.checksum += b
		}

		// "Target address"
		if ando.transferPosition == 3 {
			ando.hp64k.data.targetAddress = uint32(b) << 8
			ando.hp64k.data.checksum += b
		}
		if ando.transferPosition == 4 {
			ando.hp64k.data.targetAddress += uint32(b)
			ando.hp64k.data.checksum += b
		}
		if ando.transferPosition == 5 {
			ando.hp64k.data.targetAddress += uint32(b) << 24
			ando.hp64k.data.checksum += b
		}
		if ando.transferPosition == 6 {
			ando.hp64k.data.targetAddress += uint32(b) << 16
			ando.hp64k.data.checksum += b

			ando.hp64k.data.bytePos = 0

			// next state
			ando.hp64k.state = HP64K_Data
			ando.transferPosition = 0
		}

		if ando.hp64k.state == HP64K_Data_Header {
			ando.transferPosition++
		}
	} else if ando.hp64k.state == HP64K_Data {
		ando.hp64k.data.bytes = append(ando.hp64k.data.bytes, b)
		ando.hp64k.data.checksum += b
		ando.hp64k.data.bytePos++
		if ando.hp64k.data.bytePos == ando.hp64k.data.byteCount {
			ando.hp64k.state = HP64K_Checksum
			ando.transferPosition++
		}
	} else if ando.hp64k.state == HP64K_Checksum {
		checksum := b
		if checksum != ando.hp64k.data.checksum {
			fmt.Printf("data.checksum mismatch 0x%02x!=0x%02xd!\n\r", checksum, ando.hp64k.data.checksum)
		} else {
			//fmt.Printf("Data record checksum ok!\n\r")
		}

		DumpRecordData(ando, ando.hp64k.data)

		ando.transferPosition = 0
		ando.hp64k.state = HP64K_Data_Header
	}
}

func DumpRecordData(ando *AndoConnection, record *DataRecord) {
	fmt.Printf("data.wordCount=%d\n\r", record.wordCount)
	fmt.Printf("data.byteCount=%d\n\r", record.byteCount)
	fmt.Printf("data.targetAddress=0x%04x\n\r", record.targetAddress)
	fmt.Printf("data.bytes=[")
	for _, b := range record.bytes {
		fmt.Printf("%02x ", b)
	}
	fmt.Printf("]\n\r")
	fmt.Printf("data.checksum=0x%02x\n\r", record.checksum)
}

func handleSOFRecord(ando *AndoConnection, b uint8) {
	if ando.transferPosition == 0 {
		ando.hp64k.sof.wordCount = b
		ando.hp64k.sof.checksum = 0
	}
	if ando.transferPosition == 1 {
		ando.hp64k.sof.dataBusWidth = uint16(b) << 8
		ando.hp64k.sof.checksum += b
	}
	if ando.transferPosition == 2 {
		ando.hp64k.sof.dataBusWidth += uint16(b)
		ando.hp64k.sof.checksum += b
	}
	if ando.transferPosition == 3 {
		ando.hp64k.sof.dataWidthBase = uint16(b) << 8
		ando.hp64k.sof.checksum += b
	}
	if ando.transferPosition == 4 {
		ando.hp64k.sof.dataWidthBase += uint16(b)
		ando.hp64k.sof.checksum += b
	}

	// "Transfer address"
	if ando.transferPosition == 5 {
		ando.hp64k.sof.transferAddress = uint32(b) << 8
		ando.hp64k.sof.checksum += b
	}
	if ando.transferPosition == 6 {
		ando.hp64k.sof.transferAddress += uint32(b)
		ando.hp64k.sof.checksum += b
	}
	if ando.transferPosition == 7 {
		ando.hp64k.sof.transferAddress += uint32(b) << 24
		ando.hp64k.sof.checksum += b
	}
	if ando.transferPosition == 8 {
		ando.hp64k.sof.transferAddress += uint32(b) << 16
		ando.hp64k.sof.checksum += b
	}

	if ando.transferPosition == 9 {
		checksum := b
		if checksum != ando.hp64k.sof.checksum {
			fmt.Printf("sof.checksum mismatch 0x%02x!=0x%02xd!\n\r", checksum, ando.hp64k.sof.checksum)
		} else {
			fmt.Printf("Start-Of-File record checksum ok!\n\r")
		}

		dumpSofRecord(ando.hp64k.sof)
		ando.hp64k.state = HP64K_Data_Header
		ando.transferPosition = 0 // reset for next record
	}
	if ando.hp64k.state == HP64K_SOF {
		// move pointer forward
		ando.transferPosition++
	}

}

func dumpSofRecord(record *StartOfFileRecord) {
	fmt.Printf("sof.wordCount=%d\n\r", record.wordCount)
	fmt.Printf("sof.dataBusWidth=%d\n\r", record.dataBusWidth)
	fmt.Printf("sof.dataWidthBase=%d\n\r", record.dataWidthBase)
	fmt.Printf("sof.transferAddress=0x%04x\n\r", record.transferAddress)
	fmt.Printf("sof.checksum=0x%02x\n\r", record.checksum)
}
