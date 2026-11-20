Command to convert xray config <-> share url (like vless://*).

Based on 'official' libXray.

# Quickstart
## Install
```sh
./do install
xray-share -V
```
See `./do help`.

## Example: fetch configs from a profile, filter them by speed, convert to JSON and run Xray-core
```sh
export SUB=https://raw.githubusercontent.com/free-nodes/v2rayfree/main/sub
xray-share -T -i $SUB -o free-nodes.json
```
The `-T` command: 
- tests configs from `-i` file (in this case, remote url with share links)
- writes good configs to `-o` file (xray outbounds json)

Now you can run xray:
```sh
cp free-nodes.json config/
cp base.json config/   # this is your own base config that includes inbounds, rules, dns, observatory, ...
xray -confdir config/
```

# Commands
```sh
 ~/src/xray-share$ ./xray-share -h 
-V 
  Print version info 
-U [-i URL_SUBS=-] [-o FILEPATH_JSON=-] [-uniqtags] 
  Convert share links to json config 
-C [-i URL_JSON=-] [-o FILEPATH_SUBS=-] 
  Convert json config to share links 
-T [-tmax MS] [-tmin MS] [-turl URL=https://gstatic.com/generate_204] [-i URL_JSON_OR_SUBS=-] [-o FILEPATH_JSON=-] [-uniqtags] 
  Filter good configs 
```

Examples:
```sh
# config->sub
xray-share -C -i config.json -o sub.txt

# sub->config
all_proxy=socks5h://localhost:2080 xray-shareurl -U -i https://example/sub.txt -o outbounds.json
# Note: -U only generates outbounds config. you should also probably have /etc/xray/base-config.json without outbounds.

# test configs and keep fast and working ones
xray-share -T -i config.json -o config.filtered.json -tmax 1000 -turl https://google.com
```


