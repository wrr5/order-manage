# order-management

$env:GOOS="linux"
$env:GOARCH="amd64" 
$env:CGO_ENABLED="0"
go build -buildvcs=false -ldflags="-s -w" -o ordermanage .