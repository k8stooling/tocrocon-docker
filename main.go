package main

import (
	"encoding/json"
	"fmt"
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

	if debugmode == "true" {
		fmt.Println("Calling Broker request parameters: ", requestParams)
	}

	_tokens, err := GetTokens(requestParams)
	if err != nil {
		w.Header().Set("Tocrocon-Version", version)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	groupNames, err := getGroupNames(_tokens.AccessToken)
	fmt.Println("Initiate GRAPH-Call")

	var c jwtClaims

	if err := decodeJWTPayload(_tokens.AccessToken, &c); err != nil {
		return
	}

	if debugmode == "true" {
		fmt.Println("jwtClaims UPN:", c.UPN)
	}

	//Call Broker with Azure OIDC Information
	data := BrokerPayload{
		Sub:    c.UPN + ":" + ClientID,
		UPN:    c.UPN,
		Groups: groupNames,
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

	responseData := Tokens{
		AccessToken:  result.AccessToken,
		RefreshToken: _tokens.RefreshToken,
		Expiry:       result.Expiry,
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

	if _, err := w.Write(aJson); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}
