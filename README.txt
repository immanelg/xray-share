command to convert xray config <-> share url (like vless://*)

based on 'official' libXray

commands
    # config->sub
    xray-shareurl -C -i config.json -o sub.txt
    # sub->config
    xray-shareurl -U -i sub.txt -o outbounds.json

    # -i also supports urls
    # xray-shareurl -C -i https://raw.githubusercontent.... -o sub.txt

    # test configs
    xray-shareurl -T -tmax 1000 -i config.json -o config.json

note: -U only generates outbounds config. you should also probably have /etc/xray/base-config.json without outbounds.

install
    go build -o ~/.local/bin/xray-shareurl
