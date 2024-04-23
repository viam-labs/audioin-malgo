
all: build bundle

build:
	CGO_ENABLED=1 \
							GOOS=linux \
							GOARCH=arm64 \
							CC="zig cc -target aarch64-linux-gnu" \
							CXX="zig c++ -target aarch64-linux-gnu" \
							go build -o bin/audioin-malgo-linux-aarch64 --ldflags '-s -w -linkmode external' ./main.go
	CGO_ENABLED=1 \
							GOOS=linux \
							GOARCH=amd64 \
							CC="zig cc -target x86_64-linux-gnu" \
							CXX="zig c++ -target x86_64-linux-gnu" \
							go build -o bin/audioin-malgo-linux-amd64 --ldflags '-s -w -linkmode external' ./main.go
	CGO_ENABLED=1 \
							GOOS=darwin \
							GOARCH=amd64 \
							go build -o bin/audioin-malgo-macos-amd64 --ldflags '-s -w' ./main.go
	GOOS=darwin \
							GOARCH=arm64 \
							go build -o bin/audioin-malgo-macos-arm64 --ldflags '-s -w' ./main.go

bundle:
	mkdir -p dist/
	tar -czf dist/linux-aarch64-module.tar.gz bin/audioin-malgo-linux-aarch64
	tar -czf dist/linux-amd64-module.tar.gz bin/audioin-malgo-linux-amd64
	tar -czf dist/macos-amd64-module.tar.gz bin/audioin-malgo-macos-amd64
	tar -czf dist/macos-arm64-module.tar.gz bin/audioin-malgo-macos-arm64
