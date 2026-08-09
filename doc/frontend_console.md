# The console frontend

`console` runs the emulator on the terminal. There is no window and no SDL2
dependency: the text screen of the Apple II is drawn in place with ANSI escape
codes, and what you type is sent to the machine a line at a time.

It is the frontend for a machine without a graphical session, or for a quick
run over ssh. Only the 40 column text screen exists here: there are no
graphics, no sound and no joysticks.

## Building

Nothing to install besides Go:

``` terminal
git clone github.com/ivanizag/izapple2
cd izapple2/frontend/console
go build .
```

## Running

The trick of redrawing the screen in place does not work with the Apple //e
ROM, so use one of the Apple ][ models:

``` terminal
casa@servidor:~$ ./console -model 2plus

############################################
#                                          #
#                APPLE II                  #
#                                          #
#     DOS VERSION 3.3  SYSTEM MASTER       #
#                                          #
#                                          #
#            JANUARY 1, 1983               #
#                                          #
#                                          #
# COPYRIGHT APPLE COMPUTER,INC. 1980,1982  #
#                                          #
#                                          #
# ]10 PRINT "HELLO WORLD"                  #
#                                          #
# ]LIST                                    #
#                                          #
# 10  PRINT "HELLO WORLD"                  #
#                                          #
# ]RUN                                     #
# HELLO WORLD                              #
#                                          #
# ]_                                       #
#                                          #
#                                          #
############################################
Line:

```

The rest of the [command line options](command_line.md) work as everywhere
else, so the diskettes, the cards and the tracing are the same.

Ctrl-C quits.

## Typing

The `Line:` at the bottom is the terminal reading a line. Nothing reaches the
Apple II until you press Enter, and then the whole line goes in at once. That
is fine for Applesoft and for the monitor, and it is no good for anything that
reads the keyboard while it runs, like a game or a full screen editor.

The screen is redrawn ten times a second, and only when something changed.
