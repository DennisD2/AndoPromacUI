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
...
```
...

## Testing
