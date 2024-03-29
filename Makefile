build:
	CGO_ENABLED=1 GOARCH=arm64 GOOS=linux CC="zig cc -target aarch64-linux-gnu" CXX="zig c++ -target aarch64-linux-gnu" go build -o bin/linux-arm64-exec -ldflags "-s -w"

bundle:
	tar -czf module.tar.gz bin/linux-arm64-exec
