package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/getToken", handler)

	log.Println("Tocrocon Server running on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var apidata ApiData
	err := json.NewDecoder(r.Body).Decode(&apidata)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	requestParams := AuthorizationConfig{
		Code:        apidata.Code,
		RedirectURI: apidata.RedirectURI,
		GrantType:   apidata.GrantType,
	}

	_, _identity, err := GetTokensAndIdentity(requestParams)
	if err != nil {
		w.Header().Set("Tocrocon-Version", version)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//Call Broker with Azure OIDC Information
	data := BrokerPayload{
		Sub:    _identity.UPN + ":" + ClientID,
		UPN:    _identity.UPN,
		Groups: _identity.Groups,
	}

	resp, err := CallBroker(data)

	type BrokerResponse struct {
		AccessToken string `json:"id_token"`
		Expiry      int    `json:"expires_in"`
	}

	var result BrokerResponse

	err = json.Unmarshal([]byte(resp), &result)
	if err != nil {
		panic(err)
	}

	/*responseData := Tokens{
		AccessToken:  _token.AccessToken,
		RefreshToken: _token.RefreshToken,
		Expiry:       _token.Expiry,
	}*/

	responseData := Tokens{
		AccessToken: result.AccessToken,
		Expiry:      result.Expiry,
	}

	aJson, err := json.Marshal(responseData)
	if err != nil {
		w.Header().Set("Tocrocon-Version", version)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Tocrocon-Version", version)
	w.WriteHeader(http.StatusOK)
	w.Write(aJson)
}
