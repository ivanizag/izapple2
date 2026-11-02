# Serial Ports

The emulator has two kinds of serial ports:

- The Super Serial Card, `ssc`, a card for any slot of the Apple ][, ][+ and //e.
  Usually on slot 1 for a printer and on slot 2 for a modem.
- The two serial ports built in the Apple IIc, `iicserial`, always on slots 1 and
  2. Port 1 is the printer port and port 2 the modem port.

Both have a 6551 ACIA, the same chip with the same registers, and run the
original firmware. What the Apple sends goes to a device of the host, and what
the device sends is received by the Apple.

## Devices

The `device` parameter chooses what is on the other side of the cable:

| Device    | Sends to                    | Receives from        | Ready                      |
|-----------|-----------------------------|----------------------|----------------------------|
| `none`    | Nothing                     | Nothing              | Never                      |
| `console` | The standard output         | The standard input   | Always                     |
| `file`    | A file, appending to it     | Nothing              | Always                     |
| `tcp`     | The TCP connection          | The TCP connection   | While there is a connection |
| `netmodem`| A modem, to the internet    | The modem            | Always                     |

- `none` is the default: nothing connected, as with an empty port or a cable
  with nothing on the other side. The firmware waits forever to send, so a `PR#`
  to it hangs the machine, as on the real one. Use Ctrl-Reset to get out.
- `console` treats the bytes as text. On output the high bit is removed and the
  carriage returns are new lines. On input the new lines are carriage returns.
  The terminal sends a line once Return is pressed. It is not a good match for
  the `console` frontend, which also uses the standard input and output.
- `file` writes the bytes as they are sent. Name the file with `file`, the
  default is `serial<slot>.out`: `serial1.out` for the slot 1.
- `tcp` listens on the TCP port given with `port`. The default is 1976 plus the
  slot: 1977 for the slot 1 and 1978 for the slot 2. It takes one connection at
  a time, and a second one waits until the first is closed. What is sent while
  nobody is connected is lost. It listens on all the network interfaces, not
  only on `localhost`.
- `netmodem` is a Hayes modem that dials to the internet instead of the phone
  network, as the WiFi modems for retro computers do: `ATDT host:port`. See
  [Connecting to a BBS](#connecting-to-a-bbs).

"Ready" is the handshake line the firmware checks before sending each
character: DCD and DSR on the Super Serial Card, pin 5 of the connector on the
IIc.

## The Apple IIc

The ports are on `-s1` and `-s2`, always with `iicserial`. They can't be
removed:

```bash
# The modem port listening on TCP port 1978
izapple2 -model 2c -s2 iicserial,device=tcp

# The printer port to a file
izapple2 -model 2c -s1 iicserial,device=file,file=printer.txt
```

The IIc has no switches. The ports start as the firmware sets them up:

| Port | Baud rate | Data | Stop | Line feed after CR | Command character |
|------|-----------|------|------|--------------------|-------------------|
| 1    | 9600      | 8    | 2    | Yes                | Ctrl-I            |
| 2    | 300       | 8    | 1    | No                 | Ctrl-A            |

The rest is changed with the commands of the firmware, see below.

The keyboard buffering of the IIc works too. The firmware takes the keys with
an interrupt from the ACIA of port 2, as described in Appendix E of the Apple
IIc Technical Reference Manual. The firmware turns it on in the terminal mode.

The terminal mode of the modem port does not work with ROM 3, the keys typed
come out repeated. ROM 4 fixed "a problem when using the serial modem port in
terminal mode", use the model `2c`.

## The Super Serial Card

```bash
# On slot 2, listening on TCP port 1978
izapple2 -s2 ssc,device=tcp

# A printer on slot 1 writing to a file
izapple2 -s1 ssc,mode=printer,device=file,file=printer.txt
```

The DIP switches of the card are parameters, they are what the firmware reads
when it starts:

| Parameter    | Switches       | Values                                                          | Default          |
|--------------|----------------|-----------------------------------------------------------------|------------------|
| `baud`       | SW1-1 to SW1-4 | 50, 75, 110, 135, 150, 300, 600, 1200, 1800, 2400, 3600, 4800, 7200, 9600, 19200 | 9600 |
| `mode`       | SW1-5, SW1-6   | `communications`, `printer`, `sicp8`, `sicp8a`                   | `communications` |
| `stopbits`   | SW2-1          | 1 or 2                                                          | 1                |
| `databits`   | SW2-2          | 7 or 8, in the communications mode                              | 8                |
| `parity`     | SW2-3, SW2-4   | `none`, `odd`, `even`, in the communications mode               | `none`           |
| `crdelay`    | SW2-2          | Wait 32 ms after each CR, in the printer mode                   | false            |
| `width`      | SW2-3, SW2-4   | 40, 72, 80, 132, in the printer mode                            | 80               |
| `linefeed`   | SW2-5          | Add a line feed after each CR                                   | false            |
| `interrupts` | SW2-6          | Wire the interrupts of the card to the slot                     | true             |

The `sicp8` and `sicp8a` modes emulate the older Apple Serial Interface Card.
The command character is Ctrl-A in the communications mode and Ctrl-I in the
others.

## The Firmware Commands

The ports are configured from BASIC with commands sent through them: the
command character followed by the command. The firmware takes them out of
what is printed, so they are not sent. Some useful ones, the same on both:

| Command | What it does                                                    |
|---------|-----------------------------------------------------------------|
| `nB`    | Baud rate, `n` from 1 to 15: 6 is 300, 8 is 1200, 14 is 9600, 15 is 19200 |
| `nD`    | Data format: 0 is 8 data and 1 stop bits, 1 is 7 and 1, 4 is 8 and 2, 5 is 7 and 2 |
| `nP`    | Parity: 0 none, 1 odd, 3 even                                   |
| `L E`   | Add a line feed after each carriage return. `L D` stops it      |
| `T`     | Terminal mode, see below. `Q` quits it                          |
| `R`     | Reset the port, back to the screen and the keyboard             |

On the Super Serial Card keep the space in the commands with `E` or `D`: the
firmware skips the letters up to the space, and `LE` does nothing. The IIc
takes both.

The easiest way to send them is from a `PRINT` while the output goes to the
port. For port 2 or a card on slot 2, with Ctrl-A:

```
PR#2
PRINT CHR$(1);"1D":PRINT CHR$(1);"L E"
PR#0
```

Typed at the prompt they work as well, but BASIC also gets the line and
answers with a `?SYNTAX ERROR`.

The complete list is in the "Apple IIc Technical Reference Manual", chapters 7
and 8, and in the "Super Serial Card Installation and Operating Manual".

## Plain Text on the Host

With 8 data bits the Apple sends the characters with the high bit set, as they
are in its memory. A real port does the same, so a terminal on the other end
shows them as strange symbols. Two settings give plain text:

- **7 data bits**, the command `1D`. The ACIA drops the eighth bit.
- **A line feed after each carriage return**, the command `L E`. The Apple ends
  the lines with a carriage return alone, and without a line feed the terminal
  writes each line over the last one.

On the Super Serial Card both can be set with the switches instead, from the
start:

```bash
izapple2 -s2 ssc,device=tcp,databits=7,linefeed
```

The `console` device removes the high bit and turns the carriage returns into
new lines on its own. The `file` device keeps the bytes as they are, to see the
text of a file written with 8 bits:

```bash
tr '\200-\377' '\000-\177' < printer.txt
```

## Trying It With `nc`

Start the emulator with a TCP port, here the modem port of a IIc:

```bash
izapple2 -model 2c -s2 iicserial,device=tcp
```

And connect from another terminal:

```bash
nc -c localhost 1978
```

`-c` makes `nc` on macOS end the lines with CR LF, Applesoft needs the carriage
return to take a line. On Linux it is `-C`. `telnet localhost 1978` also works.

### Input

Type on the emulator:

```
IN#2
```

Now the lines typed on `nc` go to Applesoft: `PRINT 6*7` shows `42` on the
screen. The keyboard of the emulator still works. `IN#0` goes back to the
keyboard only.

### Output

Type on the emulator:

```
PR#2
PRINT CHR$(1);"1D":PRINT CHR$(1);"L E"
PRINT "HELLO"
```

`HELLO`, and everything printed after it, shows on `nc`. `PR#0` goes back to
the screen.

### Terminal Mode

In terminal mode the Apple is a terminal for the other side: the keys go to
`nc` and what comes from `nc` shows on the screen. Set up the port first, then
enter the terminal mode with Ctrl-A `T`:

```
IN#2
PR#2
PRINT CHR$(1);"1D":PRINT CHR$(1);"L E"
PR#0
```

Then Ctrl-A and `T` on the IIc. On the Super Serial Card, Ctrl-A shows the
`APPLE SSC:` prompt: type `T` and Return. Ctrl-A `Q` quits the terminal mode.

The order matters on the IIc: `IN#2` sets up the port again, and the commands
sent before are lost.

## Connecting to a BBS

The `netmodem` device is a modem that takes the Hayes AT commands, and dials a
host and a TCP port instead of a phone number. It is how the WiFi modems connect
the retro computers to the BBSes on the internet. The terminal mode of the
firmware works with it, and so should the terminal programs for the Apple II
that drive a Hayes modem, none has been tried yet.

To call the BBS at `cqbbs.ddns.net`, port 6502, from the modem port of a IIc:

```bash
izapple2 -model 2c -s2 iicserial,device=netmodem
```

On the Apple, start the terminal mode and set up the port. Each line is the
command character, Ctrl-A, followed by the command:

```
IN#2
Ctrl-A T       the terminal mode
Ctrl-A 1D      7 data bits, plain text for the BBS
Ctrl-A 14B     9600 baud, port 2 starts at 300
```

Then talk to the modem:

```
AT             answers OK
ATDT cqbbs.ddns.net:6502
```

The modem answers `CONNECT` and the BBS takes over. The keys go to the BBS and
what it sends shows on the screen.

To hang up, wait a second, type `+++`, and wait another second: the modem
answers `OK` and takes commands again, with the connection still open. `ATH`
hangs up and `ATO` goes back to the BBS. When the BBS hangs up the modem answers
`NO CARRIER`. Ctrl-A `Q` quits the terminal mode.

With a Super Serial Card the setup is with the switches, and the terminal mode
is Ctrl-A, `T` and Return at the `APPLE SSC:` prompt:

```bash
izapple2 -s2 ssc,device=netmodem,databits=7
```

### The AT Commands

The commands start with `AT` and end with Return. More than one can go on a
line: `ATE0V1`.

| Command              | What it does                                                    |
|----------------------|-----------------------------------------------------------------|
| `ATDT host:port`     | Dials with telnet. The port defaults to 23                       |
| `ATD host:port`      | Dials without telnet, the bytes pass as they are. `ATDP` too     |
| `+++`                | With a second of silence before and after: back to the commands, still connected |
| `ATO`                | Back online                                                     |
| `ATH`                | Hangs up                                                        |
| `ATZ`, `AT&F`        | Hangs up and goes back to the defaults                           |
| `A/`                 | Repeats the last command, without Return                        |
| `ATE0`, `ATE1`       | Echo of the commands off and on, on by default                   |
| `ATV0`, `ATV1`       | Results as numbers or in words, in words by default              |
| `ATQ0`, `ATQ1`       | Results on, or quiet                                            |
| `ATI`                | Information                                                     |
| `ATSn=v`, `ATSn?`    | Sets and shows the register `n`                                 |

The host and port can also go between quotes, `ATDT"cqbbs.ddns.net:6502"`, as
on Zimodem. The results are `OK` (0), `CONNECT` (1), `NO CARRIER` (3) and
`ERROR` (4). Any key while dialing cancels the call. The commands of the speaker
and the flow control, `L`, `M`, `&C`, `&D`, `&K` and the like, are accepted and
do nothing.

`CONNECT` comes with the speed of the connection, the baud rate the port is
set to, as `CONNECT 9600`: the terminal programs set the port to it, ProTERM
goes down to 300 baud with `CONNECT` alone. As on the Hayes modems, it is
alone at 300 baud, or with `ATX0`. As numbers, with `ATV0`, `CONNECT 1200` is
5 and `CONNECT 2400` is 10, as on the Hayes Smartmodem 2400, and the faster
speeds are 1.

The registers used are `S2`, the escape character, `+`; `S3`, `S4` and `S5`, the
carriage return, line feed and backspace; `S7`, the seconds to wait for the
connection, 50; and `S12`, the guard time of the escape in 1/50 of a second, 50.

With telnet, `ATDT`, the modem answers the options the BBS asks for: it lets the
BBS do the echo and suppress the go ahead, and says no to the rest. The telnet
commands are taken out of what is received, the byte 255 is sent doubled and a
carriage return is sent as CR NUL. Many BBSes on a port other than 23 are not
telnet: dial them with `ATD` if `ATDT` shows strange characters.

The modem is always ready, the carrier is not on DCD as with `&C0`: the IIc
sends only when its pin 5 is ready, and it has to send the AT commands. It does
not answer calls, `ATA` is an error.

## What Is Not Emulated

- Breaks, and the parity bit on the line. There are no parity or framing errors.
- Overruns. The next byte received waits until the program has read the last
  one, nothing is lost.
- The choice of the CTS pin of the Super Serial Card, switches SW1-7 and SW2-7.
- The external interrupt of the IIc, on the disk drive connector.
- The baud rates of the first IIcs, up to 2% below the nominal ones.

## References

- "Apple IIc Technical Reference Manual": chapters 7 and 8 for the ports and
  their commands, chapter 11 for the 6551, appendix E for the interrupts.
- "Super Serial Card Installation and Operating Manual", with its firmware
  [disassembled](https://6502disassembly.com/a2-rom/SSC.html).
- [Super Serial Card switch settings](https://www.mrynet.com/FTP/os/AppleII/apple-super-serial-setup.txt).
- [Hayes command set](https://en.wikipedia.org/wiki/Hayes_command_set), and the
  [Zimodem](https://github.com/bozimmerman/Zimodem/blob/master/README) firmware
  for the dialing of the WiFi modems.
- [RFC 854](https://www.rfc-editor.org/rfc/rfc854), the telnet protocol.
