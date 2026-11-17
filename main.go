package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"

	// "crypto/rand"
	"encoding/json"
	"flag"
	"strings"

	"github.com/xtls/libxray/share"
	"github.com/xtls/libxray/xray"

	"github.com/xtls/xray-core/infra/conf"
	confJSON "github.com/xtls/xray-core/infra/conf/json"
)

const version = "<devel>"
const commitHash = "<unknown commit>"

func todo() {
	fmt.Fprintf(os.Stderr, "unimplemented!\n")
}
func die(message string) {
	fmt.Fprintf(os.Stderr, "%s\n", message)
}
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
	var testFiltersBelowTimeout int64
	var testFiltersAboveTimeout int64
	flag.StringVar(&inFile, "i", "", "in file")
	flag.StringVar(&outFile, "o", "", "out file")
	flag.StringVar(&testUrl, "turl", "https://gstatic.com/generate_204", "test max timeout")
	flag.Int64Var(&testFiltersAboveTimeout, "tmax", 5000, "test max timeout")
	flag.Int64Var(&testFiltersBelowTimeout, "tmin", 100, "test min timeout")

	flag.Usage = func() {
		w := flag.CommandLine.Output()

		me := filepath.Base(os.Args[0])
		fmt.Fprintf(w, "Commands:\n")
		fmt.Fprintf(w, "%s -V\n", me)
		fmt.Fprintf(w, "\tPrint version.\n")
		fmt.Fprintf(w, "%s -U -i URL_SUBS [-o FILEPATH_JSON=generated-config.json]\n", me)
		fmt.Fprintf(w, "\tConvert share links to json config.\n")
		fmt.Fprintf(w, "%s -C -i URL_JSON [-o FILEPATH_SUBS=generated-links.txt]\n", me)
		fmt.Fprintf(w, "\tConvert json config to share links.\n")
		fmt.Fprintf(w, "%s -T [-tmax MS] [-tmin MS] [-turl URL] -i URL_JSON [-o FILEPATH_JSON=generated-config.json]\n", me)
		fmt.Fprintf(w, "\tFilter good configs.\n")
	}

	flag.Parse()

	switch {
	case printVersion:
		fmt.Println("xray-sharetool", version, commitHash, runtime.Version(), runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	case urlToConfig:
		if inFile == "" { die("no input file") }
		links := mustReadEntireFile(inFile)

		cfgjson, errconv := share.ConvertShareLinksToXrayJson(string(links), "")
		assertSuccess("cannot convert", errconv)

		cfg, errjson := json.MarshalIndent(cfgjson, "", "  ")
		assertSuccess("cannot marshal", errjson)

		if outFile == "" { outFile = "generated-config.json" }
		mustWriteEntireFile(outFile, cfg)
	case configToUrl:
		if inFile == "" { die("no input file") }
		cfg := mustReadEntireFile(inFile)

		links, errconv := share.ConvertXrayJsonToShareLinks(cfg)
		assertSuccess("cannot convert", errconv)

		if outFile == "" { outFile = "generated-links.txt" }
		mustWriteEntireFile(outFile, []byte(links))

	case testConfigs:
		// TODO hm. input can be both formats: sub.txt or sub.json from remote host / local file
		if inFile == "" { die("no input file") }

		configString := string(mustReadEntireFile(inFile))

		type pingOutboundConfig struct {
			Outbounds []conf.OutboundDetourConfig `json:"outbounds"`
		}
		var config pingOutboundConfig
		decoder := json.NewDecoder(&confJSON.Reader{Reader: strings.NewReader(configString)})
		assertSuccess("decode config", decoder.Decode(&config))

		pingItems := []xray.PingBatchItem{}
		for _, outbound := range config.Outbounds {
			pingItems = append(pingItems, xray.PingBatchItem{
				XrayJSON: configString,
				OutboundTag: outbound.Tag,
			})
		}
		pingResults, errPing := xray.PingBatch(pingItems, int(testFiltersAboveTimeout/1000), testUrl)
		assertSuccess("cannot ping batch", errPing)
		var configGood pingOutboundConfig
		for i, pingResult := range pingResults {
			outbound := config.Outbounds[i]
			fmt.Printf(">>> testing %s\n", outbound.Tag)
			if pingResult.Success && 
			   pingResult.Delay < testFiltersAboveTimeout && 
			   pingResult.Delay > testFiltersBelowTimeout {
				fmt.Printf("GOOD\n")
				configGood.Outbounds = append(configGood.Outbounds, outbound)
			} else {
				fmt.Printf("BAD\n")
				fmt.Printf("Delay: %v\n", pingResult.Delay)
				fmt.Printf("Error: %v\n", pingResult.Error)
			}
		}
		configGoodBytes, _ := json.MarshalIndent(configGood, "", "    ")
		if outFile == "" { outFile = "generated-config.json" }
		mustWriteEntireFile(outFile, configGoodBytes)
	default:
		fmt.Printf("one of commands must be specified\n")
		flag.Usage()
	}
}
