run:
	go run *.go

dev:
	air

build:
	GOOS=linux GOARCH=amd64 go build -o dist/server

send-exe-to-remote-server:
	rsync dist/server $(SSH_USER)@$(SSH_HOST):~

send-service-file-to-remote-server:
	rsync http-server.service $(SSH_USER)@$(SSH_HOST):~

deploy: build send-exe-to-remote-server send-service-file-to-remote-server
	ssh -t $(SSH_USER)@$(SSH_HOST) '\
		sudo mv ~/http-server.service /etc/systemd/system/ \
		&& sudo systemctl enable http-server \
		&& sudo systemctl restart http-server \
	'
