libXray-based command to convert xray share url (like vless://) to xray json

example

     ~/src/xray-shareurl-tool$ xray-shareurl 'vless://97e0ea78-67ae-4984-8a14-a7d1a51e58de@199.232.78.160:443?pcs=cd6e838b9bfe31cab8b3d5c858b0d6223fe84ece236c1fdb281be93c302e1e6e&security=tls&encryption=none&insecure=0&host=copy_az_channel_NUFiLTER_10i.global.ssl.fastly.net.&type=ws&allowInsecure=0&sni=cdn.akamai.steamstatic.com#bs'
    2026/11/12 22:07:04.672344 [Warning] common/errors: The feature WebSocket transport (with ALPN http/1.1, etc.) is deprecated, not recommended for using and might be removed. Please migrate to XHTTP H2 & H3 as soon as possible.
    ....

     ~/src/xray-shareurl-tool$ cat outbounds-from-url.jsonc
    {"outbounds":[{"protocol":"vless","settings":{"address":"199.232.78.160","encryption":"none","id":"97e0ea78-67ae-4984-8a14-a7d1a51e58de","port":443},"streamSettings":{"network":"ws","security":"tls","tlsSettings":{"pinnedPeerCertSha256":"cd6e838b9bfe31cab8b3d5c858b0d6223fe84ece236c1fdb281be93c302e1e6e","serverName":"cdn.akamai.steamstatic.com"},"wsSettings":{"host":"copy_az_channel_NUFiLTER_10i.global.ssl.fastly.net."}},"tag":"bs"}]}

install
    go build -o ~/.local/bin/xray-shareurl
