all: justgrep irc2json

install: all
	echo "Compressing man pages..."
	gzip -k man1/justgrep.1 -f
	gzip -k man1/irc2json.1 -f
	echo "Installing binaries..."
	install -Dm 755 justgrep "${DESTDIR}/usr/bin/justgrep"
	install -Dm 755 irc2json "${DESTDIR}/usr/bin/irc2json"
	echo "Installing man pages..."
	install -Dm 644 man1/justgrep.1.gz "${DESTDIR}/usr/share/man/man1/justgrep.1.gz"
	install -Dm 644 man1/irc2json.1.gz "${DESTDIR}/usr/share/man/man1/irc2json.1.gz"

justgrep: cmd/justgrep/justgrep.go
	go build cmd/justgrep/justgrep.go

irc2json: cmd/irc2json/irc2json.go
	go build cmd/irc2json/irc2json.go
