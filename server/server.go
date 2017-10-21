package main

import (
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
)

var tmpl = `<html>
<head>
    <title>Hello World!</title>
</head>
<body>
	<a href=/logout>Logout</a>
    {{ . }}
</body>
</html>
`

func handler(w http.ResponseWriter, r *http.Request) {
	t := template.New("main") //name of the template is main
	t, _ = t.Parse(tmpl)      // parsing of template string
	t.Execute(w, "Hello World!")
}

func handlerFile(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("../Funds-Tracker/index.html")
	t.Execute(w, "Asit")
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/template", handler)
	r.HandleFunc("/filestemplate", handlerFile)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("../Funds-Tracker/")))
	http.ListenAndServe(":8080", r)
}
