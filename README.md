# Timestamp Stopwatch

There is also a version written in shell script: https://github.com/fredrik01/ts-shell.git

Save timestamps to lists (stopwatches). Show time between each timestamp and time between each timestamp and the first one.

Made to keep track of:
- How long it takes before my baby falls asleep during walks
- How long he sleeps
- How long we are out and about

## Usage

Add timestamp to default or named stopwatch

	ts add
	ts add sleep

Show timestamps and time between them

	ts show
	ts show sleep

	# Example output:

	Timestamp              Since prev   Since first
	2021-05-26 20:04:26
	2021-05-26 20:04:51           25s           25s
	2021-05-26 20:06:07         1m16s         1m41s
	2021-05-26 20:06:58           51s         2m32s
	2021-05-26 20:08:23         1m25s         3m57s
	Now                        10m23s        14m20s

Show timestamps from all or some stopwatches in the same list

	ts show -combine

Print all lists

	ts show -all

Reset

	ts reset
	ts reset sleep
	ts reset -all

Run `ts` for all commands and options

## Installation

### MacOS

	curl -L https://github.com/fredrik01/ts/releases/latest/download/ts_darwin_amd64.tar.gz -o ts.tar.gz
	mkdir /tmp/ts
	tar -xvf ts.tar.gz -C /tmp/ts
	mv /tmp/ts/ts /usr/local/bin/ts
	rm ts.tar.gz
	rm -r /tmp/ts

### Android / Termux

	curl -L https://github.com/fredrik01/ts/releases/latest/download/ts_android_arm64.tar.gz -o ts.tar.gz
	mkdir ts
	tar -xvf ts.tar.gz -C ts
	mv ts/ts $PREFIX/bin/ts 
	rm ts.tar.gz
	rm -r ts
