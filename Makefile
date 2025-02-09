install:
	go install ./cmd/goback

debian-install:
	/usr/local/go/bin/go build -o /usr/bin/goback ./cmd/goback

test:
	@ go install ./cmd/goback
	@ goback
