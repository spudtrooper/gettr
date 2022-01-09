# gettr

Minimal API for http://gettr.com.

You need to supply the user and token. You can find your token from the dev console with the following:

```js
console.log(JSON.parse(localStorage.LS_SESSION_INFO).userinfo.token)
```

To generate HTML for user `repmattgaetz`.

1. Create a config by copying [user_creds_example.json](user_creds_example.json) to `.user_creds.json` and filling in the TODOs, and then
2. Find followers and generate HTML:

        go run main.go --other repmattgaetz Persist PersistAll
        go run html.go --other repmattgaetz --all