#!/bin/sh
# Build torex. The app comes with the runtime that the tree has; the
# installer it makes is "Torex Setup <version>-<revision>.exe", the revision
# being the short hash of the last commit when the tree is a Git repository.
#
#	sh build.sh          the app, from the runtime in the tree
#	sh build.sh all      the runtime's programs too, then the app
#	sh build.sh shell    WinKit's shell from source, then the app
#	sh build.sh vim      vim from source, then the app
#	sh build.sh ssh      the ssh client, then the app
#
# Building the shell or vim needs zig, make, windres and Git: tools/mkbusybox
# and tools/mkvim find them in $WINKIT_HOME or in PATH.
set -e
# Run from the tree, whatever the shell was in.
case "$0" in
*/*) cd "${0%/*}" ;;
esac

app() { go run ./tools/mkapp; }

case "${1:-app}" in
app)
	app
	;;
all)
	go run ./tools/mkbusybox
	go run ./tools/mkvim
	go run ./tools/mkssh
	app
	;;
shell)
	go run ./tools/mkbusybox
	app
	;;
vim)
	go run ./tools/mkvim
	app
	;;
ssh)
	go run ./tools/mkssh
	app
	;;
*)
	echo "usage: sh build.sh [all|app|shell|vim|ssh]" >&2
	exit 2
	;;
esac
