#!/usr/bin/env bash

# remove nested vendor dirs

SRCPATH="${GOPATH%%:*}/src/"
FIND_NESTED_VENDOR="find vendor -path '${SRCPATH}' -type d"
echo "found : ${FIND_NESTED_VENDOR}"
#${FIND_NESTED_VENDOR} | xargs --no-run-if-empty cp -rln -t .
#${FIND_NESTED_VENDOR} | xargs --no-run-if-empty rm -r


