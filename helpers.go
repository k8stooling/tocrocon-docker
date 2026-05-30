package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// decodeJWTPayload decodes the JWT payload and unmarshals it into out.
// This does NOT validate signature; it only reads claims.
func decodeJWTPayload(jwt string, out any) error {
	parts := strings.Split(jwt, ".")
	if len(parts) < 2 {
		return errors.New("invalid JWT format")
	}

	payloadB64 := parts[1]
	payload, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		// fallback if padding exists for some reason
		payload, err = base64.URLEncoding.DecodeString(payloadB64)
		if err != nil {
			return errors.Wrap(err, "failed to decode JWT payload")
		}
	}

	if err := json.Unmarshal(payload, out); err != nil {
		return errors.Wrap(err, "failed to unmarshal JWT payload")
	}
	return nil
}

// ResolveIdentityFromAccessToken returns group object IDs from the access token.
// If group overage is present, it uses the claim_sources endpoint to fetch them.
func ResolveIdentityFromAccessToken(accessToken string) (*IdentityInfo, error) {
	var c jwtClaims

	if err := decodeJWTPayload(accessToken, &c); err != nil {
		return nil, err
	}

	if debugmode == "true" {
		fmt.Println("jwtClaims Groups:", c.Groups)
		fmt.Println("jwtClaims UPN:", c.UPN)
	}

	identity := &IdentityInfo{
		UPN: c.UPN,
	}

	//Case 1: groups are already in the token
	if len(c.Groups) > 0 {
		if debugmode == "true" {
			fmt.Println("groups are part of the token:", c.Groups)
		}
		identity.Groups = c.Groups
		return identity, nil
	}

	//Case 2: group overage -> _claim_names and _claim_sources
	srcKey, ok := c.ClaimNames["groups"]
	if !ok || srcKey == "" {
		// no groups and no overage pointers
		return identity, nil
	}

	if debugmode == "true" {
		fmt.Println("source-key:", srcKey)
	}

	src, ok := c.ClaimSources[srcKey]

	if debugmode == "true" {
		fmt.Println("claimed-sources:", src)
	}

	if !ok || src.Endpoint == "" {
		if debugmode == "true" {
			fmt.Println("groups overage indicated but _claim_sources endpoint missing")
		}
		return nil, errors.New("groups overage indicated but _claim_sources endpoint missing")
	}

	//Fetch groups from endpoint
	groups, err := callClaimSourceEndpoint(src.Endpoint, accessToken)
	if err != nil {
		return nil, err
	}

	identity.Groups = groups
	return identity, nil
}
