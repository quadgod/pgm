env GOOS=linux GOARCH=amd64 go build -o ./build/pgm/pgm-linux-amd64 ./cmd/pgm/main.go
env GOOS=windows GOARCH=amd64 go build -o ./build/pgm/pgm-win-amd64.exe ./cmd/pgm/main.go
env GOOS=darwin GOARCH=arm64 go build -o ./build/pgm/pgm-macos-arm64 ./cmd/pgm/main.go
