FROM	golang:latest AS builder
WORKDIR	/app
COPY	go.mod go.sum ./
RUN	go mod download
COPY	main.go	./
RUN	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/app .

FROM	scratch
COPY	--from=builder /etc/ssl/ /etc/ssl/
COPY	--from=builder /go/bin/app /go/bin/app
ENTRYPOINT	["/go/bin/app"]
