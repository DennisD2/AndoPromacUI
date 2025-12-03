
### ASCII-Hex format
TBD

### HP64000 ABS OBJ format 
Binary format.

Format is:
* start-of-file record
* some number of data records
* end-of-file record

#### start-of-file record
Data file begins with start-of-file record. 

| Size in bytes | Value        | Meaning                                                                                              |
|---------------|--------------|------------------------------------------------------------------------------------------------------|
| 1             | 4 (always)   | word count - # of 16-bit words in record (w/o checksum and this byte itself  , so 8 bytes = 4 words) |
| 2             | (calculated) | data bus width                                                                                       |
| 2             | (calculated) | data width base                                                                                      |
| 4             | (calculated) | transfer address                                                                                     |
| 1             | (calculated) | checksum : modulo 256 sum of all bytes in record except the first byte                               |

#### data record
Each data record is build like following description:

| Size in bytes | Value        | Meaning                                                                                              |
|---------------|--------------|------------------------------------------------------------------------------------------------------|
| 1             | (calculated) | word count - # of 16-bit words in record (w/o checksum and this byte itself  , so 8 bytes = 4 words) |
| 2             | (calculated) | byte count - # of 8-bit data bytes                                                                   |
| 4             | (calculated) | address where following data is to be stored                                                         |
| n             | (calculated) | n data bytes                                                                                         |
| 1             | (calculated) | checksum : modulo 256 sum of all bytes in record except the first byte                               |

#### End-of-file record
End of file record has 1 byte size, byte value is zero.

| Size in bytes | Value      | Meaning                                          |
|---------------|------------|--------------------------------------------------|
| 1             | 0 (always) | End-of-file record has word count 0 |

Order of bytes in 32-bit addresses:

| Offset in file | Meaning  |
|----------------|----------|
| 0              | 2nd byte |
| 1              | LSB      |
| 2              | MSB      |
| 3              | 3rd byte |

To be constructed 32 bit value:
 MSB.3rd.2nd.LSB

From Promac manual additional infos on the format:
- eight address bits (4 bytes)
- data bus width = data word width = 8
- understand no end-of-file record (???, manual page 28)

So start-of-file looks like this:

| Size in bytes | Value        | Meaning                                                                                              |
|---------------|--------------|------------------------------------------------------------------------------------------------------|
| 1             | 4 (always)   | word count - # of 16-bit words in record (w/o checksum and this byte itself  , so 8 bytes = 4 words) |
| 2             | 8            | data bus width                                                                                       |
| 2             | 8            | data width base                                                                                      |
| 4             | ?            | transfer address                                                                                     |
| 1             | (calculated) | checksum : modulo 256 sum of all bytes in record except the first byte                               |

I found that Ando Eprommer sends always records with 16 data bytes inside.

File format infos were taken from some other eprom programmer manual:

HP64000 ABS OBJ Format:
![hp64000-abs-obj-format-from-boardsite4000-manual.png](docs/hp64000-abs-obj-format-from-boardsite4000-manual.png)

Even more info:
* https://srecord.sourceforge.net/man/man5/srec_hp64k.5.html

Example of downloaded HP64000 ABS format:
```text
04 00 08 00 08 00 00 00 00 10 

0b 00 10 00 00 00 00 20 6d 86 ff b7 01 0a 20 66 7f 01 
0a 20 61 bd d3 05 

0b 00 10 00 10 00 00 17 2b 5f 24 10 bd c8 15 bd ea 24 bd d3 a4 bd d3 1e 0b 
00 10 00 20 00 00 17 2b 4f 25 4d c1 43 27 46 c1 50 26 03 5f 20 38 95 0b 00 10 00 30 00 00 bd 
c8 1b 2b 3d b7 01 0b bd d3 17 2b 35 24 10 bd 03 0b 00 10 00 40 00 00 c8 15 bd ea 24 bd d3 a4 
bd d3 17 2b 25 25 23 c1 2c 0b 00 10 00 50 00 00 43 27 1c bd c8 1b 2b 1a f6 01 0b 58 58 58 
58 1b 48 0b 00 10 00 60 00 00 81 63 22 0e 16 bd d9 8b f7 01 09 53 bd f0 d9 7e 13 0b 00 10 00 
70 00 00 c8 03 c6 12 bd e2 78 20 f6 c6 01 20 02 c6 02 f7 f8 0b 00 10 00 80 00 00 01 98 20 eb 
7f 01 98 20 e6 86 01 9a 46 97 46 b6 4c 0b 00 10 00 90 00 00 01 08 26 4e bd e2 4f 20 49 86 
fe 94 46 97 46 20 cf 0b 00 10 00 a0 00 00 41 4f 20 22 86 01 20 1e 86 02 20 1a 86 03 20 16 c8 
0b 00 10 00 b0 00 00 86 04 20 12 86 05 20 0e 86 06 20 0a 86 07 20 06 9e 0b 00 10 00 c0 00 00 
86 08 20 02 86 09 bd dd dd 20 17 c6 01 86 c0 20 ea 0b 00 10 00 d0 00 00 0a c6 02 86 a0 20 04 
c6 04 86 90 7f 01 03 d1 35 65 0b 00 10 00 e0 00 00 26 4f 7e c8 03 7d 02 75 26 03 7e d1 a2 
bd d3 17 63 0b 00 10 00 f0 00 00 2b 44 24 18 c6 8f bd e2 8d bd c8 15 bd ea 24 bd 4e 0b 00 10 
01 00 00 00 e2 32 bd d3 a4 bd d3 17 2b 2c 25 2a d1 43 27 d2 b3 0b 00 10 01 10 00 00 bd c8 1b 
2b 21 81 04 22 1d 4d 27 1a b1 01 03 27 3b 0b 00 10 01 20 00 00 c1 b7 01 03 16 ce dd 94 bd f4 
28 de 8e c6 08 a6 bb 0b 00 10 01 30 00 00 00 bd e2 b9 20 ac c6 03 7e d2 40 bd dc eb bd d3 
d2 0b 00 10 01 40 00 00 17 2b 4c 24 10 bd c8 15 bd ea 24 bd d3 a4 bd d3 3c 0b 00 10 01 50 00 
00 17 2b 3c 25 3a bd c8 1b 2b 35 d6 58 c1 11 26 0b 6f 0b 00 10 01 60 00 00 97 58 c6 20 d7 59 
bd e6 4e 20 d3 97 59 58 58 58 52 0b 00 10 01 70 00 00 58 1b 97 08 bd e6 4e c6 02 d7 0b f6 
eb cb bd de 75 0b 00 10 01 80 00 00 46 f6 eb ce bd e2 3b bd ea 24 bd c8 12 20 45 c6 ed 0b 00 
...
00 97 5e 96 00 80 10 27 15 97 00 87 0b 00 10 0f 70 00 00 b6 98 08 88 02 b7 98 08 88 02 b7 98 
08 7e df 4c 50 0b 00 10 0f 80 00 00 7f 00 5e b6 98 08 88 01 b7 98 08 f6 98 0b c8 05 18 0b 00 
10 0f 90 00 00 f7 98 0b 86 ff b7 98 0a ca 04 f7 98 0b bd df a5 d0 0b 00 10 0f a0 00 00 39 
c6 10 20 02 c6 20 d7 3e f6 eb d0 d7 40 c6 80 f9 0b 00 10 0f b0 00 00 d7 3d bd eb 60 39 d6 5e 
17 84 0f 81 03 22 1c c1 85 0b 00 10 0f c0 00 00 12 27 11 c4 f0 c1 30 27 0d 2e 06 c1 20 26 0a 
4d 94 0b 00 10 0f d0 00 00 39 8b 06 39 4f 39 8b 03 39 d6 5e 86 ff 39 96 2e f7 0b 00 10 0f e0 
00 00 bd e0 bb d7 05 bd e0 ab 86 07 d6 05 c1 01 2f 07 db 0b 00 10 0f f0 00 00 bd fe 12 bd 
e2 53 39 2d 07 df 90 ce 02 76 20 05 
15 00 

```
## DG-Binary