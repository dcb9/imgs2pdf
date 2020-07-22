.PHONY: buildBinary
buildBinary:
	packr2

.PHONY: cleanBinary
cleanBinary:
	packr2 clean

.PHONY: buildWin
buildWin:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o imgs2pdf.exe .

.PHONY: compressWin
compressWin:
	tar cfJ imgs2pdf.tar.xz imgs2pdf.exe

.PHONY: win
win: buildBinary buildWin cleanBinary compressWin

