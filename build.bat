set GOOS=js
set GOARCH=wasm
go build -ldflags "-s -w" .\cli\wasm
move wasm assets/neptune.wasm
