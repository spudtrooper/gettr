package main

import (
	"flag"

	"github.com/spudtrooper/gettr/api"
	html "github.com/spudtrooper/gettr/htmlgen"
	"github.com/spudtrooper/gettr/model"
	"github.com/spudtrooper/goutil/check"
)

var (
	other                     = flag.String("other", "mtg4america", "other username")
	limit                     = flag.Int("limit", 0, "only include this many rows")
	threads                   = flag.Int("threads", 0, "threads to calls")
	all                       = flag.Bool("all", false, "write all files and override inidividual flags")
	writeCSV                  = flag.Bool("write_csv", false, "write CSV file")
	writeHTML                 = flag.Bool("write_html", false, "write HTML file")
	writeSimpleHTML           = flag.Bool("write_simple_html", false, "write HTML file")
	writeDescriptionsHTML     = flag.Bool("write_desc_html", false, "write HTML file for entries with descriptions")
	writeTwitterFollowersHTML = flag.Bool("write_twitter_followers_html", false, "write HTML file for entries with twitter followers")
	outputDir                 = flag.String("output_dir", "../gettrdata/output", "output directory for files")
)

func realMain() error {
	client, err := api.MakeClientFromFlags()
	if err != nil {
		return err
	}
	cache, err := model.MakeCacheFromFlags()
	if err != nil {
		return err
	}

	if err := html.Generate(*outputDir, client, cache, *other,
		html.GenerateLimit(*limit),
		html.GenerateAll(*all),
		html.GenerateThreads(*threads),
		html.GenerateWriteCSV(*writeCSV),
		html.GenerateWriteDescriptionsHTML(*writeDescriptionsHTML),
		html.GenerateWriteHTML(*writeHTML),
		html.GenerateWriteSimpleHTML(*writeSimpleHTML),
		html.GenerateWriteTwitterFollowersHTML(*writeTwitterFollowersHTML),
	); err != nil {
		return err
	}

	return nil
}

func main() {
	flag.Parse()
	check.Err(realMain())
}
