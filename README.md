Command to convert xray config <-> share url (like vless://*).

Based on 'official' libXray.

Usage:
```sh
Commands: 
xray-share -V 
        Print version. 
xray-share -U [-i URL_SUBS] [-o FILEPATH_JSON=generated-config.json] 
        Convert share links to json config. 
xray-share -C [-i URL_JSON] [-o FILEPATH_SUBS=generated-links.txt] 
        Convert json config to share links. 
xray-share -T [-tmax MS] [-tmin MS] [-turl URL] [-i URL_JSON] [-o FILEPATH_JSONFILEPATH_JSON=generated-config.json] 
        Filter good configs. 
```
Examples:
```sh
# config->sub
xray-share -C -i config.json -o sub.txt

# sub->config
xray-shareurl -U -i https://example/sub.txt -o outbounds.json

# test configs and keep fast and working ones
xray-share -T -i config.json -o config.filtered.json -tmax 1000 -turl https://google.com
```


Note: -U only generates outbounds config. you should also probably have /etc/xray/base-config.json without outbounds.

Install
    go build -o ~/.local/bin/xray-share


