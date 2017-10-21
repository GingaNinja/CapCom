package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"github.com/dghubble/oauth1"
)

var config = oauth1.Config{
	ConsumerKey:    "s4s5dt41li1av0focz2fffuknyrzjekjf1wcecc4",
	ConsumerSecret: "mifcleu5gvmdol5jhwxdsygxqxzblwui0yts5sln",
	CallbackURL:    "http://localhost:8080/authredirect",
	Endpoint: oauth1.Endpoint{
		RequestTokenURL: "https://apisandbox.openbankproject.com/oauth/initiate",
		AuthorizeURL:    "https://apisandbox.openbankproject.com/oauth/authorize",
		AccessTokenURL:  "https://apisandbox.openbankproject.com/oauth/token",
	},
}

var requestSecret string

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", rootHandler)
	r.HandleFunc("/authredirect", redirectHandler)
	http.ListenAndServe(":8080", r)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	requestToken, requestSecret1, err := config.RequestToken()
	if err != nil {
		fmt.Printf("error getting token: %v\n", err)
	}
	requestSecret = requestSecret1

	fmt.Printf("RequestToken: %v, RequestSecret: %v\n", requestToken, requestSecret)
	authorizationURL, err := config.AuthorizationURL(requestToken)
	if err != nil {
		fmt.Printf("error getting auth url: %v\n", err)
	}
	fmt.Printf("auth url: %s\n", authorizationURL)

	http.Redirect(w, r, authorizationURL.String(), http.StatusFound)
	//w.Write([]byte("yoyo"))
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	requestToken, verifier, err := oauth1.ParseAuthorizationCallback(r)
	if err != nil {
		w.Write([]byte(fmt.Sprintf("Couldn't get the auth info from callback: %v\n", err)))
	}
	accessToken, accessSecret, err := config.AccessToken(requestToken, requestSecret, verifier)
	// handle error
	if err != nil {
		w.Write([]byte(fmt.Sprintf("Couldn't get the access token: %v\n", err)))
	}
	token := oauth1.NewToken(accessToken, accessSecret)
	w.Write([]byte(fmt.Sprintf("RequestToken: %s\nVerifier: %s\n", requestToken, verifier)))
	w.Write([]byte(fmt.Sprintf("AccessToken: %s\nAccessSecret: %s\n", accessToken, accessSecret)))
	w.Write([]byte(fmt.Sprintf("Token is %s\nSecret is %s\n", token.Token, token.TokenSecret)))

	httpClient := config.Client(oauth1.NoContext, token)

	// example Twitter API request
	path := "https://apisandbox.openbankproject.com/obp/v1.2.1/accounts/private"
	resp, _ := httpClient.Get(path)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var dat map[string]interface{}

	if err := json.Unmarshal(body, &dat); err != nil {
		panic(err)
	}
	b, err := json.MarshalIndent(dat, "", "  ")
	if err != nil {
		panic(err)
	}
	b2 := append(b, '\n')
	os.Stdout.Write(b2)
	os.Stdout.Write(b)
	w.Write(b)
	//w.Write([]byte("redirected yay"))
}
