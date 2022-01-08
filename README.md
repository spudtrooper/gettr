# gettr

Minimal API for http://gettr.com.

You need to supply the user and token. You can find your token from the dev console with the following:

```js
console.log(JSON.parse(localStorage.LS_SESSION_INFO).userinfo.token)
```