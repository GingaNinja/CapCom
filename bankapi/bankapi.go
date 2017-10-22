package bankapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const apiBaseURL = "https://apisandbox.openbankproject.com"
const apiNameAndVer = "obp/v1.2.1"

// BankAPI allows communication with a bank
type BankAPI struct {
	client *http.Client
}

type value struct {
	Amount float64 `json:",string"`
}

type details struct {
	Value       value
	Description string
	Completed   time.Time
}

type transaction struct {
	ID      string
	Details details
}

type shortTransaction struct {
	ID string
}

type transactionList struct {
	Transactions []shortTransaction
}

type Transaction struct {
	ID          string
	Amount      float64
	Description string
	Date        time.Time
}

// NewBankAPI is the constructor for the BankAPI
func NewBankAPI(client *http.Client) BankAPI {
	return BankAPI{client: client}
}

func (b BankAPI) GetTransactionsForAccount(bankID string, accountID string) ([]string, error) {
	path := getFullApiUrl(fmt.Sprintf("/banks/%s/accounts/%s/owner/transactions", bankID, accountID))
	resp, err := b.client.Get(path)
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

func (b BankAPI) GetTransactionFromID(bankName, accountNumber, transactionID string) (tran Transaction, err error) {
	path := getFullApiUrl(fmt.Sprintf("/banks/%s/accounts/%s/owner/transactions/%s/transaction", bankName, accountNumber, transactionID))
	resp, err := b.client.Get(path)
	if err != nil {
		return tran, fmt.Errorf("error requesting transaction information: %v", err)
	}
	defer resp.Body.Close()
	byt, _ := ioutil.ReadAll(resp.Body)

	var apiTransaction transaction

	if err := json.Unmarshal(byt, &apiTransaction); err != nil {
		return tran, fmt.Errorf("error parsing json: %v", err)
	}

	tran = Transaction{
		ID:          apiTransaction.ID,
		Amount:      apiTransaction.Details.Value.Amount,
		Description: apiTransaction.Details.Description,
		Date:        apiTransaction.Details.Completed,
	}
	return
}

func (b BankAPI) GetPrivateAccounts() (jsonBytes []byte, err error) {
	path := getFullApiUrl("/accounts/private")
	resp, err := b.client.Get(path)
	if err != nil {
		return jsonBytes, fmt.Errorf("error getting accounts: %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return jsonBytes, fmt.Errorf("error reading response body: %v", err)
	}
	var dat map[string]interface{}

	if err := json.Unmarshal(body, &dat); err != nil {
		return jsonBytes, fmt.Errorf("error parsing json: %v", err)
	}
	jsonBytes, err = json.MarshalIndent(dat, "", "  ")
	if err != nil {
		return
	}

	return
}

func getFullApiUrl(resource string) string {
	return apiBaseURL + "/" + apiNameAndVer + resource
}
