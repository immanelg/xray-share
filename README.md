Command to convert xray config <-> share url (like vless://*).

Based on 'official' libXray.

# Quickstart
```sh
./do install
```

```sh
export SUB=https://raw.githubusercontent.com/free-nodes/v2rayfree/main/sub
xray-share -T -i $SUB -o free-nodes.json -tmax 3000 -turl https://google.com 
```
The `-T` command: 
- tests configs from `-i` file (in this case, remote url)
- writes good configs (xray json) to `-o` file.

Now you can run xray:
```sh
cp free-nodes.json config/
cp base.json config/   # this is your own base config that includes inbounds, rules, dns, observatory, ...
xray -confdir config/
```

# Commands
```sh
-V
  Print version info
-U -i URL_SUBS [-o FILEPATH_JSON=generated-config.json]
  Convert share links to json config
-C -i URL_JSON [-o FILEPATH_SUBS=generated-links.txt]
  Convert json config to share links
-T [-tmax MS] [-tmin MS] [-turl URL=https://gstatic.com/generate_204] -i URL_JSON_OR_SUBS [-o FILEPATH_JSON=generated-config.json]
  Filter good configs
```

Examples:
```sh
# config->sub
xray-share -C -i config.json -o sub.txt

# sub->config
xray-shareurl -U -i https://example/sub.txt -o outbounds.json
# Note: -U only generates outbounds config. you should also probably have /etc/xray/base-config.json without outbounds.

# test configs and keep fast and working ones
xray-share -T -i config.json -o config.filtered.json -tmax 1000 -turl https://google.com
```


