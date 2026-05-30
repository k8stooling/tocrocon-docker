package main

import (
	"bytes"
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

func GetTokensAndIdentity(c AuthorizationConfig) (t Tokens, identity *IdentityInfo, err error) {

	//First: get tokens
	t, err = GetTokens(c)
	if err != nil {
		return t, nil, err
	}

	//Then: resolve identity from access token
	identity, err = ResolveIdentityFromAccessToken(t.AccessToken)
	if err != nil {
		return t, nil, errors.Wrap(err, "failed to resolve identity")
	}

	if debugmode == "true" {
		fmt.Println("Groups:", identity.Groups)
		fmt.Println("UPN:", identity.UPN)
	}

	return t, identity, nil
}

func callClaimSourceEndpoint(endpoint, accessToken string) ([]string, error) {
	if debugmode == "true" {
		fmt.Println("endpoint:", endpoint)
		fmt.Println("accessToken:", accessToken)
	}

	reqBody := getMemberObjectsRequest{SecurityEnabledOnly: false}
	b, err := json.Marshal(reqBody)
	if debugmode == "true" {
		fmt.Println("Request-Body:", b)
	}
	if err != nil {
		return nil, errors.Wrap(err, "marshal getMemberObjects request")
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(b))
	if err != nil {
		return nil, errors.Wrap(err, "build claim_sources request")
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	if debugmode == "true" {
		fmt.Println("SendingGroup Objects Request Body:", req.Body)
		fmt.Println("SendingGroup Objects Request Header:", req.Header)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "call claim_sources endpoint")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if debugmode == "true" {
		fmt.Println("Claim Request Response Body:", body)
	}
	if err != nil {
		return nil, errors.Wrap(err, "read claim_sources response body")
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.Errorf("claim_sources endpoint returned %d: %s", resp.StatusCode, string(body))
	}

	var parsed getMemberObjectsResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, errors.Wrap(err, "unmarshal claim_sources response")
	}
	if debugmode == "true" {
		fmt.Println("Parsed Body response:", parsed.Value)
	}
	return parsed.Value, nil
}
