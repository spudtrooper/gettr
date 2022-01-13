package main

import (
	"flag"

	"github.com/pkg/errors"
	"github.com/spudtrooper/gettr/htmlgen"
	"github.com/spudtrooper/gettr/model"
	"github.com/spudtrooper/goutil/check"
	"github.com/spudtrooper/goutil/flags"
)

var (
	other                     = flags.String("other", "other username")
	limit                     = flags.Int("limit", "only include this many rows")
	threads                   = flags.Int("threads", "threads to calls")
	all                       = flags.Bool("all", "write all files and override inidividual flags")
	writeCSV                  = flags.Bool("write_csv", "write CSV file")
	writeHTML                 = flags.Bool("write_html", "write HTML file")
	writeSimpleHTML           = flags.Bool("write_simple_html", "write HTML file")
	writeDescriptionsHTML     = flags.Bool("write_desc_html", "write HTML file for entries with descriptions")
	writeTwitterFollowersHTML = flags.Bool("write_twitter_followers_html", "write HTML file for entries with twitter followers")
	outputDir                 = flag.String("output_dir", "../gettrdata/output", "output directory for files")
)

func realMain() error {
	if *other == "" {
		return errors.Errorf("--other required")
	}
	if *outputDir == "" {
		return errors.Errorf("--output_dir required")
	}

	factory, err := model.MakeFactoryFromFlags()
	if err != nil {
		return err
	}

	if err := htmlgen.Generate(*outputDir, factory, *other,
		htmlgen.GenerateLimit(*limit),
		htmlgen.GenerateAll(*all),
		htmlgen.GenerateThreads(*threads),
		htmlgen.GenerateWriteCSV(*writeCSV),
		htmlgen.GenerateWriteDescriptionsHTML(*writeDescriptionsHTML),
		htmlgen.GenerateWriteHTML(*writeHTML),
		htmlgen.GenerateWriteSimpleHTML(*writeSimpleHTML),
		htmlgen.GenerateWriteTwitterFollowersHTML(*writeTwitterFollowersHTML),
	); err != nil {
		return err
	}

	return nil
}

func main() {
	flag.Parse()
	check.Err(realMain())
}
