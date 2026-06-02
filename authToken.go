package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

func GetTokens(c AuthorizationConfig) (t Tokens, err error) {

	if debugmode == "true" {
		fmt.Println("GrantType: ", c.GrantType)
		fmt.Println("Code: ", c.Code)
		fmt.Println("RedirectURI: ", c.RedirectURI)
	}

	var formVals = url.Values{}
	formVals.Set("client_id", ClientID)

	if debugmode == "true" {
		fmt.Println("FormVals: ", formVals.Encode())
		fmt.Println("grant", c.GrantType)
	}

	if c.GrantType == "authorization_code" {
		formVals.Set("client_secret", getSecretData())
		formVals.Set("code", c.Code)
		formVals.Set("grant_type", c.GrantType)
		formVals.Set("redirect_uri", c.RedirectURI)
		formVals.Set("scope", scope)
	} else if c.GrantType == "urn:ietf:params:oauth:grant-type:device_code" {
		formVals.Set("grant_type", c.GrantType)
		formVals.Set("device_code", c.Code)
	} else if c.GrantType == "refresh_token" {
		formVals.Set("client_secret", getSecretData())
		formVals.Set("grant_type", c.GrantType)
		formVals.Set("refresh_token", c.Code)
	}

	response, err := http.PostForm(TokenURL, formVals)
	if err != nil {
		return t, errors.Wrap(err, "error while trying to get tokens")
	}
	defer response.Body.Close()

	if debugmode == "true" {
		fmt.Println("Formvals 2nd: ", formVals.Encode())
		fmt.Println("TokenURL: ", TokenURL)
	}

	if err != nil {
		return t, errors.Wrap(err, "error while trying to get tokens")
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return t, errors.Wrap(err, "error while trying to read token json body")
	}

	if debugmode == "true" {
		fmt.Println("Response-Body: ", body)
	}

	if err != nil {
		return t, errors.Wrap(err, "error while trying to read token json body")
	}

	err = json.Unmarshal(body, &t)
	if err != nil {
		return t, errors.Wrap(err, "error while trying to parse token json body")
	}

	if debugmode == "true" {
		str1 := fmt.Sprintf("%s", body)
		fmt.Println("Tocrocon-Version:", version)
		fmt.Println("Complete Body Response =", str1)
		fmt.Println("AccessToken: ", t.AccessToken)
		fmt.Println("RefreshToken: ", t.RefreshToken)
		fmt.Println("Expiry: ", t.Expiry)
	}

	fmt.Println("GrantType: ", c.GrantType, " processing successfull!")

	return
}

func getGroupNames(token string) ([]string, error) {
	req, err := http.NewRequest("GET", graphURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Handle non-200 responses
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("graph API error: %s - %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result GraphResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	// Extract display names
	var groupNames []string
	for _, group := range result.Value {
		groupNames = append(groupNames, group.DisplayName)
	}

	return groupNames, nil
}
