package main

import "fmt"

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
	sof  *StartOfFileRecord
	data *DataRecord
}

// initHp64KFormat initializes required structures
func initHp64KFormat(ando *AndoConnection) {
	var sofRecord = new(StartOfFileRecord)
	var dataRecord = new(DataRecord)
	var hp64k = new(HP64KInfo)
	hp64k.sof = sofRecord
	hp64k.data = dataRecord
	ando.hp64k = hp64k
}

// parseHp64KFormat parses all records in data.
func parseHp64KFormat(ando *AndoConnection, lineNumber *int, errors *int) {
	fmt.Printf("Parsing HP64K format\n\r")

	i := 0
	valid := readSOFRecord(ando, &i, errors)
	//dumpSOFRecord(ando, ando.hp64k.sof)
	if !valid {
		fmt.Printf("Error reading SOF record\n\r")
		return
	}

	for i < len(genericState.rawData) {
		valid := readRecord(ando, &i, errors)
		if !valid {
			if ando.hp64k.data.wordCount == 0 && *errors == 0 {
				fmt.Printf("Reading Data complete\n\r")
			} else {
				if *errors > 0 {
					fmt.Printf("Error reading Data record\n\r")
				}
			}
			return
		} else {
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
		}
	}
}

// readSOFRecord reads Start-Of-File record. Returns true if everything is fine, false on error.
func readSOFRecord(ando *AndoConnection, i *int, errors *int) bool {
	b := genericState.rawData[*i]
	if b != 0x4 {
		fmt.Printf("Illegal wordCount byte with value %v in raw data (value should be always 0x4)\n\r", b)
		*errors++
		return false
	}
	ando.hp64k.sof.wordCount = b

	*i++
	b = genericState.rawData[*i]
	ando.hp64k.sof.dataBusWidth = uint16(b) << 8
	ando.hp64k.sof.checksum += b
	*i++
	b = genericState.rawData[*i]
	ando.hp64k.sof.dataBusWidth += uint16(b)
	ando.hp64k.sof.checksum += b

	*i++
	b = genericState.rawData[*i]
	ando.hp64k.sof.dataWidthBase = uint16(b) << 8
	ando.hp64k.sof.checksum += b
	*i++
	b = genericState.rawData[*i]
	ando.hp64k.sof.dataWidthBase += uint16(b)
	ando.hp64k.sof.checksum += b

	// "Transfer address"
	*i++
	b = genericState.rawData[*i]
	ando.hp64k.sof.transferAddress = uint32(b) << 8
	ando.hp64k.sof.checksum += b
	*i++
	b = genericState.rawData[*i]
	ando.hp64k.sof.transferAddress += uint32(b)
	ando.hp64k.sof.checksum += b
	*i++
	b = genericState.rawData[*i]
	ando.hp64k.sof.transferAddress += uint32(b) << 24
	ando.hp64k.sof.checksum += b
	*i++
	b = genericState.rawData[*i]
	ando.hp64k.sof.transferAddress += uint32(b) << 16
	ando.hp64k.sof.checksum += b

	*i++
	b = genericState.rawData[*i]
	if b != ando.hp64k.sof.checksum {
		fmt.Printf("sof.checksum mismatch 0x%02x!=0x%02xd!\n\r", b, ando.hp64k.sof.checksum)
		*errors++
		return false
	} else {
		if ando.debug >= 1 {
			fmt.Printf("Start-Of-File record checksum ok!\n\r")
		}
	}
	*i++
	return true
}

// readRecord reads a record. value i must point to byte 0 of this record.
// Returns true as long as there are no errors and End-Of-File record was not read.
func readRecord(ando *AndoConnection, i *int, errors *int) bool {
	var b byte

	// init some values
	ando.hp64k.data.checksum = 0
	ando.hp64k.data.bytes = nil

	if !readRecordHeader(ando, i) {
		// End-Of-File record was read
		return false
	}

	// data bytes in record
	dataBytesEnd := *i + int(ando.hp64k.data.byteCount)
	for *i < dataBytesEnd {
		b = genericState.rawData[*i]
		ando.hp64k.data.bytes = append(ando.hp64k.data.bytes, b)
		ando.hp64k.data.checksum += b
		*i++
	}

	// checksum
	b = genericState.rawData[*i]
	if b != ando.hp64k.data.checksum {
		fmt.Printf("data.checksum mismatch read:0x%02x != calculated:0x%02x! pos=%v\n\r", b, ando.hp64k.data.checksum, *i)
		*errors++
		return false
	} else {
		if ando.debug > 2 {
			fmt.Printf("Data record checksum ok!\n\r")
		}
	}

	// move i to byte 0 of next record
	*i++
	return true
}

// readRecordHeader reads header of a record. Returns tue for a common data record and false for the End-Of-File record.
// Cursor value i must point on calling to first byte of header. cursor will point to first byte of next record on exit.
func readRecordHeader(ando *AndoConnection, i *int) bool {
	// wordCount
	b := genericState.rawData[*i]
	ando.hp64k.data.wordCount = b
	if b == 0x0 {
		fmt.Printf("End-Of-File record received\n\r")
		return false
	}
	*i++
	// byteCount
	b = genericState.rawData[*i]
	ando.hp64k.data.byteCount = uint16(b) << 8
	ando.hp64k.data.checksum += b
	*i++
	b = genericState.rawData[*i]
	ando.hp64k.data.byteCount = uint16(b)
	ando.hp64k.data.checksum += b

	// Target address
	*i++
	b = genericState.rawData[*i]
	ando.hp64k.data.targetAddress = uint32(b) << 8
	ando.hp64k.data.checksum += b
	*i++
	b = genericState.rawData[*i]
	ando.hp64k.data.targetAddress += uint32(b)
	ando.hp64k.data.checksum += b
	*i++
	b = genericState.rawData[*i]
	ando.hp64k.data.targetAddress += uint32(b) << 24
	ando.hp64k.data.checksum += b
	*i++
	b = genericState.rawData[*i]
	ando.hp64k.data.targetAddress += uint32(b) << 16
	ando.hp64k.data.checksum += b

	// move i to byte 0 of next record
	*i++
	return true
}

// dumpDataRecord dump a Data record
func dumpDataRecord(ando *AndoConnection, record *DataRecord) {
	if ando.debug > 1 {
		fmt.Printf("\n\rdata.wordCount=%d\n\r", record.wordCount)
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

// dumpSOFRecord dump a Start-Of-File record
func dumpSOFRecord(ando *AndoConnection, record *StartOfFileRecord) {
	if ando.debug > 1 {
		fmt.Printf("sof.wordCount=%d\n\r", record.wordCount)
		fmt.Printf("sof.dataBusWidth=%d\n\r", record.dataBusWidth)
		fmt.Printf("sof.dataWidthBase=%d\n\r", record.dataWidthBase)
		fmt.Printf("sof.transferAddress=0x%04x\n\r", record.transferAddress)
		fmt.Printf("sof.checksum=0x%02x\n\r", record.checksum)
	}
}
