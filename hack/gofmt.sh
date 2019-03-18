#!/bin/bash
files=$(find . -type f -name '*.go' | grep -v vendor | grep -v .git)
out=$(gofmt -d -s $files)
if [ "$out" != "" ]; then
	echo "$out"
	echo
	echo "You might want to run something like 'find . -name '*.go' | xargs gofmt -w -s'"
	exit 2
fi
exit 0

