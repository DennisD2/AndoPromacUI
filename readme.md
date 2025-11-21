## AndoPromacUI
Control of vintage Eprommer via textual UI.
Eprommer is accessed via serial cable and a serial<->USB adapter.

Eprommer can be controlled (e.g. the key functions are supported)
and EPROM data can be uploaded and downloaded.

Support for following EPROM Programmers:
* ANDO AF-9704 
* Promac Model 2A 

All tests were made with Ando AF-9704. The Promac Model 2A is 99-100% equal to 
the Ando AF-9704 and should behave the same :-)

## Build
```shell
go build .
```
will create the executable AndoPromacUI.

## Use
```shell
./AndoPromacUI
```

Execution without any arguments will use defaults and prints out some debug info:
```shell
% ./AndoPromacUI 
Ando/Promac EPROM Programmer Communication UI
--device, TTY Device: /dev/ttyUSB0
--dry-run: false
--debug: 0
--baudrate: 19200
--outfile: out-<checksum>.bin
--batch: false (batch mode not yet supported)
--infile: in.bin
Commands:
 @              - RESET
 P A <CR>       - DEVICE-COPY
 P C <CR>       - DEVICE-BLANK
 P D <CR>       - DEVICE-PROGRAM
 P E <CR>       - DEVICE-VERIFY
 U 9 <CR>       - Quit REMOTE CONTROL
 U 6 <CR>       - Send data to EPrommer
 U 7 <CR>       - Receive Data from EPrommer
 U 8 <CR>       - VERIFY
Compound Commands:
 : q            - Quit Ando/Promac EPROM Programmer Communication UI
 : d            - Download EPROM data (like U7)
 : w            - Write EPROM data to file out-<checksum>.bin
 : u            - Upload EPROM data from file in.bin to EPrommer

Command > 
```

During download from EPrommer, a checksum is calculated from all bytes downloaded.
This is an uint32 sum of all byte values in EPROM. The checksum is being used for the
filename for saved EPROM data.
The last 4 digits of the checksum should be identical to checksum from Ando AF-9704
programmer, which is shown after DEVICE->COPY on its display.

## Cabeling
I am using a simple USB<->Serial adapter. See what adaptors I've used to have it working.

![20251118_090940.jpg](docs/20251118_090940.jpg)

## Restrictions
The software uses package "golang.org/x/term" and was only tested with Linux.
I do not know if that package exists for other operating systems,
So software might only run on Linux.
