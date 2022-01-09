# gettr

Minimal API for http://gettr.com.

You need to supply the user and token. You can find your token from the dev console with the following:

```js
console.log(JSON.parse(localStorage.LS_SESSION_INFO).userinfo.token)
```

To generate HTML for user `repmattgaetz`

```bash
USER=repmattgaetz go run html.go --other $USER --write_twitter_followers_html  --write_desc_html --write_simple_html --write_html --write_csv
```