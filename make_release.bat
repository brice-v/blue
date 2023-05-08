go clean -testcache && go test ./... && go build -ldflags="-s -w -extldflags='-static'" && strip blue.exe && upx --best blue.exe && cp blue.exe C:/Users/brice/OneDrive/Documents/.blue/bin/
