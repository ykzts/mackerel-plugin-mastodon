package mpmastodon

import (
	"flag"
	"fmt"
	"io"
	"strings"

	"net/http"
	"strconv"

	mp "github.com/mackerelio/go-mackerel-plugin-helper"
	"gopkg.in/xmlpath.v2"
)

var graphdef = map[string]mp.Graphs{
	"user": {
		Label: "Mastodon users",
		Unit:  "integer",
		Metrics: []mp.Metrics{
			{Name: "user_count", Label: "Confirmed users count", Diff: false},
		},
	},
	"toot": {
		Label: "Mastodon toots",
		Unit:  "integer",
		Metrics: []mp.Metrics{
			{Name: "toot_count", Label: "Posted toots count", Diff: false},
		},
	},
	"instance": {
		Label: "Mastodon instances",
		Unit:  "integer",
		Metrics: []mp.Metrics{
			{Name: "instance_count", Label: "Connected other instances count", Diff: false},
		},
	},
}

// MastodonPlugin mackerel plugin for mastodon
type MastodonPlugin struct {
	Host     string
	Tempfile string
	Prefix   string
}

// GraphDefinition interface for mackerel plugin
func (m MastodonPlugin) GraphDefinition() map[string]mp.Graphs {
	return graphdef
}

// MetricKeyPrefix interface for mackerel plugin
func (m MastodonPlugin) MetricKeyPrefix() string {
	if m.Prefix == "" {
		m.Prefix = "mastodon"
	}
	return m.Prefix
}

// FetchMetrics interface for mackerel plugin
func (m MastodonPlugin) FetchMetrics() (map[string]interface{}, error) {
	uri := fmt.Sprintf("https://%s/about/more", m.Host)
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return m.parseStats(resp.Body)
}

func (m MastodonPlugin) parseStats(body io.Reader) (map[string]interface{}, error) {
	stat := make(map[string]interface{})

	path := xmlpath.MustCompile("//*[@class='information-board']/*[@class='section']/strong")
	root, err := xmlpath.ParseHTML(body)
	if err != nil {
		return nil, err
	}
	iter := path.Iter(root)

	keys := []string{"user_count", "toot_count", "instance_count"}

	for _, key := range keys {
		if !iter.Next() {
			break
		}
		stat[key], err = parseCount(iter.Node().String())
		if err != nil {
			return nil, err
		}
	}

	return stat, nil
}

func parseCount(s string) (float64, error) {
	count := strings.Replace(s, ",", "", -1)
	return strconv.ParseFloat(count, 64)
}

// Do the plugin
func Do() {
	optHost := flag.String("host", "", "Host")
	optTempfile := flag.String("tempfile", "", "Temp file name")
	flag.Parse()

	var mastodon MastodonPlugin
	mastodon.Host = *optHost

	helper := mp.NewMackerelPlugin(mastodon)
	helper.Tempfile = *optTempfile
	helper.Run()
}
