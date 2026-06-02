package main

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
)

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
