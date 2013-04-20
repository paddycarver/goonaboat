package main

import (
	"code.google.com/p/goauth2/oauth"
	"encoding/json"
	"flag"
	"github.com/bmizerany/pat"
	"github.com/garyburd/go-websocket/websocket"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

var addr = flag.String("addr", ":8081", "HTTP address to listen on")
var clientID = flag.String("clientid", "", "OAuth client ID from Google")
var clientSecret = flag.String("secret", "", "OAuth client secret from Google")
var file = flag.String("file", "index.html", "The path to the HTML file containing your slides")
var temp *template.Template

type TemplateData struct {
	Controller bool
	Session    string
	Host       string
	Auth       string
}

type positions struct {
	Indexh int `json:"indexh"`
	Indexv int `json:"indexv"`
	Fragment int `json:"fragment"`
}

type Account struct {
	Email string
	Error *struct {
		StatusCode int
		Message    string
	}
}

var authCache = make(map[string]string)
var authExpirations = make(map[string]time.Time)

func serve(w http.ResponseWriter, r *http.Request) {
	session := r.URL.Query().Get(":session")
	if r.Method == "GET" && r.URL.Query().Get("ws") != "" {
		ws, err := websocket.Upgrade(w, r.Header, nil, 1024, 1024)
		if _, ok := err.(websocket.HandshakeError); ok {
			http.Error(w, "Not a websocket handshake", 400)
			return
		} else if err != nil {
			log.Println(err)
			return
		}
		c := &connection{send: make(chan []byte, 256), ws: ws}
		if _, ok := sessionMap[session]; !ok {
			sessionMap[session] = &hub{
				broadcast:   make(chan []byte),
				register:    make(chan *connection),
				unregister:  make(chan *connection),
				connections: make(map[*connection]bool),
			}
			go sessionMap[session].run()
		}
		sessionMap[session].register <- c
		c.writePump()
		return
	}
	url := "http://" + r.Host + "/"
	config := &oauth.Config{
		ClientId:     *clientID,
		ClientSecret: *clientSecret,
		Scope:        "https://www.googleapis.com/auth/userinfo.profile https://www.googleapis.com/auth/userinfo.email",
		AuthURL:      "https://accounts.google.com/o/oauth2/auth",
		TokenURL:     "https://accounts.google.com/o/oauth2/token",
		RedirectURL:  url,
	}
	if r.URL.Query().Get("authenticate") != "" && session == "" {
		http.Redirect(w, r, config.AuthCodeURL(""), http.StatusFound)
		return
	}
	auth := ""
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "OAuth" {
			http.Error(w, "Authorization header not formatted correctly.", 401)
			return
		}
		auth = parts[1]
	}
	if r.URL.Query().Get("code") != "" && session == "" {
		t := &oauth.Transport{Config: config}
		token, err := t.Exchange(r.URL.Query().Get("code"))
		if err != nil {
			log.Println(err)
			http.Error(w, "OAuth error", 401)
			return
		}
		auth = token.AccessToken
	}
	if r.URL.Query().Get("auth") != "" && session != "" {
		auth = r.URL.Query().Get("auth")
	}
	user := ""
	if auth != "" {
		if _, ok := authCache[auth]; ok {
			if _, ok := authExpirations[auth]; !ok {
				delete(authCache, auth)
			} else {
				if authExpirations[auth].Before(time.Now()) {
					delete(authCache, auth)
					delete(authExpirations, auth)
				} else {
					user = authCache[auth]
				}
			}
		}
	}
	if auth != "" && user == "" {
		t := &oauth.Transport{
			Config: config,
			Token: &oauth.Token{
				AccessToken: auth,
			},
		}
		req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v1/userinfo", nil)
		if err != nil {
			log.Println(err)
			http.Error(w, "Error verifying OAuth authorization.", 401)
			return
		}
		resp, err := t.RoundTrip(req)
		if err != nil {
			log.Println(err)
			http.Error(w, "Error verifying OAuth authorization.", 401)
			return
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			http.Error(w, "Error verifying OAuth authorization.", 401)
			return
		}
		var account Account
		err = json.Unmarshal(body, &account)
		if err != nil {
			log.Println(err)
			http.Error(w, "Error verifying OAuth authorization.", 401)
			return
		}
		if account.Error != nil {
			log.Println("Google error.")
			log.Println(account.Error.Message)
			http.Error(w, account.Error.Message, account.Error.StatusCode)
			return
		}
		user = account.Email
	}
	if user != "" && auth != "" {
		authCache[auth] = user
		authExpirations[auth] = time.Now()
	}
	log.Println(auth)
	log.Println(user)
	log.Println(session)
	if user != "" && session == "" {
		http.Redirect(w, r, url+user+"?auth="+auth, http.StatusFound)
	}
	if r.Method == "POST" {
		if user == session && user != "" {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Println(err)
				http.Error(w, "Error reading request body.", 400)
				return
			}
			var slidepos positions
			err = json.Unmarshal(body, &slidepos)
			if err != nil {
				log.Println(err)
				http.Error(w, "Error parsing JSON.", 400)
				return
			}
			bytes, err := json.Marshal(slidepos)
			if err != nil {
				log.Println(err)
				http.Error(w, "Error marshalling response.", 500)
				return
			}
			if _, ok := sessionMap[session]; ok {
				sessionMap[session].broadcast <- bytes
			}
			w.WriteHeader(200)
			w.Write([]byte("success"))
		} else {
			http.Error(w, "Invalid auth.", 401)
			return
		}
	} else {
		err := temp.Execute(w, TemplateData{Controller: user == session, Session: session, Host: r.Host, Auth: auth})
		if err != nil {
			log.Println(err)
		}
	}
}

const (
	writeWait      = 10 * time.Second
	readWait       = 10 * time.Second
	pingPeriod     = (readWait * 9) / 10
	maxMessageSize = 512
)

type connection struct {
	ws   *websocket.Conn
	send chan []byte
}

func (c *connection) write(opCode int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(opCode, payload)
}

func (c *connection) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.write(websocket.OpClose, []byte{})
				return
			}
			if err := c.write(websocket.OpText, message); err != nil {
				log.Println(err)
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.OpPing, []byte{}); err != nil {
				return
			}
		}
	}
}

type hub struct {
	connections map[*connection]bool
	broadcast   chan []byte
	register    chan *connection
	unregister  chan *connection
}

var sessionMap = make(map[string]*hub)

func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			h.connections[c] = true
		case c := <-h.unregister:
			delete(h.connections, c)
			close(c.send)
		case m := <-h.broadcast:
			for c := range h.connections {
				select {
				case c.send <- m:
				default:
					close(c.send)
					delete(h.connections, c)
				}
			}
		}
	}
}

func main() {
	flag.Parse()
	temp = template.Must(template.ParseFiles(*file))

	router := pat.New()

	router.Get("/", http.HandlerFunc(serve))
	router.Get("/:session", http.HandlerFunc(serve))
	router.Post("/:session", http.HandlerFunc(serve))

	http.Handle("/", router)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
