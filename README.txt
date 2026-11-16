command to convert xray config <-> share url (like vless://*)

based on 'official' libXray

commands
    xray-shareurl -C -i config.json -o sub.txt
    xray-shareurl -U -i sub.txt -o outbounds.json

note: -U only generates outbounds config. you should also probably have /etc/xray/base-config.json without outbounds.

install
    go build -o ~/.local/bin/xray-shareurl
