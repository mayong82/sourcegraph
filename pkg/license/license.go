// Package license provides license key generation and verification.
//
// License keys are generated and signed using Sourcegraph's private key. Sourcegraph instances must
// be able to verify the license key offline, so all license information (such as the max user
// count) is encoded in the license itself.
//
// Key rotation, license key revocation, etc., are not implemented.
package license

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode"

	"golang.org/x/crypto/ssh"
)

// Info contains information about a license key. In the signed license key that Sourcegraph
// provides to customers, this value is signed but not encrypted. This value is not secret, and
// anyone with a license key can view (but not forge) this information.
//
// NOTE: If you change these fields, you MUST handle backward compatibility. Existing licenses that
// were generated with the old fields must still work until all customers have added the new
// license. Increment (encodedInfo).Version and formatVersion when you make backward-incompatbile
// changes.
type Info struct {
	Plan         string     `json:"p"`   // the plan that this license is valid for
	MaxUserCount *uint      `json:"uc"`  // the maximum number of users that this license is valid for
	Expiry       *time.Time `json:"exp"` // the date when this license expires
}

// IsExpired reports whether the license has expired.
func (l Info) IsExpired() bool {
	return l.Expiry != nil && l.Expiry.Before(time.Now())
}

type encodedInfo struct {
	Version int     `json:"v"` // version number of the license key info format (not Sourcegraph product/build version)
	Nonce   [8]byte `json:"n"` // random nonce so that licenses with identical Info values
	Info
}

const formatVersion = 1 // (encodedInfo).Version value

func (l Info) encode() ([]byte, error) {
	e := encodedInfo{Version: formatVersion, Info: l}
	if _, err := rand.Read(e.Nonce[:8]); err != nil {
		return nil, err
	}
	return json.Marshal(e)
}

func (l *Info) decode(data []byte) error {
	var e encodedInfo
	if err := json.Unmarshal(data, &e); err != nil {
		return err
	}
	if e.Version != formatVersion {
		return fmt.Errorf("license key format is version %d, expected version %d", e.Version, formatVersion)
	}
	*l = e.Info
	return nil
}

type signedKey struct {
	Signature   *ssh.Signature `json:"sig"`
	EncodedInfo []byte         `json:"info"`
}

// GenerateSignedKey generates a new signed license key with the given license information, using
// the private key for the signature.
func GenerateSignedKey(info Info, privateKey ssh.Signer) (string, error) {
	encodedInfo, err := info.encode()
	if err != nil {
		return "", err
	}
	sig, err := privateKey.Sign(rand.Reader, encodedInfo)
	if err != nil {
		return "", err
	}
	signedKeyData, err := json.Marshal(signedKey{Signature: sig, EncodedInfo: encodedInfo})
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(signedKeyData), nil
}

// ParseSignedKey parses and verifies the signed license key. If parsing or verification fails, a
// non-nil error is returned.
func ParseSignedKey(text string, publicKey ssh.PublicKey) (*Info, error) {
	// Ignore whitespace, in case the license key was (e.g.) wrapped in an email message.
	text = strings.Map(func(c rune) rune {
		if unicode.IsSpace(c) {
			return -1 // drop
		}
		return c
	}, text)

	signedKeyData, err := base64.RawURLEncoding.DecodeString(text)
	if err != nil {
		return nil, err
	}
	var signedKey signedKey
	if err := json.Unmarshal(signedKeyData, &signedKey); err != nil {
		return nil, err
	}
	if err := publicKey.Verify(signedKey.EncodedInfo, signedKey.Signature); err != nil {
		return nil, err
	}
	var info Info
	if err := json.Unmarshal(signedKey.EncodedInfo, &info); err != nil {
		return nil, err
	}
	return &info, nil
}
