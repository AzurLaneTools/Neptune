set GOOS=js
set GOARCH=wasm
go build -ldflags "-s -w" .\cmd\wasm
move wasm assets/neptune.wasm
