install:
	go build -tags luajit --ldflags '-extldflags "-lm -ldl -static"'
	mv heart ~/.bin
