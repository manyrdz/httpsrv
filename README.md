A plain static file server with auto-reload for html files.

### To install
`go get github.com/manyrdz/httpsrv`

### To use
`httpsrv [ -help | -d <dir> | -h <host> ]`
Where <dir> is a relative directory to the current `pwd` (default is current directory); and <host> is in the form of `ip:port` (default: localhost:4000)

### Notes:
httpsrv uses the text/template golang package to inject a javascript script which enables the auto-reload feature. So it needs the presence of the string `{{.Script}}` just before `</body>`
```
<!DOCTYPE html>
<html>
<body>
  Hello, world!
  {{.Script}}
</body>
</html>
```
