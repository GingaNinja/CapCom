package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ginganinja/capcom/bankapi"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/dghubble/oauth1"
)

const sessionName = "capcomsession"

// TOTALLY TEMP TO SAVE GETTING ACCOUNT NUMBER
// MUST LOG IN AS cap-com / kingdomcodelondon
const accountNumber = "20171020"
const bankName = "rbs"
const apiBaseURL = "https://apisandbox.openbankproject.com"

var config = oauth1.Config{
	ConsumerKey:    "s4s5dt41li1av0focz2fffuknyrzjekjf1wcecc4",
	ConsumerSecret: "mifcleu5gvmdol5jhwxdsygxqxzblwui0yts5sln",
	CallbackURL:    "http://localhost:8080/authredirect",
	Endpoint: oauth1.Endpoint{
		RequestTokenURL: apiBaseURL + "/oauth/initiate",
		AuthorizeURL:    apiBaseURL + "/oauth/authorize",
		AccessTokenURL:  apiBaseURL + "/oauth/token",
	},
}

var requestSecret string

var store = sessions.NewCookieStore([]byte("sdfuyisadhgjfbshjgdfatasdfguyhsdfb"))

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", rootHandler)
	r.HandleFunc("/authredirect", redirectHandler)
	r.HandleFunc("/logout", logoutHandler)
	r.HandleFunc("/getprivateaccounts", getPrivateAccountsHandler)
	r.HandleFunc("/api/getnexttransaction", apiGetNextTransactionHandler)
	r.HandleFunc("/cleardate", clearDateHandler)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("../Funds-Tracker/"))))
	http.ListenAndServe(":8080", r)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if !isAuthd(r) {
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
		return
	}

	http.Redirect(w, r, "/static/", http.StatusFound)
	//w.Write([]byte("Hello! You're logged in"))
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
	// w.Write([]byte(fmt.Sprintf("RequestToken: %s\nVerifier: %s\n", requestToken, verifier)))
	// w.Write([]byte(fmt.Sprintf("AccessToken: %s\nAccessSecret: %s\n", accessToken, accessSecret)))
	// w.Write([]byte(fmt.Sprintf("Token is %s\nSecret is %s\n", token.Token, token.TokenSecret)))

	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty.
	session, _ := store.Get(r, sessionName)
	store.MaxAge(65000)
	// Set some session values.
	values := session.Values
	values["token"] = token.Token
	values["tokenSecret"] = token.TokenSecret
	session.Values = values

	// Save it before we write to the response/return from the handler.
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusFound)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	store.MaxAge(-1)
	w.Write([]byte("logged out <a href='/'>Log back in</a>"))
}

func clearDateHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, sessionName)
	session.Values["lastTransactionDate"] = time.Date(1900, 1, 1, 1, 1, 1, 1, time.Local).Format("Mon Jan 2 15:04:05 -0700 MST 2006")
	session.Save(r, w)
}

func isAuthd(r *http.Request) bool {
	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty.
	session, _ := store.Get(r, sessionName)
	_, ok := session.Values["token"]
	return ok
}

func getPrivateAccountsHandler(w http.ResponseWriter, r *http.Request) {
	if !isAuthd(r) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	session, _ := store.Get(r, sessionName)
	token := session.Values["token"].(string)
	tokenSecret := session.Values["tokenSecret"].(string)
	tokenObject := oauth1.NewToken(token, tokenSecret)
	httpClient := config.Client(oauth1.NoContext, tokenObject)
	bankAPI := bankapi.NewBankAPI(httpClient)

	b, err := bankAPI.GetPrivateAccounts()
	if err != nil {
		panic(err)
	}
	//b2 := append(b, '\n')
	//os.Stdout.Write(b2)
	//os.Stdout.Write(b)
	w.Write(b)
}

func apiGetNextTransactionHandler(w http.ResponseWriter, r *http.Request) {
	if !isAuthd(r) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	session, err := store.Get(r, sessionName)
	if err != nil {
		w.Write([]byte(fmt.Sprintf("error getting session: %v\n", err)))
		return
	}

	var lastTransactionDate time.Time
	stringDate, ok := session.Values["lastTransactionDate"]
	if !ok {
		lastTransactionDate = time.Date(1900, 1, 1, 1, 1, 1, 1, time.Local)
	} else {
		lastTransactionDate, _ = time.Parse("Mon Jan 2 15:04:05 -0700 MST 2006", stringDate.(string))
	}
	fmt.Printf("lastTransactionDate: %v\n", lastTransactionDate)

	token := session.Values["token"].(string)
	tokenSecret := session.Values["tokenSecret"].(string)
	tokenObject := oauth1.NewToken(token, tokenSecret)
	httpClient := config.Client(oauth1.NoContext, tokenObject)
	bankAPI := bankapi.NewBankAPI(httpClient)

	transactionList, err := bankAPI.GetTransactionsForAccount("rbs", "20171020")
	if err != nil {
		panic(err)
	}
	i := len(transactionList) - 1
	for ; i > -1 && !transactionList[i].Date.After(lastTransactionDate); i-- {
		fmt.Printf("date %d: %v\n", i, transactionList[i].Date.Format("Mon Jan 2 15:04:05 -0700 MST 2006"))
	}
	if i > -1 {
		fmt.Printf("Last date: %s\n", transactionList[i].Date.Format("Mon Jan 2 15:04:05 -0700 MST 2006"))
		transaction, err := bankAPI.GetTransactionFromID("rbs", "20171020", transactionList[i].ID)
		if err != nil {
			panic(err)
		}

		b, err := json.Marshal(transaction)
		if err != nil {
			panic(err)
		}

		session.Values["lastTransactionDate"] = transactionList[i].Date.Format("Mon Jan 2 15:04:05 -0700 MST 2006")
		if err := session.Save(r, w); err != nil {
			w.Write([]byte(fmt.Sprintf("error saving session: %v\n", err)))
		}
		w.Write(b)
	} else {
		result := bankapi.Transaction{}
		b, err := json.Marshal(result)
		if err != nil {
			panic(err)
		}
		w.Write(b)
	}

	w.Header().Set("Content-type", "application/json")
}
