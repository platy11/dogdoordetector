# Dog Door Detector

## Description

Recently we had a dog door installed for our dog, and we wanted to find
whether the dog was actually using the door at night and when we weren't home.
The door uses a small magnet to clip shut, and I found that this magnet could
be detected using the magnetometer in an old Android phone that I had
available, so I decided to make a program that would notify us when the dog
door was opened.

The program is written in Go, and it runs on Termux (a Linux environment
for Android) using Termux:API.

Once the program is started, the phone is left taped against the door, just
below the dog door and its magnet. When the door is detected to be opened based
on the calibration values set in `main.go`, a message will be sent via Telegram
to the configured group.

The program uses Telegram to send the messages because it allows a record of
times the dog door was used to be kept. It also allows the users recieving the
notifications to be easily updated by just changing the group members.
Addditionally, sending notifications this way requires no constanly running
HTTP server on the phone since messages can be sent with just a GET request.  

## Installation and Configuration

This program runs inside Termux, so you will need to install Termux and
Termux:API. If your device's Android API version is too low for the Play Store
version (mine was), you can use F-Droid and enable the archive repository to
download an older version. In Termux, you will need to install the following
packages:

```bash
max@termux$ pkg install termux-api golang
```

(If you plan to cross-compile the program from your computer you can omit
`golang`, although in this case a C compiler and the Android SDK should be
installed on your computer)

You can then download the
[latest version](https://github.com/platy11/dogdoordetector/releases/latest)
of the code. You'll need to configure the constants at the start of main.go and the JSON
key name struct tag on line 102 to suit your requrements (check the output of
`$ termux-sensor -n 1 -s magnet` for the magnetic sensor name). After that, you
can transfer your source files to your device (cross-compiling is also an
option, but you may need to enable Cgo and install a C compiler & the Android
SDK). Any method to get the files across works, I used Python (version 3) on
the PC, usually installed by default on Linux and macOS and `wget` on Termux,
installable with `$ pkg install wget`:

```bash
max@pc$ python -m http.server

max@termux$ wget -r [IP of your PC]:8000 -P detector
max@termux$ cd detector
```

Then compile:

```bash
max@termux$ go build .
```

## Usage

Once you have compiled the program, you can run the compiled binary:

```bash
max@termux$ DETECTOR_TG_KEY=[Telegram bot API key] DETECTOR_TG_CHAT=[Telegram chat ID] ./detector
```

You can obtain a Telegram bot API key by messaging
[@botfather](https://t.me/botfather) on Telegram
([see here for details](https://core.telegram.org/bots/#creating-a-new-bot)).
To get the chat ID of a group, add [@platy11bot](https://t.me/platy11bot) to
the group and send the message `/id@platy11bot`. Copy the given chat ID, making
sure to include any negative signs.

However, I recommend that you create a small shell script to run the program
so the log output can be saved and you don't have to keep typing your secrets,
for example:

`./run`:
```bash
#!/bin/sh
# Saves stdout and stderr to a log file with the current time while
# also displaying it in the terminal
DETECTOR_TG_KEY=[Telegram bot API key] DETECTOR_TG_CHAT=[Telegram chat ID] ./detector 2>&1 | tee -a log/log-$(date +%s).txt
```

```bash
max@termux$ chmod +x run
max@termux$ ./run
```

## Known Issues

- For unknown reasons, the os.Interrupt signal does not seem to work correctly
when running in Termux. Instead of using Ctrl-C (interrupt) to end the script,
you should use Ctrl-D (end of file) which the program will understand as
signalling that it should gracefully exit.

## Copyright/License

Copyright (c) 2020 platy11.

This program is free software: you can redistribute it and/or modify it under
the terms of the GNU General Public License as published by the Free Software
Foundation, version 3.

This program is distributed in the hope that it will be useful, but WITHOUT ANY
WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with
this program. If not, see <https://www.gnu.org/licenses/>.