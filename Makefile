all:
	GOARCH=386 GOOS=windows go build -ldflags "-w -s" -o build/expandAutoTileToTiled-win-x86.exe
	GOARCH=amd64 GOOS=windows go build -ldflags "-w -s" -o build/expandAutoTileToTiled-win-x64.exe
	GOARCH=386 GOOS=linux go build -ldflags "-w -s" -o build/expandAutoTileToTiled-linux-x86
	GOARCH=amd64 GOOS=linux go build -ldflags "-w -s" -o build/expandAutoTileToTiled-linux-x64
	GOARCH=amd64 GOOS=darwin go build -ldflags "-w -s" -o build/expandAutoTileToTiled-drawin-x64

clean:
	rm build -rf