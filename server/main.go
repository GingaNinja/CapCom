package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/dghubble/oauth1"
)

const sessionName = "capcomsession"

// TOTALLY TEMP TO SAVE GETTING ACCOUNT NUMBER
// MUST LOG IN AS cap-com / kingdomcodelondon
const accountNumber = 20171020

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

var store = sessions.NewCookieStore([]byte("sdfuyisadhgjfbshjgdfatasdfguyhsdfb"))

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", rootHandler)
	r.HandleFunc("/authredirect", redirectHandler)
	r.HandleFunc("/logout", logoutHandler)
	r.HandleFunc("/getprivateaccounts", getPrivateAccountsHandler)
	r.HandleFunc("/api/getnexttransaction", apiGetNextTransactionHandler)
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
}

type Transaction struct {
	ID          string
	Amount      float64
	Description string
	Date        time.Time
}

type Value struct {
	Amount float64 `json:",string"`
}

type Details struct {
	Value       Value
	Description string
	Completed   time.Time
}

type ApiTransaction struct {
	ID      string
	Details Details
}

type shortTransaction struct {
	ID string
}

type transactionList struct {
	Transactions []shortTransaction
}

func getTransactionsForAccount(client *http.Client, bankID string, accountID string) ([]string, error) {
	path := fmt.Sprintf("https://apisandbox.openbankproject.com/obp/v1.2.1/banks/%s/accounts/%s/owner/transactions", bankID, accountID)
	fmt.Printf("Aboutn to call %s\n", path)
	resp, err := client.Get(path)
	if err != nil {
		return nil, fmt.Errorf("issue getting request: %v", err)
	}
	defer resp.Body.Close()
	byt, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body: %v", err)
	}

	var transactions transactionList

	if err := json.Unmarshal(byt, &transactions); err != nil {
		return nil, err
	}

	ids := []string{}
	for _, t := range transactions.Transactions {
		ids = append(ids, t.ID)
	}
	return ids, nil
}

func apiGetNextTransactionHandler(w http.ResponseWriter, r *http.Request) {
	if !isAuthd(r) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	session, _ := store.Get(r, sessionName)
	token := session.Values["token"].(string)
	tokenSecret := session.Values["tokenSecret"].(string)
	tokenObject := oauth1.NewToken(token, tokenSecret)
	httpClient := config.Client(oauth1.NoContext, tokenObject)

	transactions, err := getTransactionsForAccount(httpClient, "rbs", "20171020")
	if err != nil {
		panic(err)
	}

	path := fmt.Sprintf("https://apisandbox.openbankproject.com/obp/v1.2.1/banks/rbs/accounts/20171020/owner/transactions/%s/transaction", transactions[0])
	resp, _ := httpClient.Get(path)
	defer resp.Body.Close()
	byt, _ := ioutil.ReadAll(resp.Body)

	var apiTransaction ApiTransaction

	if err := json.Unmarshal(byt, &apiTransaction); err != nil {
		panic(err)
	}

	transaction := Transaction{
		ID:          apiTransaction.ID,
		Amount:      apiTransaction.Details.Value.Amount,
		Description: apiTransaction.Details.Description,
		Date:        apiTransaction.Details.Completed,
	}

	b, err := json.Marshal(transaction)
	if err != nil {
		panic(err)
	}
	w.Write(b)
	w.Header().Set("Content-type", "application/json")
}
