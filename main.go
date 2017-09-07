package main

import (
	"bytes"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/gorilla/websocket"

	"github.com/fsnotify/fsnotify"
)

var (
	d = flag.String("d", "", "Set the served directory")
	h = flag.String("h", "localhost:4000", "Set the listening host")

	script string
)

var (
	signals = make(chan *string)
	errors  = make(chan *string)
	connbuf = make(map[*websocket.Conn]bool)
)

func main() {
	flag.Parse()
	parse()
	go watch()
	go alert()
	http.HandleFunc("/", serve)
	http.HandleFunc("/ws", socket)
	http.ListenAndServe(*h, nil)
}

type config struct {
	Host string
}

func parse() {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	switch {
	case *d == "" || *d == "." || *d == "./":
	case strings.HasPrefix(*d, "./"):
		dir += (*d)[1:]
	}
	d = &dir

	t := template.Must(template.New("script").Parse(_script))
	b := new(bytes.Buffer)
	t.Execute(b, &config{Host: *h})
	script = b.String()
}

func watch() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("Couldn't add watcher to directory")
		return
	}
	watcher.Add(*d) //TODO: Add directories recursively

	for {
		select {
		case e := <-watcher.Events:
			switch {
			case e.Op&fsnotify.Write == fsnotify.Write:
				signals <- &e.Name
			}
		case err := <-watcher.Errors:
			e := err.Error()
			errors <- &e
		}
	}
}

func alert() {
	for {
		select {
		case <-signals:
			for conn := range connbuf {
				if err := conn.WriteMessage(websocket.TextMessage, []byte("Reload")); err != nil {
					delete(connbuf, conn)
				}
			}
		}
	}
}

type dto struct {
	Script string
}

func serve(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.HasSuffix(r.URL.Path, ".html") || strings.HasSuffix(r.URL.Path, "/"):
		filename := r.URL.Path
		if strings.HasSuffix(filename, "/") {
			filename += "/index.html"
		}
		f, err := file(path.Join(*d, filename))
		if err != nil {
			http.ServeFile(w, r, path.Join(*d, r.URL.Path))
			return
		}
		t, err := template.New("file").Parse(*f)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		b := new(bytes.Buffer)
		t.Execute(b, &dto{Script: script})
		w.Write(b.Bytes())
	default:
		http.ServeFile(w, r, path.Join(*d, r.URL.Path))
	}
}

var upgrader = &websocket.Upgrader{}

func socket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	connbuf[conn] = true
}

func file(n string) (*string, error) {
	var v *string
	f, err := os.Open(n)
	if err != nil {
		return v, err
	}
	b := new(bytes.Buffer)
	_, err = b.ReadFrom(f)
	f.Close()
	if err != nil {
		return v, err
	}
	s := b.String()
	v = &s
	return v, nil
}

const _script = `
<script>
(function(){
  const host = 'ws://{{.Host}}';
	const ws = new WebSocket(host+'/ws');
	ws.onmessage = function(e) {
		switch(e.data) {
		case "Reload": window.location = window.location; break;
		}
	}
})()
</script>
`
