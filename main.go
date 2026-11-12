package main

import (
	"fmt"
	"os"
	"strings"
	// "crypto/rand"
	"encoding/json"
	"github.com/xtls/libxray/share"
)

func main() {
	links := strings.Join(os.Args[1:], "\n")
	cfgjson, err := share.ConvertShareLinksToXrayJson(links, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot convert: %v", err)
		return
	}
	cfg, errjson := json.MarshalIndent(cfgjson, "", "  ")
	if errjson != nil {
		fmt.Fprintf(os.Stderr, "cannot marshal: %v", errjson)
		return
	}
	// fh, errcreate := os.Create("outbounds-generated-"+rand.Text()+".jsonc")
	fh, errcreate := os.Create("outbounds-from-url.jsonc")
	if errcreate != nil {
		fmt.Fprintf(os.Stderr, "cannot create: %v", errcreate)
		return
	}
	_, errwrite := fh.Write(cfg)
	if errwrite != nil {
		fmt.Fprintf(os.Stderr, "cannot write: %v", errwrite)
		return
	}
}
