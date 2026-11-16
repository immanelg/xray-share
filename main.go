package main

import (
	"fmt"
	"os"
	// "crypto/rand"
	"encoding/json"
	"github.com/xtls/libxray/share"
	"flag"
)

func assertSuccess(message string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v", message, err)
		os.Exit(1)
	}
}

func main() {
	var configToUrl bool
	var urlToConfig bool
	var inFile string
	var outFile string
	flag.BoolVar(&configToUrl, "C", false, "config to url")
	flag.BoolVar(&urlToConfig, "U", false, "urls to config")
	flag.StringVar(&inFile, "i", "", "in file")
	flag.StringVar(&outFile, "o", "", "out file")
	flag.Parse()

	switch {
	case urlToConfig:
		links, errReadIn := os.ReadFile(inFile)
		assertSuccess("cannot read input file", errReadIn)

		cfgjson, errconv := share.ConvertShareLinksToXrayJson(string(links), "")
		assertSuccess("cannot convert", errconv)

		cfg, errjson := json.MarshalIndent(cfgjson, "", "  ")
		assertSuccess("cannot marshal", errjson)

		if outFile == "" { outFile = "generated-config.json" }
		// fh, errcreate := os.Create("outbounds-generated-"+rand.Text()+".jsonc")
		fh, errcreate := os.Create(outFile)
		assertSuccess("cannot create", errcreate)

		_, errwrite := fh.Write(cfg)
		assertSuccess("cannot write", errwrite)
	case configToUrl:
		cfg, errReadIn := os.ReadFile(inFile)
		assertSuccess("cannot read input file", errReadIn)

		links, errconv := share.ConvertXrayJsonToShareLinks(cfg)
		assertSuccess("cannot convert", errconv)

		if outFile == "" { outFile = "generated-links.txt" }
		fh, errcreate := os.Create(outFile)
		assertSuccess("cannot create", errcreate)

		_, errwrite := fh.Write([]byte(links))
		assertSuccess("cannot write", errwrite)

	default:
		fmt.Printf("one of commands must be specified\n")
		flag.Usage()
	}
}
