#! /bin/bash

if [ -n "$DEV" ]; then
        mkdir -p /etc/letsencrypt/live/$DOMAIN
        openssl genrsa -out /etc/letsencrypt/live/$DOMAIN/privkey.pem 2048 && openssl req -new -x509 -sha256 -key /etc/letsencrypt/live/$DOMAIN/privkey.pem -out /etc/letsencrypt/live/$DOMAIN/fullchain.pem -days 3650 -subj '/CN=$DOMAIN:$PORT/O=CollabTest/C=CA'
        RED='\033[0;31m'
        NC='\033[0m'
        go run server.go &
        echo -e "\n${RED}started${NC}\n"
        inotifywait -rm -e close_write,moved_to,create . |
        while read -r directory events filename; do
                if [[ $filename == *.go ]]
                then
                        pkill server
                        go run server.go &
                        echo -e "\n${RED}restarted${NC}\n"
                fi
        done
else
        # TODO write a script that generates certs with letsencrypt and then starts the server
        go-wrapper run
fi