export GOOS=linux
go build -o build/smoltext src/main.go
export GOOS=windows
go build -o build/smoltext.exe src/main.go