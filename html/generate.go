package html

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"

	"github.com/spudtrooper/gettr/api"
	"github.com/spudtrooper/gettr/model"
	"github.com/spudtrooper/goutil/check"
	"github.com/spudtrooper/goutil/html"
	"github.com/spudtrooper/goutil/io"
)

func Generate(client *api.Client, cache model.Cache, other string, gOpts ...GeneratOption) error {
	opts := MakeGeneratOptions(gOpts...)

	limit := opts.Limit()

	var users []model.User

	factory := model.MakeFactory(cache, client)
	u := factory.MakeCachedUser(other)
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
	log.Printf("Sorting %d users...", len(cachedFollowers))
	sort.Slice(cachedFollowers, func(i, j int) bool {
		a, b := cachedFollowers[i], cachedFollowers[j]
		ai, err := a.UserInfo()
		check.Err(err)
		bi, err := b.UserInfo()
		check.Err(err)
		return ai.Followers() > bi.Followers()
	})

	// Put the other user first
	users = append(users, factory.MakeCachedUser(other))
	users = append(users, cachedFollowers...)

	bytes, err := json.Marshal(users)
	if err != nil {
		return err
	}
	if err := cache.SetBytes(bytes, "users", other, "sortedFollowers"); err != nil {
		return err
	}

	outDir, err := io.MkdirAll("data")
	if err != nil {
		return err
	}

	if opts.WriteCSV() {
		csvOutFile := path.Join(outDir, other+".csv")
		log.Printf("writing CSV to %s...", csvOutFile)

		csvFile, err := os.Create(csvOutFile)
		check.Err(err)
		defer csvFile.Close()
		w := csv.NewWriter(csvFile)
		w.Write([]string{
			"ICO",
			"USER",
			"DESCRIPTION",
			"FOLLOWERS",
			"FOLLOWING",
			"TWITTER FOLLOWERS",
			"TWITTER FOLLOWING",
			"FAKE FOLLOWERS",
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
			if err != nil {
				return err
			}

			var ico string
			if userInfo.ICO != "" {
				src := fmt.Sprintf("https://media.gettr.com/%s", userInfo.ICO)
				ico = fmt.Sprintf(`<img style="width:30px; height:30px" src="%s">`, src)
			}
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
	}

	createHTMLData := func(onlyNonEmptyDescs bool, onlyTwitterFollowers bool) (html.TableRowData, []html.TableRowData, error) {
		head := html.TableRowData{
			"USER",
			"DESCRIPTION",
			"FOLLOWERS",
			"FOLLOWING",
			"TWITTER FOLLOWERS",
			"TWITTER FOLLOWING",
			"FAKE FOLLOWERS",
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

	if opts.WriteSimpleHTML() {
		htmlOutFile := path.Join(outDir, other+"_simple.html")
		log.Printf("writing simple HTML to %s...", htmlOutFile)

		head, rows, err := createHTMLData(false, false)
		if err != nil {
			return err
		}
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

		if err := ioutil.WriteFile(htmlOutFile, []byte(html), 0755); err != nil {
			return err
		}
		log.Printf("wrote simple HTML to %s", htmlOutFile)
	}

	if opts.WriteDescriptionsHTML() {
		log.Printf("creating desc HTML...")

		head, rows, err := createHTMLData(true, false)
		if err != nil {
			return err
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
			html, err := html.RenderSimple(htmlData)
			if err != nil {
				return err
			}

			if err := ioutil.WriteFile(htmlOutFile, []byte(html), 0755); err != nil {
				return err
			}
			log.Printf("wrote desc simple HTML to %s", htmlOutFile)
		}
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
			html, err := html.Render(htmlData)
			if err != nil {
				return err
			}

			if err := ioutil.WriteFile(htmlOutFile, []byte(html), 0755); err != nil {
				return err
			}
			log.Printf("wrote desc HTML to %s", htmlOutFile)
		}
	}

	if opts.WriteTwitterFollowersHTML() {
		log.Printf("creating twitter followers HTML...")

		head, rows, err := createHTMLData(false, true)
		if err != nil {
			return err
		}

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
			html, err := html.Render(htmlData)
			if err != nil {
				return err
			}

			if err := ioutil.WriteFile(htmlOutFile, []byte(html), 0755); err != nil {
				return err
			}
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
			html, err := html.RenderSimple(htmlData)
			if err != nil {
				return err
			}

			if err := ioutil.WriteFile(htmlOutFile, []byte(html), 0755); err != nil {
				return err
			}
			log.Printf("wrote twitter followers simple HTML to %s", htmlOutFile)
		}
	}

	if opts.WriteHTML() {
		htmlOutFile := path.Join(outDir, other+".html")
		log.Printf("writing HTML to %s...", htmlOutFile)

		head, rows, err := createHTMLData(false, false)
		if err != nil {
			return err
		}
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

		if err := ioutil.WriteFile(htmlOutFile, []byte(html), 0755); err != nil {
			return err
		}
		log.Printf("wrote HTML to %s", htmlOutFile)
	}

	return nil
}
