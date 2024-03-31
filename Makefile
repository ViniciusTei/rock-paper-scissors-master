run:
	go run *.go

dev:
	air

build:
	GOOS=linux GOARCH=amd64 go build -o dist/server

send-exe-to-remote-server:
	rsync dist/server root@$(REMOTE_SERVER_IP):~

send-service-file-to-remote-server:
	rsync http-server.service root@$(REMOTE_SERVER_IP):~

deploy: build send-exe-to-remote-server send-service-file-to-remote-server
	ssh -t root@$(REMOTE_SERVER_IP) '\
		sudo mv ~/http-server.service /etc/systemd/system/ \
		&& sudo systemctl enable http-server \
		&& sudo systemctl restart http-server \
	'
