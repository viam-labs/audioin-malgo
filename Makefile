build:
	CGO_ENABLED=1 \
							GOOS=linux \
							GOARCH=arm \
							GOARM=7 \
							CC="zig cc -target arm-linux-musleabihf" \
							CXX="zig c++ -target arm-linux-musleabihf" \
							go build -o bin/audioin-malgo-linux-arm7 --ldflags '-s -w -linkmode external -extldflags "--static"' ./main.go
	CGO_ENABLED=1 \
							GOOS=linux \
							GOARCH=arm64 \
							CC="zig cc -target aarch64-linux-musl" \
							CXX="zig c++ -target aarch64-linux-musl" \
							go build -o bin/audioin-malgo-linux-aarch64 --ldflags '-s -w -linkmode external -extldflags "--static"' ./main.go
	CGO_ENABLED=1 \
							GOOS=linux \
							GOARCH=amd64 \
							CC="zig cc -target x86_64-linux-musl" \
							CXX="zig c++ -target x86_64-linux-musl" \
							go build -o bin/audioin-malgo-linux-amd64 --ldflags '-s -w -linkmode external -extldflags "--static"' ./main.go
	CGO_ENABLED=1 \
							GOOS=darwin \
							GOARCH=amd64 \
							go build -o bin/audioin-malgo-macos-amd64 --ldflags '-s -w' ./main.go
	CGO_ENABLED=1 \
							GOOS=darwin \
							GOARCH=arm64 \
							go build -o bin/audioin-malgo-macos-arm64 --ldflags '-s -w' ./main.go

