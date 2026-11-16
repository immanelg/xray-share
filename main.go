package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"runtime"

	// "crypto/rand"
	"encoding/json"
	"flag"

	"github.com/xtls/libxray/share"
	// "github.com/xtls/libxray/xray"
)

const version = "<devel>"
const commitHash = "<unknown commit>"

func assertSuccess(message string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v", message, err)
		os.Exit(1)
	}
}

func mustReadEntireFile(name string) []byte {
	path, errParseUrl := url.Parse(name)
	assertSuccess("cannot parse file url", errParseUrl)
	
	if path.Scheme == "" {
		content, errReadIn := os.ReadFile(name)
		assertSuccess("cannot read input file", errReadIn)
		return content
	}
	hclient := http.DefaultClient
	response, errReadIn := hclient.Get(name)
	assertSuccess("cannot read remote file", errReadIn)
	body, _ := io.ReadAll(response.Body)
	return body
}

func mustWriteEntireFile(name string, content []byte) {
	fh, errCreat := os.Create(name)
	assertSuccess("cannot create", errCreat)

	_, errWrite := fh.Write(content)
	assertSuccess("cannot write", errWrite)
}

func main() {
	// commands
	var printVersion bool
	var configToUrl bool
	var urlToConfig bool
	var testConfigs bool
	flag.BoolVar(&printVersion, "V", false, "version")
	flag.BoolVar(&configToUrl, "C", false, "config to url")
	flag.BoolVar(&urlToConfig, "U", false, "urls to config")
	flag.BoolVar(&testConfigs, "T", false, "filter out bad configs")

	// parameters
	var inFile string
	var outFile string
	var testUrl string
	var testFiltersBelowTimeout int
	var testFiltersAboveTimeout int
	flag.StringVar(&inFile, "i", "", "in file")
	flag.StringVar(&outFile, "o", "", "out file")
	flag.StringVar(&testUrl, "turl", "https://gstatic.com/generate_204", "test max timeout")
	flag.IntVar(&testFiltersAboveTimeout, "tmax", 3000, "test max timeout")
	flag.IntVar(&testFiltersBelowTimeout, "tmix", 100, "test min timeout")

	flag.Parse()

	switch {
	case printVersion:
		fmt.Println("xray-sharetool", version, commitHash, runtime.Version(), runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	case urlToConfig:
		links := mustReadEntireFile(inFile)

		cfgjson, errconv := share.ConvertShareLinksToXrayJson(string(links), "")
		assertSuccess("cannot convert", errconv)

		cfg, errjson := json.MarshalIndent(cfgjson, "", "  ")
		assertSuccess("cannot marshal", errjson)

		if outFile == "" { outFile = "generated-config.json" }
		mustWriteEntireFile(outFile, cfg)
	case configToUrl:
		cfg := mustReadEntireFile(inFile)

		links, errconv := share.ConvertXrayJsonToShareLinks(cfg)
		assertSuccess("cannot convert", errconv)

		if outFile == "" { outFile = "generated-links.txt" }
		mustWriteEntireFile(outFile, []byte(links))

	case testConfigs:
		fmt.Fprintf(os.Stderr, "unimplemented!\n")
		// xray.PingBatch()
	default:
		fmt.Printf("one of commands must be specified\n")
		flag.Usage()
	}
}
