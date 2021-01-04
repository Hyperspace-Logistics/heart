.PHONY: build
build:
	go build -tags luajit --ldflags '-extldflags "-lm -ldl -static"'

install: build
	mv heart ~/.bin
