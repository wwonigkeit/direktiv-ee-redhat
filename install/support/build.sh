#!/bin/sh

GO_BUILD_TAGS="osusergo,netgo" CGO_ENABLED=0 go build -o ../00_prepare/prepare main.go


GOOS=linux GOARCH=arm64 GO_BUILD_TAGS="osusergo,netgo" CGO_ENABLED=0 go build -o ../00_prepare/prepare_arm main.go


