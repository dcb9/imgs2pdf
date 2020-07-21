.PHONY: buildBinary
buildBinary:
	packr2

.PHONY: cleanBinary
cleanBinary:
	packr2 clean

.PHONY: win
win: buildBinary buildWin cleanBinary

.PHONY: buildWin
buildWin:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o imgs2pdf.exe .
