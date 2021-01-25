package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"text/template"

	"github.com/gongqi-zhen/goblueprints/chapter1/trace"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/facebook"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/gomniauth/providers/google"
	"github.com/stretchr/objx"
)

// templ represents a single template
type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

// ServeHTTP handles the HTTP request.
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})

	data := map[string]interface{}{
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}

	t.templ.Execute(w, data)
}

var host = flag.String("host", ":8080", "The host of the application.")

func main() {

	flag.Parse() // parse the flags

	// setup gomniauth
	gomniauth.SetSecurityKey("signature.RandomKey(64)")
	gomniauth.WithProviders(
		github.New("d305f46119031a731609", "2ac3a6d3b5dde276d19db4e73850ac045c68f2b5", "http://localhost:8080/auth/callback/github"),
		google.New("139135232256-fgd745hqk56oo1erc4ik9pp1ua78iabi.apps.googleusercontent.com", "UuSv7o6oX8gMNr8f3tKSwxUW", "http://localhost:8080/auth/callback/google"),
		facebook.New("1585847938283339", "69ea8a98d6e5624d47c937e738b2c802", "http://localhost:8080/auth/callback/facebook"),
	)

	r := newRoom()
	r.tracer = trace.New(os.Stdout)

	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}))
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.HandleFunc("/auth/", loginHandler)
	http.Handle("/room", r)

	// get the room going
	go r.run()

	// start the web server
	log.Println("Starting web server on", *host)
	if err := http.ListenAndServe(*host, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}

}
