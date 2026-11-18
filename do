#!/usr/bin/bash
set -eo pipefail

BIN=xray-share

DEFAULT=build
@build() {
    # TODO: x build

    CGO_ENABLED=0 \
        go build \
        -trimpath \
        -ldflags "-s -w -X main.version=$(git describe --tags --always) -X main.commitHash=$(git rev-parse --short HEAD)"  \
        -o "$BIN"
}

@watch() {
    watchexec -e go ./do build
}

@github_release() {
    todo
}

PREFIX=$HOME/.local
BINDIR=$PREFIX/bin

@install() {
    @build && cp -f "$BIN" "$BINDIR"
}

@uninstall() {
    rm -f "$BINDIR/$BIN"
}

@run() {
    @build
    ./$BIN "$@"
}

@clean() {
    go clean
}

todo() {
    echo "TODO"
    exit 1
}

##################################################################
@help() {
    echo "do™️: Control this project. This file is a simple, self-contained bash script."
    echo 
    echo "Available commands:"
    declare -F | grep "^declare -f @" | cut -f 2 -d @ | sed "s|^|   -> $0 |"
}
if [[ -z $1 ]]; then
    # this has alphabetical sorting, so this doesnt work.
        # find first task:
        # task="$(declare -F | grep "^declare -f @" | head -n1 | cut -f 2 -d @)"
    echo "$0 $DEFAULT"
    eval "@$DEFAULT"
else
    # check task exists
    declare -F | grep -qx "declare -f @$1" || { echo "No such task: $1"; echo; @help; exit 1; }
    task=@$1; shift
    eval "$task "$@""
fi
