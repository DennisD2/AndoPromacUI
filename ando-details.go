package main

import (
	"fmt"
	"log"
)

type DataFormat struct {
	id   byte
	name string
	info string
}

var dataFormats = []DataFormat{
	DataFormat{
		id:   0,
		name: "Intellec",
		info: "Subformat end char required",
	},
	DataFormat{
		id:   1,
		name: "Motorola",
		info: "",
	},
	DataFormat{
		id:   2,
		name: "Tektronix",
		info: "",
	},
	DataFormat{
		id:   '5',
		name: "ASCII Hex",
		info: "Subformat end char required",
	},
	DataFormat{
		id:   6,
		name: "DG Binary",
		info: "",
	},
	DataFormat{
		id:   7,
		name: "DEC Binary",
		info: "",
	},
	DataFormat{
		id:   8,
		name: "Ex TekHex",
		info: "",
	},
	DataFormat{
		id:   9,
		name: "ASM86-Hex",
		info: "Subformat end char required",
	},
	DataFormat{
		id:   'A',
		name: "HP64000ABS",
		info: "",
	},
	DataFormat{
		id:   0xb,
		name: "JEDEC",
		info: "",
	},
	DataFormat{
		id:   0xc,
		name: "Dump-List",
		info: "",
	},
}

func setTransferFormat(ando *AndoConnection, name string) bool {
	var id byte = 0x0
	for _, f := range dataFormats {
		if name == f.name {
			id = f.id
		}
	}
	if id == 0x0 {
		log.Printf("Can't find transfer format named %v", name)
		return false
	}
	fmt.Printf("Setting transfer format named %v to '%c'\n", name, id)

	bbuf := make([]byte, 8)
	bbuf[0] = 'U'
	bbuf[1] = '5'
	bbuf[2] = id
	//bbuf[3] = ' ' // subtype, seems not to work like this
	//bbuf[4] = 0x1a
	bbuf[3] = '\r'
	ando.serial.tty.Write(bbuf)

	return true
}
