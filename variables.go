package main

import "os"

const baseURL = "https://login.microsoftonline.com/"
const tokenPath = "/oauth2/v2.0/token"
const version = "1.1.0"

var TokenURL = baseURL + os.Getenv("TENANT_ID") + tokenPath
var ClientID = os.Getenv("CLIENT_ID")
var scope = "https://graph.microsoft.com/.default offline_access"

var debugmode = os.Getenv("DEBUGMODE")

type Tokens struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	Expiry       int      `json:"expires_in"`
	Groups       []string `json:"groups,omitempty"`
}
type jwtClaims struct {
	UPN string `json:"upn"`
}

type AuthorizationConfig struct {
	Code        string
	RedirectURI string
	GrantType   string
}

type ApiData struct {
	Code        string `json:"code"`
	RedirectURI string `json:"redirect_uri"`
	GrantType   string `json:"grant_type"`
}

type BrokerPayload struct {
	Sub    string   `json:"sub"`
	UPN    string   `json:"upn"`
	Groups []string `json:"groups"`
}

type Group struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
}

type GraphResponse struct {
	Value []Group `json:"value"`
}
