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
--outfile: out.bin
--batch: false (batch mode not yet supported)
--infile: in.bin
Commands:
 @              - RESET
 U 9 <CR>       - Quit REMOTE CONTROL
 U 6 <CR>       - Send data to Eprommer
 U 7 <CR>       - Receive Data from Eprommer
 U 8 <CR>       - VERIFY
 : q            - Quit Ando/Promac EPROM Programmer Communication UI
 : d            - Download EPROM data (like U7)
 : w            - Write EPROM data to file out.bin
 : u            - Upload EPROM data from file in.bin to EPrommer
```

## Testing
