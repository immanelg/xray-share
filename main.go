package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"

	"encoding/json"
	"flag"
	"strings"

	"github.com/xtls/libxray/nodep"
	"github.com/xtls/libxray/share"
	"github.com/xtls/libxray/xray"

	"github.com/xtls/xray-core/infra/conf"
	confjson "github.com/xtls/xray-core/infra/conf/json"
)

var version = "<devel>"
var commitHash = "<unknown commit>"

func todo() {
	fmt.Fprintf(os.Stderr, "unimplemented!\n")
}
func die(message string) {
	fmt.Fprintf(os.Stderr, ansi("@R{%s}\n"), message)
	os.Exit(1)
}
func assertSuccess(message string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, ansi("@R{%s: %v}\n"), message, err)
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
	flag.StringVar(&inFile, "i", "", "")
	flag.StringVar(&outFile, "o", "", "")
	flag.StringVar(&testUrl, "turl", "https://gstatic.com/generate_204", "")
	flag.Int64Var(&testFiltersAboveTimeout, "tmax", 5000, "")
	flag.Int64Var(&testFiltersBelowTimeout, "tmin", 100, "")

	flag.Usage = func() {
		w := flag.CommandLine.Output()

		fmt.Fprint(w, ansi(
`@Y{-V}
  Print version info
@Y{-U -i URL_SUBS [-o FILEPATH_JSON=generated-config.json]}
  Convert share links to json config
@Y{-C -i URL_JSON [-o FILEPATH_SUBS=generated-links.txt]}
  Convert json config to share links
@Y{-T [-tmax MS] [-tmin MS] [-turl URL=https://gstatic.com/generate_204] -i URL_JSON_OR_SUBS [-o FILEPATH_JSON=generated-config.json]}
  Filter good configs
`))
	}

	flag.Parse()

	switch {
	case printVersion:
		fmt.Printf("%s %s %s %s %s %s\n", os.Args[0], version, commitHash, runtime.Version(), runtime.GOOS, runtime.GOARCH)
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
		if inFile == "" { die("no input file") }

		configString := string(mustReadEntireFile(inFile))

		{
			var c struct { Outbounds []json.RawMessage `json:"outbounds"` }
			if nil != json.NewDecoder(&confjson.Reader{Reader: strings.NewReader(configString)}).Decode(&c) {
				rawMsg, errConv := share.ConvertShareLinksToXrayJson(configString, "")
			    assertSuccess("convert share links to json", errConv)
				configBytes, errJson := json.Marshal(rawMsg)
			    assertSuccess("marshal json", errJson)
				configString = string(configBytes)
			}
		}

		// NOTE: This is the config we will use internally for accessing fields.
		// ...
		type pingOutboundConfig struct { Outbounds []conf.OutboundDetourConfig `json:"outbounds"` }
		var config pingOutboundConfig
		assertSuccess("decode config", 
			json.NewDecoder(&confjson.Reader{Reader: strings.NewReader(configString)}).Decode(&config))

		// ...
		// NOTE: This is the config we will use to write output JSON. keep the message as-is, don't 
		// generate a bunch of garbage nulls (because Marshal will generate garbage)
		type pingOutboundConfigRaw struct { Outbounds []json.RawMessage `json:"outbounds"` }
		var configRaw pingOutboundConfigRaw
		assertSuccess("decode raw config", 
			json.NewDecoder(&confjson.Reader{Reader: strings.NewReader(configString)}).Decode(&configRaw))

		pingItems := []xray.PingBatchItem{}
		for _, outbound := range config.Outbounds {
			pingItems = append(pingItems, xray.PingBatchItem{
				XrayJSON: configString,
				OutboundTag: outbound.Tag,
			})
		}

		pingResults := []xray.PingBatchResult{}

		const BATCH = 5
		for start := 0; start < len(pingItems); start += BATCH {
			end := min(start+BATCH, len(pingItems))
			batchOfPingItems := pingItems[start:end]
			batchOfPingResults, errPing := xray.PingBatch(batchOfPingItems, int(testFiltersAboveTimeout/1000), testUrl)
			assertSuccess("cannot ping batch", errPing)
			pingResults = append(pingResults, batchOfPingResults...)
		}

		var configGood pingOutboundConfigRaw
		for i, pingResult := range pingResults {
			fmt.Printf("Tag: %s\n", config.Outbounds[i].Tag)

			var delayString string
			switch pingResult.Delay {
			case nodep.PingDelayError:   delayString = "(error)"
			case nodep.PingDelayTimeout: delayString = "(timeout)"
			default:                     delayString = strconv.FormatInt(pingResult.Delay, 10)+"ms"
			}

			if pingResult.Success && 
			pingResult.Delay < testFiltersAboveTimeout && 
			pingResult.Delay > testFiltersBelowTimeout {
				fmt.Printf(ansi("@G{Verdict: GOOD\nDelay: %v}\n\n"), delayString)
				configGood.Outbounds = append(configGood.Outbounds, configRaw.Outbounds[i])
			} else {
				fmt.Printf(ansi("@R{Verdict: BAD\nDelay: %v\nError: %v}\n\n"), delayString, pingResult.Error)
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
