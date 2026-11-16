#!/usr/bin/bash
set -eo pipefail

BIN=xray-sharetool

DEFAULT=build
@build() {
    # TODO: x build
    # TODO: x build
    go build -o $BIN
}

@watch() {
    watchexec -e go ./do build
}

@github_release() {
    todo
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
    echo "do™️: Control this project."
    echo 
    echo "Available commands:"
    declare -F | grep "^declare -f @" | cut -f 2 -d @ | sed "s|^|\t$0 |"
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
