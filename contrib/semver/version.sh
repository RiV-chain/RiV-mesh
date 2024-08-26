#!/bin/sh

case "$*" in
  *--bare*)
    # Remove the "v" prefix
    git describe --abbrev=0 --tags --match="[0-9]*\.[0-9]*\.[0-9]*\.[0-9]*"
    ;;
  *)
    git describe --abbrev=0 --tags --match="v[0-9]*\.[0-9]*\.[0-9]*\.[0-9]*"
    ;;
esac