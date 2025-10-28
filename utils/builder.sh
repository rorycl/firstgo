#!/bin/bash

# build many go binary output targets
# https://freshman.tech/snippets/go/cross-compile-go-programs/

set -e

if [ "$#" -ne 2 ]; then
    echo "Required arguments not found."
    echo "please provide the output directory path"
    echo "followed by the basename of the binary name to compile"
    echo "as arguments."
    echo "eg $0 bin firstgo"
	exit 1
fi

OUTDIR=$1
BINBASENAME=$2
# THISDIR=$(dirname "$0")
THISDIR=$(pwd)
FULLDIRPATH=${THISDIR}/${OUTDIR}

# clean bin
if [ ! -d "${FULLDIRPATH}" ]; then
    echo "output directory ${FULLDIRPATH} not found"
    exit 1
fi
rm -f $OUTDIR/$BINBASENAME*

LINUX='linux:0:amd64:linux-amd64'
WIN='windows:0:amd64:win-amd64.exe'
MACAMD='darwin:0:amd64:darwin-amd64'
MACARM='darwin:0:arm64:darwin-arm64'

echo "-----------------"
echo "building binaries"
echo "-----------------"
for II in $LINUX $WIN $MACAMD $MACARM; do
	os=$(echo $II | cut -d":" -f1)
	cgo=$(echo $II | cut -d":" -f2)
	arch=$(echo $II | cut -d":" -f3)
	suffix=$(echo $II | cut -d":" -f4)
	# echo $os $arch $suffix;
	echo GOOS=${os} GOARCH=${arch} CGO_ENABLED=${cgo} go build -o ${FULLDIRPATH}/${BINBASENAME}-${suffix} .
	GOOS=${os} GOARCH=${arch} CGO_ENABLED=${cgo} go build -o ${FULLDIRPATH}/${BINBASENAME}-${suffix} .
done

echo ""
echo "------------------------------------"
echo "now release using gh, the github cli"
echo "------------------------------------"
echo "examples:"
echo "gh release --help"
echo "gh release list"
echo "gh release view v0.0.5"
echo "gh release delete v0.0.5"
echo "gh release create v0.0.5 --generate-notes ${OUTDIR}/${BINBASENAME}-*"
echo "gh release view v0.0.5"
echo "------------------------------------"

