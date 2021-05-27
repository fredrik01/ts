# timestamp-stopwatch

There is also a version written in shell script: https://github.com/fredrik01/ts-shell.git

Save timestamps to lists (stopwatches). Show time between each timestamp and time between each timestamp and the first one.

Made to keep track of:
- How long it takes before my baby falls asleep during walks
- How long he sleeps
- How long we are out and about

## Usage

Add timestamp to default or named stopwatch

	ts save
	ts save baby

Show timestamps and time between them

	ts show
	ts show baby

	# Example output:

	Timestamp              Since prev   Since first
	2021-05-26 20:04:26
	2021-05-26 20:04:51           25s           25s
	2021-05-26 20:06:07         1m16s         1m41s
	2021-05-26 20:06:58           51s         2m32s
	2021-05-26 20:08:23         1m25s         3m57s
	Now                        10m23s        14m20s

Reset

	ts reset
	ts reset baby

List stopwatches (TODO)

	ts list

Cat log file (TODO)

	ts raw
	ts raw baby

## Installation

Download binary from the release page.

	curl -L https://github.com/fredrik01/ts-go/releases/latest/download/ts_linux_amd64.tar.gz -o ts.tar.gz
	mkdir ~/.ts
	tar -xvf ts.tar.gz -C ~/.ts
	rm ts.tar.gz

Add this to `.zshrc`

	[ -f ~/.ts/ts.zsh ] && source ~/.ts/ts.zsh

For Termux skip the thing above and do this instead

	ln -s $HOME/.ts/ts $PREFIX/bin/ts

## Update

	curl -L https://github.com/fredrik01/ts-go/releases/latest/download/ts_linux_amd64.tar.gz -o ts.tar.gz
	tar -xvf ts.tar.gz -C ~/.ts
	rm ts.tar.gz
