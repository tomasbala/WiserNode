package sip

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type DigestAuth struct {
	Username  string
	Realm     string
	Nonce     string
	URI       string
	Response  string
	Algorithm string
	QOP       string
	NC        string
	CNonce    string
}

func GenerateNonce() string {
	data := fmt.Sprintf("%d%d", time.Now().UnixNano(), time.Now().Unix())
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

func CalculateResponse(username, realm, password, method, uri, nonce, qop, nc, cnonce string) string {
	ha1 := md5.Sum([]byte(fmt.Sprintf("%s:%s:%s", username, realm, password)))
	ha1Str := hex.EncodeToString(ha1[:])

	ha2 := md5.Sum([]byte(fmt.Sprintf("%s:%s", method, uri)))
	ha2Str := hex.EncodeToString(ha2[:])

	var response string
	if qop != "" {
		response = fmt.Sprintf("%s:%s:%s:%s:%s:%s", ha1Str, nonce, nc, cnonce, qop, ha2Str)
	} else {
		response = fmt.Sprintf("%s:%s:%s", ha1Str, nonce, ha2Str)
	}

	respHash := md5.Sum([]byte(response))
	return hex.EncodeToString(respHash[:])
}

func ParseAuthHeader(header string) *DigestAuth {
	auth := &DigestAuth{
		Algorithm: "MD5",
	}

	header = strings.TrimPrefix(header, "Digest ")
	header = strings.TrimSpace(header)

	parts := strings.Split(header, ",")
	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}

		key := strings.TrimSpace(kv[0])
		value := strings.Trim(kv[1], `"`)

		switch strings.ToLower(key) {
		case "username":
			auth.Username = value
		case "realm":
			auth.Realm = value
		case "nonce":
			auth.Nonce = value
		case "uri":
			auth.URI = value
		case "response":
			auth.Response = value
		case "algorithm":
			auth.Algorithm = value
		case "qop":
			auth.QOP = value
		case "nc":
			auth.NC = value
		case "cnonce":
			auth.CNonce = value
		}
	}

	return auth
}

func BuildAuthHeader(username, realm, password, method, uri, nonce string) string {
	response := CalculateResponse(username, realm, password, method, uri, nonce, "", "", "")

	return fmt.Sprintf(
		`Digest username="%s", realm="%s", nonce="%s", uri="%s", response="%s", algorithm=MD5`,
		username, realm, nonce, uri, response,
	)
}

func VerifyDigestAuth(auth *DigestAuth, password, method string) bool {
	expected := CalculateResponse(
		auth.Username, auth.Realm, password,
		method, auth.URI, auth.Nonce,
		auth.QOP, auth.NC, auth.CNonce,
	)

	return strings.EqualFold(expected, auth.Response)
}

func BuildWWWAuthenticate(realm, nonce string) string {
	return fmt.Sprintf(
		`Digest realm="%s", nonce="%s", algorithm=MD5`,
		realm, nonce,
	)
}
