
ROOT_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))
EXE_NAME := aospparse
BIN := bin

ifeq ($(OS),Windows_NT)
	RM_F_CMD = Remove-Item -erroraction silentlycontinue -Force
    RM_RF_CMD = ${RM_F_CMD} -Recurse
	exe =${BIN}/${EXE_NAME}.exe
else
	RM_F_CMD = rm -f
	RM_RF_CMD = ${RM_F_CMD} -r
	exe =${BIN}/${EXE_NAME}
endif

build:
	echo "Compiling for every OS and Platform"
	set GOOS=linux
	set GOARCH=arm64
	go build -o ${BIN}/${EXE_NAME} main.go
	set GOOS=windows
	set GOARCH=arm64
	go build -o ${BIN}/${EXE_NAME}.exe main.go

test:
	echo "===========Testing==============="
	${exe} sample/Android.bp srcs

clean:
	$(RM_RF_CMD) ${BIN}/${EXE_NAME}

all: build test
.DEFAULT_GOAL := all