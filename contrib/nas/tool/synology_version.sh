#!/bin/sh

# Get the last tag
TAG=$(git describe --abbrev=0 --tags --match="v[0-9]*\.[0-9]*\.[0-9]*" 2>/dev/null)

# Did getting the tag succeed?
if [ $? != 0 ] || [ -z "$TAG" ]; then
  printf -- "unknown"
  exit 0
fi

# Get the current branch
BRANCH=$(git symbolic-ref -q HEAD --short 2>/dev/null)

# Did getting the branch succeed?
if [ $? != 0 ] || [ -z "$BRANCH" ]; then
  BRANCH="master"
fi

#replace last dot with -
STAG=$(echo $TAG | sed 's/v//' | sed 's/[^0123456789.].//' | sed 's/\.\([^.]*\)$/-\1/')
#get tail after - and add 6000 for padding
TAG_TAIL=$(($(echo $STAG | sed -n -e 's/^.*-//p')+6000))
#replace tail after -
SYNO_VERSION=$(echo $STAG | sed "s/-.*/-$TAG_TAIL/")
case "$*" in
  *--bare*)
    printf '%s\n' "$SYNO_VERSION"
        ;;
  *)
    printf '%s' "$SYNO_VERSION"
    ;;
esac
