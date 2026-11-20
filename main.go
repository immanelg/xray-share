package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"time"

	"sort"

	"encoding/json"
	"flag"
	"strings"

	"github.com/xtls/libxray/nodep"
	"github.com/xtls/libxray/share"
	"github.com/xtls/libxray/xray"

	// "github.com/xtls/xray-core/infra/conf"
	confjson "github.com/xtls/xray-core/infra/conf/json"
)

var version = "<devel>"
var commitHash = "<unknown commit>"

func todo() {
	log.Fatal("unimplemented!\n")
}
func die(message string) {
	log.Fatalf(fmtcolor("@R{%s}\n"), message)
}
func mustOk(message string, err error) {
	if err != nil {
		log.Fatalf(fmtcolor("@R{%s: %v}\n"), message, err)
	}
}
func logIfError(message string, err error) {
	if err != nil {
		log.Fatalf(fmtcolor("@R{%s: %v}\n"), message, err)
	}
}

func mustReadEntireFile(name string) []byte {
	log.Printf("reading %s", name)
	if name == "-" {
		log.Print("waiting for stdin...\n")
		content, errReadStdin := io.ReadAll(os.Stdin)
		mustOk("read stdin", errReadStdin)
		return content
	}
	path, errParseUrl := url.Parse(name)
	mustOk("parse file url", errParseUrl)
	
	if path.Scheme == "" {
		content, errReadIn := os.ReadFile(name)
		mustOk("read input file", errReadIn)
		return content
	}
	// TODO: -iretries
	httpClient := &http.Client{Timeout: 10*time.Second}
	req, err := http.NewRequest(http.MethodGet, name, nil)
	mustOk("create request", err)
	proxyURL, errProxyEnv := http.ProxyFromEnvironment(req)
	mustOk("get proxy env var", errProxyEnv)
	if proxyURL != nil { log.Printf("using proxy %s", proxyURL.String()) }
	response, errHttp := httpClient.Do(req)
	mustOk("read http file", errHttp)
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	return body
}

func mustWriteEntireFile(name string, content []byte) {
	log.Printf("writing %s", name)
	if name == "-" {
		fmt.Print(string(content))
		return
	}
	fh, errCreat := os.Create(name)
	mustOk("create", errCreat)
	defer fh.Close()

	_, errWrite := fh.Write(content)
	mustOk("write", errWrite)
}

var gTagCounter int
func subToJson(links []byte, normalizeTags bool) json.RawMessage {
	cfgjson, errconv := share.ConvertShareLinksToXrayJson(string(links), "")
	mustOk("convert", errconv)

	cfg, errjson := json.MarshalIndent(cfgjson, "", "  ")
	mustOk("marshal", errjson)

	if normalizeTags {
		re := regexp.MustCompile(`"tag":\s*"([^"]*)"`)
		cfg = re.ReplaceAllFunc(cfg, func(tag []byte) []byte {
			gTagCounter++
			return []byte(fmt.Sprintf(`"tag": "proxy-#%d-%s-%s"`, gTagCounter, rand.Text(), re.FindSubmatch(tag)[1]))
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
	flag.BoolVar(&printVersion, "V", false, "")
	flag.BoolVar(&configToUrl, "C", false, "")
	flag.BoolVar(&urlToConfig, "U", false, "")
	flag.BoolVar(&testConfigs, "T", false, "")

	// parameters
	var inFile string
	var outFile string
	var normalizeTags bool = true // XXX: configs from retarded sources have duplicate tags, which is a problem for xray. this flag controls our dumb substitution for tags.
	var testUrl string
	var testFiltersBelowTimeout int64
	var testFiltersAboveTimeout int64
	flag.StringVar(&inFile, "i", "-", "")
	flag.StringVar(&outFile, "o", "-", "")
	flag.BoolVar(&normalizeTags, "uniqtags", true, "")
	flag.StringVar(&testUrl, "turl", "https://gstatic.com/generate_204", "")
	flag.Int64Var(&testFiltersAboveTimeout, "tmax", 3000, "")
	flag.Int64Var(&testFiltersBelowTimeout, "tmin", 100, "")

	flag.Usage = func() {
		w := flag.CommandLine.Output()

		fmt.Fprint(w, fmtcolor(
`@g{Usage:}
  @Y{xray-share -V}
    @g{Print version info}
  @Y{xray-share -U [-i URL_SUBS=-] [-o FILEPATH_JSON=-] [-uniqtags]}
    @g{Convert share links to json config}
  @Y{xray-share -C [-i URL_JSON=-] [-o FILEPATH_SUBS=-]}
    @g{Convert json config to share links}
  @Y{xray-share -T [-tmax MS] [-tmin MS] [-turl URL=https://gstatic.com/generate_204] [-i URL_JSON_OR_SUBS=-] [-o FILEPATH_JSON=-] [-uniqtags]}
    @g{Filter working configs (json OR share links) and output json}
`))

		os.Exit(1)
	}

	flag.Parse()

	if len(flag.Args()) > 0 {
		fmt.Print(fmtcolor("@R{trailing arguments}\n"))
		flag.Usage()
	}

	c := 0
	if printVersion { c++ }
	if urlToConfig { c++ }
	if configToUrl { c++ }
	if testConfigs { c++ }
	if c != 1 {
		fmt.Printf(fmtcolor("@R{exactly one of commands must be specified, but %d were provided}\n\n"), c)
		flag.Usage()
	}

	switch {
	case printVersion:
		fmt.Printf(fmtcolor("@G{%s} @g{%s %s %s %s %s}\n"), os.Args[0], version, commitHash, runtime.Version(), runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	case urlToConfig:
		links := mustReadEntireFile(inFile)

		// FIXME: Additionally, produced config isn't useful because Observatory selector wants you
		// to **have a common prefix** like `proxy-` for rotating configs. 
		// How do we edit this without shitting nulls into this json? 
		// go encoder is retarded and will insert nulls which causes errors in xray interpreter.
		cfg := subToJson(links, normalizeTags)
		
		mustWriteEntireFile(outFile, cfg)
	case configToUrl:
		cfg := mustReadEntireFile(inFile)

		links, errconv := share.ConvertXrayJsonToShareLinks(cfg)
		mustOk("convert", errconv)

		mustWriteEntireFile(outFile, []byte(links))

	case testConfigs:
		configIn := mustReadEntireFile(inFile)

		// we want json input file. if it's not, try to normalize it first by converting from share links.
		{
			// if nil != json.NewDecoder(
			// 	&confjson.Reader{ Reader: strings.NewReader(string(configIn))},
			// ).Decode(
			// 	&struct{Outbounds []json.RawMessage `json:"outbounds"`}{},
			// ) {
			// 	configIn = subToJson(configIn, normalizeTags)
			// }
			if !json.Valid(configIn) {
			 	configIn = subToJson(configIn, normalizeTags)
			}
		}
		configString := string(configIn)

		// NOTE: This is the config we will use to write output JSON. keep the message as-is.
		// if we use Outbounds []OutboundDetour something something,
		// if generate a bunch of garbage nulls (because xray doesnt put omitempty)
		type configOut struct { 
			Outbounds []json.RawMessage `json:"outbounds,omitempty"` 
		}
		var config configOut
		mustOk("decode raw config", 
			json.NewDecoder(&confjson.Reader{Reader: strings.NewReader(configString)}).Decode(&config))

		var tags []string
		for _, outbound := range config.Outbounds {
			var t struct{ Tag string `json:"tag"` }
			json.Unmarshal(outbound, &t)
			tags = append(tags, t.Tag)
		}

		pingItems := []xray.PingBatchItem{}
		for i := range config.Outbounds {
			pingItems = append(pingItems, xray.PingBatchItem{
				XrayJSON: configString,
				OutboundTag: tags[i],
			})
		}

		pingResults := []xray.PingBatchResult{}

		const BATCH = 5
		for start := 0; start < len(pingItems); start += BATCH {
			end := min(start+BATCH, len(pingItems))
			batchOfPingItems := pingItems[start:end]
			batchOfPingResults, errPing := xray.PingBatch(batchOfPingItems, max(1, int(testFiltersAboveTimeout/1000)), testUrl)
			if errPing != nil {
				batchOfPingResults = make([]xray.PingBatchResult, len(batchOfPingItems))
				// NOTE: len(pingResults) stays in sync with len(tags).
			}
			logIfError("ping batch", errPing)
			pingResults = append(pingResults, batchOfPingResults...)
		}

		type outbound struct{o json.RawMessage; d int64}; goodOutbounds := []outbound{}
		for i, pingResult := range pingResults {
			log.Printf("Tag: %s\n", tags[i])

			var delayString string
			switch pingResult.Delay {
			case nodep.PingDelayError:   delayString = "(error)"
			case nodep.PingDelayTimeout: delayString = "(timeout)"
			default:                     delayString = strconv.FormatInt(pingResult.Delay, 10)+"ms"
			}

			if pingResult.Success && 
			pingResult.Delay < testFiltersAboveTimeout && 
			pingResult.Delay > testFiltersBelowTimeout {
				log.Printf(fmtcolor("\n@G{Verdict: GOOD\nDelay: %v}\n\n"), delayString)
				goodOutbounds = append(goodOutbounds, outbound{config.Outbounds[i], pingResult.Delay})
			} else {
				log.Printf(fmtcolor("\n@R{Verdict: BAD\nDelay: %v\nError: %v}\n\n"), delayString, pingResult.Error)
			}
		}
		log.Printf("filtered %d good configs", len(goodOutbounds))
		sort.Slice(goodOutbounds, func(i, j int) bool { return goodOutbounds[i].d < goodOutbounds[j].d })
		var configGood configOut
		for _, outbound := range goodOutbounds { configGood.Outbounds = append(configGood.Outbounds, outbound.o) } 
		configGoodBytes, _ := json.MarshalIndent(configGood, "", "    ")
		mustWriteEntireFile(outFile, configGoodBytes)
	}
}
