package eml

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

// Headers
const (
	headerActualAuth    = "Authorization"
	headerForwardedAuth = "X-Forwarded-Authorization"
	headerUserInfo      = "X-Endpoint-API-UserInfo"
	/* transaction@1.0.0 */
	headerEmlSpecification = "X-Message-Specification"
)

const hmacSha256 = "HMAC_SHA256"

type Request struct {
	*http.Request
	PathParams map[string]string
}

func (r *Request) UnmarshalJson(out interface{}) error {
	return r.UnmarshalJsonAndCopy(out, nil)
}

func (r *Request) UnmarshalJsonAndCopy(out interface{}, writer io.Writer) error {
	reader := r.Body
	if writer != nil {
		reader = ioutil.NopCloser(io.TeeReader(reader, writer))
	}
	if err := json.NewDecoder(reader).Decode(out); err != nil {
		return BadError(ErrorParsingBody, ContextualError(err, "json.Decode"))
	}
	return nil
}

//func (r *Request) SetPathFormat(format string) {
//	r.PathParams = utils.PathParams(r.URL.Path, format)
//}

func (r *Request) GetEmlMessageSpec() (messageType, version string, err error) {
	spec := r.Header.Get(headerEmlSpecification)
	if spec == "" || !strings.Contains(spec, "@") {
		err = BadError(ErrorInvalidHeader.Format(headerEmlSpecification), fmt.Errorf("invalid specification header: %s", spec))
	} else {
		parts := strings.Split(spec, "@")
		messageType, version = parts[0], parts[1]
	}
	return
}

// Read from the passed reader, returning another reader to re-read the body
func (r *Request) CheckHmacSignature(keys []Key, body io.Reader) (io.Reader, error) {
	if len(keys) == 0 {
		return body, fmt.Errorf("no keys provided")
	}
	header := r.Header.Get(headerActualAuth)
	if header == "" {
		return body, fmt.Errorf("no Authorization header")
	}
	parts := strings.Split(header, " ")
	if parts[0] != hmacSha256 {
		return body, fmt.Errorf("invalid HMAC algorithm %s", parts[0])
	}
	if len(parts) < 2 || !strings.Contains(parts[1], ";") {
		return body, fmt.Errorf("invalid Auth header %s", header)
	}
	values := strings.Split(parts[1], ";")
	keyId, messageHex := values[0], values[1]
	var key *Key
	for _, k := range keys {
		if k.Id == keyId {
			key = &k
			break
		}
	}
	if key == nil {
		return body, fmt.Errorf("unknown key ID %s", keyId)
	}
	secretBytes, err := key.SecretBytes()
	if err != nil {
		return body, err
	}
	mac := hmac.New(sha256.New, secretBytes)
	var buf bytes.Buffer
	if _, err := io.Copy(mac, io.TeeReader(body, &buf)); err != nil {
		return &buf, err
	}
	expectedMac := mac.Sum(nil)

	messageMac, err := hex.DecodeString(messageHex)
	if err != nil {
		return &buf, err
	}
	if !hmac.Equal(messageMac, expectedMac) {
		return &buf, fmt.Errorf("hash %s doesn't match expected %s\n", messageHex, hex.EncodeToString(expectedMac))
	}
	return &buf, nil
}

func NewRequest(r *http.Request) (*Request, error) {
	req := &Request{Request: r}
	return req, nil
}
