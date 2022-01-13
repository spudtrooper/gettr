package htmlgen

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"sync"

	"github.com/spudtrooper/gettr/api"
	"github.com/spudtrooper/gettr/log"
	"github.com/spudtrooper/gettr/model"
	"github.com/spudtrooper/goutil/check"
	"github.com/spudtrooper/goutil/html"
	"github.com/spudtrooper/goutil/io"
	"github.com/spudtrooper/goutil/or"
)

var (
	debugResolvedUserInfo = flag.Bool("debug_resolved_user_info", false, "print verbose logs for resolving user info")
)

func Generate(outputDirName string, factory model.Factory, other string, gOpts ...GeneratOption) error {
	opts := MakeGeneratOptions(gOpts...)

	limit := opts.Limit()
	threads := or.Int(opts.Threads(), 100)

	log.Printf("using %d threads for HTML generation", threads)

	var users []*model.User

	u := factory.MakeUser(other)
	followers := make(chan *model.User)
	followersForResolution := make(chan *model.User)
	go func() {
		users, errs := u.Followers(api.AllFollowersThreads(threads))
		for u := range users {
			followers <- u
			followersForResolution <- u
		}
		for e := range errs {
			log.Printf("Followers: ignoring error: %v", e)
		}
		close(followers)
		close(followersForResolution)
	}()

	// Resolve the user info in multiple threads rather than doing it syncronously in the sort.
	{
		log.Printf("resolving user info...")
		go func() {
			var wg sync.WaitGroup
			for i := 0; i < threads; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for f := range followersForResolution {
						u, err := f.UserInfo()
						if err != nil {
							log.Printf("UserInfo: ignoring error: %v", err)
						}
						if *debugResolvedUserInfo {
							log.Printf("resolved userInfo: %v", u)
						}
					}
				}()
			}
			wg.Wait()
		}()
	}

	var cachedFollowers []*model.User
	for f := range followers {
		cachedFollowers = append(cachedFollowers, f)
	}
	log.Printf("sorting %d users...", len(cachedFollowers))
	sort.Slice(cachedFollowers, func(i, j int) bool {
		a, b := cachedFollowers[i], cachedFollowers[j]
		ai, err := a.UserInfo()
		if err != nil {
			log.Printf("a.UserInfo: ignoring error: %v", err)
			return true
		}
		bi, err := b.UserInfo()
		if err != nil {
			log.Printf("b.UserInfo: ignoring error: %v", err)
			return true
		}
		return ai.Followers() > bi.Followers()
	})

	// Put the other user first
	users = append(users, factory.MakeUser(other))
	users = append(users, cachedFollowers...)

	outDir, err := io.MkdirAll(outputDirName)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	if opts.All() || opts.WriteCSV() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			csvOutFile := path.Join(outDir, other+".csv")
			log.Printf("writing CSV to %s...", csvOutFile)

			csvFile, err := os.Create(csvOutFile)
			check.Err(err)
			defer csvFile.Close()
			w := csv.NewWriter(csvFile)
			w.Write([]string{
				"ICO",
				"BG",
				"USER",
				"DESCRIPTION",
				"GETTR FOLLOWERS",
				"GETTR FOLLOWING",
				"TWITTER FOLLOWERS",
				"TWITTER FOLLOWING",
				"GETTR+TWITTER FOLLOWERS",
				"FOLLOWERS % DIFF",
				"GETTR",
				"TWITTER",
			})
			for i, f := range users {
				if limit > 0 && i > limit {
					break
				}
				username := f.Username()
				userInfo, err := f.UserInfo()
				check.Err(err)

				ico := userInfo.ICO
				bg := userInfo.BGImg
				desc := userInfo.Desc
				followers := userInfo.Followers()
				following := userInfo.Following()
				twitterFollowers := userInfo.TwitterFollowers()
				twitterFollowing := userInfo.TwitterFollowing()
				fakeFollowers := followers + twitterFollowers
				var fakeFollowersPercDiff float64
				if followers > 0 {
					fakeFollowersPercDiff = float64(fakeFollowers-followers) / float64(followers) * 100.0
				}
				gettrURI := fmt.Sprintf("https://gettr.com/user/%s", username)
				twitterURI := fmt.Sprintf("https://twitter.com/%s", username)

				w.Write([]string{
					ico,
					bg,
					username,
					desc,
					fmt.Sprintf("%d", followers),
					fmt.Sprintf("%d", following),
					fmt.Sprintf("%d", twitterFollowers),
					fmt.Sprintf("%d", twitterFollowing),
					fmt.Sprintf("%d", fakeFollowers),
					fmt.Sprintf("%f", fakeFollowersPercDiff),
					gettrURI,
					twitterURI,
				})
			}
			w.Flush()

			log.Printf("wrote CSV to %s", csvOutFile)
		}()
	}

	createHTMLData := func(onlyNonEmptyDescs bool, onlyTwitterFollowers bool) (html.TableRowData, []html.TableRowData, error) {
		head := html.TableRowData{
			"USER",
			"DESCRIPTION",
			"GETTR FOLLOWERS",
			"GETTR FOLLOWING",
			"TWITTER FOLLOWERS",
			"TWITTER FOLLOWING",
			"GETTR+TWITTER FOLLOWERS",
			"FOLLOWERS % DIFF",
		}
		var rows []html.TableRowData
		for i, f := range users {
			if limit > 0 && i > limit {
				break
			}
			userInfo, err := f.UserInfo()
			if err != nil {
				return nil, nil, err
			}

			desc := userInfo.Desc
			if onlyNonEmptyDescs && desc == "" {
				continue
			}

			twitterFollowers := userInfo.TwitterFollowers()
			if onlyTwitterFollowers && twitterFollowers == 0 {
				continue
			}

			username := f.Username()
			followers := userInfo.Followers()
			following := userInfo.Following()
			twitterFollowing := userInfo.TwitterFollowing()
			fakeFollowers := followers + twitterFollowers
			var fakeFollowersPercDiff float64
			if followers > 0 {
				fakeFollowersPercDiff = float64(fakeFollowers-followers) / float64(followers) * 100.0
			}
			var user string
			{
				gettrURI := fmt.Sprintf("https://gettr.com/user/%s", username)
				gettrLink := fmt.Sprintf(`<a href="%s" target="_">getter</a>`, gettrURI)
				userLinks := "<b>" + username + "</b><br/>(" + gettrLink
				if twitterFollowers > 0 {
					twitterURI := fmt.Sprintf("https://twitter.com/%s", username)
					twitterLink := fmt.Sprintf(`<a href="%s" target="_">twitter</a>`, twitterURI)
					userLinks += " | " + twitterLink
				}
				userLinks += ")"

				ico := `<div style="width:30px; height:30px"></div>`
				if userInfo.ICO != "" {
					src := fmt.Sprintf("https://media.gettr.com/%s", userInfo.ICO)
					ico = fmt.Sprintf(`<img style="width:30px; height:30px" src="%s">`, src)
				}

				user += "<table border=0>"
				user += "<tr>"
				user += "<td>" + ico + "</td>"
				user += "<td>" + userLinks + "</td>"
				user += "</tr>"
				user += "</table>"
			}
			row := html.TableRowData{
				user,
				desc,
				fmt.Sprintf("%d", followers),
				fmt.Sprintf("%d", following),
				fmt.Sprintf("%d", twitterFollowers),
				fmt.Sprintf("%d", twitterFollowing),
				fmt.Sprintf("%d", fakeFollowers),
				fmt.Sprintf("%.2f%%", fakeFollowersPercDiff),
			}
			rows = append(rows, row)
		}
		return head, rows, nil
	}

	render := func(htmlData html.Data) string {
		html, err := html.Render(htmlData, html.RenderNoFormat(true))
		check.Err(err)
		return html
	}

	renderSimple := func(htmlData html.Data) string {
		html, err := html.RenderSimple(htmlData, html.RenderNoFormat(true))
		check.Err(err)
		return html
	}

	if opts.All() || opts.WriteSimpleHTML() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			htmlOutFile := path.Join(outDir, other+"_simple.html")
			log.Printf("writing simple HTML to %s...", htmlOutFile)

			head, rows, err := createHTMLData(false, false)
			check.Err(err)
			htmlData := html.Data{
				Entities: []html.DataEntity{
					html.MakeSimpleDataEntityFromTable(html.TableData{
						Head: head,
						Rows: rows,
					}),
				}}
			html := renderSimple(htmlData)
			check.Err(ioutil.WriteFile(htmlOutFile, []byte(html), 0755))
			log.Printf("wrote simple HTML to %s", htmlOutFile)
		}()
	}

	if opts.All() || opts.WriteDescriptionsHTML() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Printf("creating desc HTML...")

			head, rows, err := createHTMLData(true, false)
			check.Err(err)

			{
				htmlOutFile := path.Join(outDir, other+"_desc.html")
				log.Printf("writing desc HTML to %s...", htmlOutFile)

				htmlData := html.Data{
					Entities: []html.DataEntity{
						html.MakeDataEntityFromTable(html.TableData{
							Head: head,
							Rows: rows,
						}),
					}}
				html := render(htmlData)
				check.Err(ioutil.WriteFile(htmlOutFile, []byte(html), 0755))
				log.Printf("wrote desc HTML to %s", htmlOutFile)
			}
			{
				htmlOutFile := path.Join(outDir, other+"_desc_simple.html")
				log.Printf("writing desc simple HTML to %s...", htmlOutFile)

				htmlData := html.Data{
					Entities: []html.DataEntity{
						html.MakeSimpleDataEntityFromTable(html.TableData{
							Head: head,
							Rows: rows,
						}),
					}}
				html := renderSimple(htmlData)
				check.Err(ioutil.WriteFile(htmlOutFile, []byte(html), 0755))
				log.Printf("wrote desc simple HTML to %s", htmlOutFile)
			}
		}()
	}

	if opts.All() || opts.WriteTwitterFollowersHTML() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Printf("creating twitter followers HTML...")

			head, rows, err := createHTMLData(false, true)
			check.Err(err)

			{
				htmlOutFile := path.Join(outDir, other+"_twitter_followers.html")
				log.Printf("writing twitter followers HTML to %s...", htmlOutFile)

				htmlData := html.Data{
					Entities: []html.DataEntity{
						html.MakeDataEntityFromTable(html.TableData{
							Head: head,
							Rows: rows,
						}),
					}}
				html := render(htmlData)
				check.Err(ioutil.WriteFile(htmlOutFile, []byte(html), 0755))
				log.Printf("wrote twitter followers HTML to %s", htmlOutFile)
			}
			{
				htmlOutFile := path.Join(outDir, other+"_twitter_followers_simple.html")
				log.Printf("writing twitter followers simple HTML to %s...", htmlOutFile)

				htmlData := html.Data{
					Entities: []html.DataEntity{
						html.MakeSimpleDataEntityFromTable(html.TableData{
							Head: head,
							Rows: rows,
						}),
					}}
				html := renderSimple(htmlData)
				check.Err(ioutil.WriteFile(htmlOutFile, []byte(html), 0755))
				log.Printf("wrote twitter followers simple HTML to %s", htmlOutFile)
			}
		}()
	}

	if opts.All() || opts.WriteHTML() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			htmlOutFile := path.Join(outDir, other+".html")
			log.Printf("writing HTML to %s...", htmlOutFile)

			head, rows, err := createHTMLData(false, false)
			check.Err(err)
			htmlData := html.Data{
				Entities: []html.DataEntity{
					html.MakeDataEntityFromTable(html.TableData{
						Head: head,
						Rows: rows,
					}),
				}}
			html := render(htmlData)
			check.Err(ioutil.WriteFile(htmlOutFile, []byte(html), 0755))
			log.Printf("wrote HTML to %s", htmlOutFile)
		}()
	}

	wg.Wait()

	log.Println("done")

	return nil
}
