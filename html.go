package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"

	"github.com/pkg/errors"
	"github.com/spudtrooper/gettr/api"
	"github.com/spudtrooper/gettr/model"
	"github.com/spudtrooper/goutil/check"
	"github.com/spudtrooper/goutil/html"
	"github.com/spudtrooper/goutil/io"
)

var (
	user                  = flag.String("user", "", "auth username")
	debug                 = flag.Bool("debug", false, "whether to debug requests")
	token                 = flag.String("token", "", "auth token")
	other                 = flag.String("other", "mtg4america", "other username")
	cacheDir              = flag.String("cache_dir", ".cache", "cache directory")
	userCreds             = flag.String("user_creds", ".user_creds.json", "file with user credentials")
	limit                 = flag.Int("limit", 0, "only include this many rows")
	writeCSV              = flag.Bool("write_csv", false, "write CSV file")
	writeHTML             = flag.Bool("write_html", false, "write HTML file")
	writeSimpleHTML       = flag.Bool("write_simple_html", false, "write HTML file")
	writeDescriptionsHTML = flag.Bool("write_desc_html", false, "write HTML file for entries with descriptions")
)

func realMain() error {
	var client *api.Client
	if *user != "" && *token != "" {
		client = api.MakeClient(*user, *token, api.MakeClientDebug(*debug))
	} else if *userCreds != "" {
		c, err := api.MakeClientFromFile(*userCreds, api.MakeClientDebug(*debug))
		check.Err(err)
		client = c
	} else {
		return errors.Errorf("Must set --user & --token or --creds_file")
	}

	cache := model.MakeCache(*cacheDir)
	factory := model.MakeFactory(cache, client)
	u := factory.MakeCachedUser(*other)

	followers := make(chan model.User)
	go func() {
		users, _ := u.Followers()
		for u := range users {
			followers <- u
		}
		close(followers)
	}()

	var cachedFollowers []model.User
	for f := range followers {
		cachedFollowers = append(cachedFollowers, factory.MakeCachedUser(f.Username()))
	}
	sort.Slice(cachedFollowers, func(i, j int) bool {
		a, b := cachedFollowers[i], cachedFollowers[j]
		ai, err := a.UserInfo()
		check.Err(err)
		bi, err := b.UserInfo()
		check.Err(err)
		return ai.Followers() > bi.Followers()
	})

	// Put the other user first
	var users []model.User
	users = append(users, factory.MakeCachedUser(*other))
	users = append(users, cachedFollowers...)

	outDir, err := io.MkdirAll("data")
	check.Err(err)

	if *writeCSV {
		csvOutFile := path.Join(outDir, *other+".csv")
		csvFile, err := os.Create(csvOutFile)
		check.Err(err)
		defer csvFile.Close()
		w := csv.NewWriter(csvFile)
		w.Write([]string{
			"USER",
			"DESCRIPTION",
			"FOLLOWERS",
			"FOLLOWING",
			"TWITTER FOLLOWERS",
			"TWITTER FOLLOWING",
			"FAKE FOLLOWERS",
			"GETTR",
			"TWITTER",
		})
		for i, f := range users {
			if i%1000 == 0 {
				log.Printf("%d/%d %.2f%%", i, len(cachedFollowers), 100.*float64(i)/float64(len(cachedFollowers)))
			}
			if *limit > 0 && i > *limit {
				break
			}
			username := f.Username()
			userInfo, err := f.UserInfo()
			check.Err(err)

			desc := userInfo.Desc
			followers := userInfo.Followers()
			following := userInfo.Following()
			twitterFollowers := userInfo.TwitterFollowers()
			twitterFollowing := userInfo.TwitterFollowing()
			fakeFollowers := followers + twitterFollowers
			gettrURI := fmt.Sprintf("https://gettr.com/user/%s", username)
			twitterURI := fmt.Sprintf("https://twitter.com/%s", username)

			w.Write([]string{
				username,
				desc,
				fmt.Sprintf("%d", followers),
				fmt.Sprintf("%d", following),
				fmt.Sprintf("%d", twitterFollowers),
				fmt.Sprintf("%d", twitterFollowing),
				fmt.Sprintf("%d", fakeFollowers),
				gettrURI,
				twitterURI,
			})
		}
		w.Flush()

		log.Printf("wrote CSV to %s", csvOutFile)
	}

	createHTMLData := func(onlyNonEmptyDescs bool) (html.TableRowData, []html.TableRowData) {
		head := html.TableRowData{
			"ICO",
			"GETTR",
			"TWITTER",
			"DESCRIPTION",
			"FOLLOWERS",
			"FOLLOWING",
			"TWITTER FOLLOWERS",
			"TWITTER FOLLOWING",
			"FAKE FOLLOWERS",
		}
		var rows []html.TableRowData
		for i, f := range users {
			if i%1000 == 0 {
				log.Printf("%d/%d %.2f%%", i, len(cachedFollowers), 100.*float64(i)/float64(len(cachedFollowers)))
			}
			if *limit > 0 && i > *limit {
				break
			}
			userInfo, err := f.UserInfo()
			check.Err(err)

			desc := userInfo.Desc
			if onlyNonEmptyDescs && desc == "" {
				continue
			}

			username := f.Username()
			gettrURI := fmt.Sprintf("https://gettr.com/user/%s", username)
			gettrLink := fmt.Sprintf(`<a href="%s" target="_">%s</a>`, gettrURI, username)
			twitterURI := fmt.Sprintf("https://twitter.com/%s", username)
			twitterLink := fmt.Sprintf(`<a href="%s" target="_">%s</a>`, twitterURI, username)
			followers := userInfo.Followers()
			following := userInfo.Following()
			twitterFollowers := userInfo.TwitterFollowers()
			twitterFollowing := userInfo.TwitterFollowing()
			fakeFollowers := followers + twitterFollowers
			var ico string
			if userInfo.ICO != "" {
				src := fmt.Sprintf("https://media.gettr.com/%s", userInfo.ICO)
				ico = fmt.Sprintf(`<img style="width:30px; height:30px" src="%s">`, src)
			}
			row := html.TableRowData{
				ico,
				gettrLink,
				twitterLink,
				desc,
				fmt.Sprintf("%d", followers),
				fmt.Sprintf("%d", following),
				fmt.Sprintf("%d", twitterFollowers),
				fmt.Sprintf("%d", twitterFollowing),
				fmt.Sprintf("%d", fakeFollowers),
			}
			rows = append(rows, row)
		}
		return head, rows
	}

	if *writeSimpleHTML {
		head, rows := createHTMLData(false)
		htmlData := html.Data{
			Entities: []html.DataEntity{
				html.MakeSimpleDataEntityFromTable(html.TableData{
					Head: head,
					Rows: rows,
				}),
			}}
		html, err := html.RenderSimple(htmlData)
		if err != nil {
			return err
		}

		htmlOutFile := path.Join(outDir, *other+"_simple.html")
		if err := ioutil.WriteFile(htmlOutFile, []byte(html), 0755); err != nil {
			return err
		}
		log.Printf("wrote simple HTML to %s", htmlOutFile)
	}

	if *writeDescriptionsHTML {
		head, rows := createHTMLData(true)
		htmlData := html.Data{
			Entities: []html.DataEntity{
				html.MakeSimpleDataEntityFromTable(html.TableData{
					Head: head,
					Rows: rows,
				}),
			}}
		html, err := html.RenderSimple(htmlData)
		if err != nil {
			return err
		}

		htmlOutFile := path.Join(outDir, *other+"_desc_simple.html")
		if err := ioutil.WriteFile(htmlOutFile, []byte(html), 0755); err != nil {
			return err
		}
		log.Printf("wrote desc simple HTML to %s", htmlOutFile)
	}

	if *writeHTML {
		head, rows := createHTMLData(false)
		htmlData := html.Data{
			Entities: []html.DataEntity{
				html.MakeDataEntityFromTable(html.TableData{
					Head: head,
					Rows: rows,
				}),
			}}
		html, err := html.Render(htmlData)
		if err != nil {
			return err
		}

		htmlOutFile := path.Join(outDir, *other+".html")
		if err := ioutil.WriteFile(htmlOutFile, []byte(html), 0755); err != nil {
			return err
		}
		log.Printf("wrote HTML to %s", htmlOutFile)
	}

	return nil
}

func main() {
	flag.Parse()
	check.Err(realMain())
}
