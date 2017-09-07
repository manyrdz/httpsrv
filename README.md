A plain static file server with auto-reload for html files.

### Usage:
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