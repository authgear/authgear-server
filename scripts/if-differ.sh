#!/bin/sh

if [ "$#" -lt 3 ]; then
    echo "Usage: if-differ.sh <new_file> <old_file> <command>"
    exit;
fi

FILE1=$1
FILE2=$2
shift
shift
CMD=$@

if [ ! -f "$FILE1" ]; then
    echo "$FILE1 is not a regular file."
    exit;
fi

if [ ! -f $FILE2 ] || [ "$(diff -q $FILE1 $FILE2)" != "" ] ; then
    echo "$FILE1 and $FILE2 differ."
    $CMD
else
    echo "$FILE1 and $FILE2 is the same."
fi
cp "$FILE1" "$FILE2"


