deps:
	go get github.com/jasonrogena/gonx
	go get github.com/mattn/go-sqlite3
	go install github.com/mattn/go-sqlite3
	go get github.com/BurntSushi/toml
	go get gopkg.in/cheggaaa/pb.v1
	go get github.com/satori/go.uuid
macos:
	env GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -o log-analyse-darwin-amd64

