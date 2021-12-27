all: justgrep irc2json

install: all
	install -Dm 755 justgrep "${DESTDIR}/usr/bin/justgrep"
	install -Dm 755 irc2json "${DESTDIR}/usr/bin/irc2json"

justgrep: cmd/justgrep/justgrep.go
	go build cmd/justgrep/justgrep.go

irc2json: cmd/irc2json/irc2json.go
	go build cmd/irc2json/irc2json.go
