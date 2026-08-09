# The headless frontend

`headless` has no window and no screen. It builds the machine, leaves it
paused and gives you a prompt to drive it: run so many cycles, type something,
print the text screen, save a snapshot. It is the frontend for a script, for a
test, or for looking at what a machine does without watching it happen.

## Building

Nothing to install besides Go:

``` terminal
git clone github.com/ivanizag/izapple2
cd izapple2/frontend/headless
go build .
```

## Running

``` terminal
casa@servidor:~$ ./headless
* run 40000
* text
* type 10 PRINT "HELLO WORLD"
* enter
* run 100
* text
* quit
```

The machine is built with the [command line options](command_line.md) like
everywhere else, and starts paused. Nothing happens until you say `start` or
`run`.

## Commands

| Command | What it does |
| --- | --- |
| `help` | Prints the list of commands |
| `quit` | Quits |
| `start` | Runs the machine at its normal speed, and returns to the prompt while it runs |
| `pause` | Stops it |
| `run <cycles>` | Runs `<cycles>` thousand cycles at full speed and waits until they are done. Only while paused |
| `cycle` | Prints the cycle count |
| `reset` | Resets the machine |
| `key <number>` | Queues a key, as a decimal number from 0 to 127 |
| `type <text>` | Queues the characters of `<text>`. No quotes, it can have spaces |
| `enter` | Queues the Return key. The same as `key 13` |
| `clearkeys` | Empties the key queue |
| `text` | Prints the text screen on the terminal |
| `png` | Saves the screen as `snapshot.png` in NTSC colour |
| `pngm` | The same in monochrome |
| `gif` | Records one second of the screen to `snapshot.gif`, 20 frames of 5 hundredths |

The keys are queued, not typed: they are handed to the machine as it asks for
them, so a `type` before a `run` is a fine way to feed a program that is not
running yet.

`run` is the useful one for a script. It asks for full speed, sets a breakpoint
that many cycles ahead and waits, so a DOS 3.3 boot is over in a moment and the
prompt comes back exactly when it is done. A boot from a diskette spends some
37 million cycles, mostly waiting for the drive, so `run 40000` covers it.

The `help` the program prints is ahead of what it does in a few places: there
is no `stop` (it is `pause`) and no `gifm`, and `png`, `pngm` and `gif` take no
arguments, they always write `snapshot.png` and `snapshot.gif`.

## What is missing

There is nothing to insert a diskette while it runs, and nothing to move the
joysticks or press their buttons. Both have to come from the command line, or
not at all.
