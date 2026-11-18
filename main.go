package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"strconv"

	"sort"

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
func softAssertSuccess(message string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, ansi("@R{%s: %v}\n"), message, err)
	}
}

func mustReadEntireFile(name string) []byte {
	path, errParseUrl := url.Parse(name)
	assertSuccess("parse file url", errParseUrl)
	
	if path.Scheme == "" {
		content, errReadIn := os.ReadFile(name)
		assertSuccess("read input file", errReadIn)
		return content
	}
	hclient := http.DefaultClient
	response, errReadIn := hclient.Get(name)
	assertSuccess("read remote file", errReadIn)
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	return body
}

func mustWriteEntireFile(name string, content []byte) {
	fh, errCreat := os.Create(name)
	assertSuccess("create", errCreat)
	defer fh.Close()

	_, errWrite := fh.Write(content)
	assertSuccess("write", errWrite)
}


func subToJson(links []byte, normalizeTags bool) json.RawMessage {
	cfgjson, errconv := share.ConvertShareLinksToXrayJson(string(links), "")
	assertSuccess("convert", errconv)

	cfg, errjson := json.MarshalIndent(cfgjson, "", "  ")
	assertSuccess("marshal", errjson)

	if normalizeTags {
		counter := 0
		re := regexp.MustCompile(`"tag":\s*"([^"]*)"`)
		cfg = re.ReplaceAllFunc(cfg, func(tag []byte) []byte {
			counter += 1
			return []byte(fmt.Sprintf(`"tag": "proxy-#%d-%s-%s"`, counter, rand.Text(), re.FindSubmatch(tag)[1]))
		})
	}


	return cfg
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
	var normalizeTags bool = true; _ = normalizeTags
	var testUrl string
	var testFiltersBelowTimeout int64
	var testFiltersAboveTimeout int64
	flag.StringVar(&inFile, "i", "", "")
	flag.StringVar(&outFile, "o", "", "")
	flag.BoolVar(&normalizeTags, "uniqtags", true, "")
	flag.StringVar(&testUrl, "turl", "https://gstatic.com/generate_204", "")
	flag.Int64Var(&testFiltersAboveTimeout, "tmax", 5000, "")
	flag.Int64Var(&testFiltersBelowTimeout, "tmin", 100, "")

	flag.Usage = func() {
		w := flag.CommandLine.Output()

		fmt.Fprint(w, ansi(
`@Y{-V}
  Print version info
@Y{-U -i URL_SUBS [-o FILEPATH_JSON=generated-config.json] [-uniqtags]}
  Convert share links to json config
@Y{-C -i URL_JSON [-o FILEPATH_SUBS=generated-links.txt]}
  Convert json config to share links
@Y{-T [-tmax MS] [-tmin MS] [-turl URL=https://gstatic.com/generate_204] -i URL_JSON_OR_SUBS [-o FILEPATH_JSON=generated-config.json] [-uniqtags]}
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

		// FIXME: configs from retarded sources **have duplicate tags**, which is a problem for xray.
		// FIXME: Additionally, produced config isn't useful because Observatory selector wants you
		// to **have a common prefix** like `proxy-` for rotating configs. 
		// How do we edit this without shitting nulls into this json? 
		// go encoder is retarded and will insert nulls which causes errors in xray interpreter.
		cfg := subToJson(links, normalizeTags)
		
		if outFile == "" { outFile = "generated-config.json" }
		mustWriteEntireFile(outFile, cfg)
	case configToUrl:
		if inFile == "" { die("no input file") }
		cfg := mustReadEntireFile(inFile)

		links, errconv := share.ConvertXrayJsonToShareLinks(cfg)
		assertSuccess("convert", errconv)

		if outFile == "" { outFile = "generated-links.txt" }
		mustWriteEntireFile(outFile, []byte(links))

	case testConfigs:
		if inFile == "" { die("no input file") }

		configIn := mustReadEntireFile(inFile)

		// we want json input file. if it's not, try to normalize it first by converting from share links.
		{
			if nil != json.NewDecoder(&confjson.Reader{Reader: strings.NewReader(string(configIn))}).Decode(&struct { Outbounds []json.RawMessage `json:"outbounds"` }{}) {
				var errJson error
				configIn, errJson = json.Marshal(subToJson(configIn, normalizeTags))
			    assertSuccess("marshal json", errJson)
			}
		}
		configString := string(configIn)

		// NOTE: This is the config we will use internally for accessing fields.
		// ...
		type pingOutboundConfig struct { Outbounds []conf.OutboundDetourConfig `json:"outbounds"` }
		var config pingOutboundConfig
		assertSuccess("decode config", 
			json.NewDecoder(&confjson.Reader{Reader: strings.NewReader(configString)}).Decode(&config))

		// ...
		// NOTE: This is the config we will use to write output JSON. keep the message as-is, don't 
		// generate a bunch of garbage nulls (because Marshal will generate garbage)
		type pingOutboundConfigRaw struct { 
			Outbounds []json.RawMessage `json:"outbounds,omitempty"` 
		}
		var configRaw pingOutboundConfigRaw
		assertSuccess("decode raw config", 
			json.NewDecoder(&confjson.Reader{Reader: strings.NewReader(configString)}).Decode(&configRaw))

		pingItems := []xray.PingBatchItem{}
		for _, outbound := range config.Outbounds {
			pingItems = append(pingItems, xray.PingBatchItem{
				XrayJSON: configString,
				OutboundTag: outbound.Tag,
				// FIXME: configs from retarded sources have duplicate tags, which this thing obviously rejects. 
			})
		}

		pingResults := []xray.PingBatchResult{}

		const BATCH = 5
		for start := 0; start < len(pingItems); start += BATCH {
			end := min(start+BATCH, len(pingItems))
			batchOfPingItems := pingItems[start:end]
			batchOfPingResults, errPing := xray.PingBatch(batchOfPingItems, int(testFiltersAboveTimeout/1000), testUrl)
			softAssertSuccess("ping batch", errPing)
			pingResults = append(pingResults, batchOfPingResults...)
		}

		var configGood pingOutboundConfigRaw
		type outbound struct{o json.RawMessage; d int64}; outbounds := []outbound{}
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
				outbounds = append(outbounds, outbound{configRaw.Outbounds[i], pingResult.Delay})
				configGood.Outbounds = append(configGood.Outbounds, configRaw.Outbounds[i])
			} else {
				fmt.Printf(ansi("@R{Verdict: BAD\nDelay: %v\nError: %v}\n\n"), delayString, pingResult.Error)
			}
		}
		sort.SliceStable(outbounds, func(i, j int) bool { return outbounds[i].d < outbounds[j].d })
		for _, outbound := range outbounds { configRaw.Outbounds = append(configRaw.Outbounds, outbound.o) } 
		configGoodBytes, _ := json.MarshalIndent(configGood, "", "    ")
		if outFile == "" { outFile = "generated-config.json" }
		mustWriteEntireFile(outFile, configGoodBytes)
	default:
		fmt.Printf("one of commands must be specified\n")
		flag.Usage()
	}
}
