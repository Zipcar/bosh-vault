package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cloudfoundry-community/bosh-vault/secret"
)

const NoOverwriteMode = "no-overwrite"

// @see: https://golang.org/pkg/encoding/json/#RawMessage
// Keeping generic parameters as RawMessages allows us to delay unmarhsaling the largest part of this struct until we've
// determined what type of credential we're dealing with. This allows us to efficiently determine credential type
// without needing to use expensive reflection operations.
type GenericCredentialGenerationRequest struct {
	Name       string          `json:"name"`
	Type       string          `json:"type"`
	Parameters json.RawMessage `json:"parameters,omitempty"`
	Mode       string          `json:"mode"`
}

type GenericCredentialSetRequest struct {
	Name  string          `json:"name"`
	Type  string          `json:"type"`
	Value json.RawMessage `json:"value,omitempty"`
}

type CredentialSetRequest struct {
	Name   string
	Type   string
	Record CredentialRecordInterface
}

type CredentialResponse interface{}

type CredentialRecordInterface interface {
	Store(secretStore secret.Store, name string) (CredentialResponse, error)
}

type CredentialGenerationRequest interface {
	Generate(secretStore secret.Store) (CredentialRecordInterface, error)
	Validate() bool
	CredentialType() string
	CredentialName() string
}

func ParseCredentialSetRequest(requestBody []byte) (CredentialSetRequest, error) {
	var g GenericCredentialSetRequest

	err := json.Unmarshal(requestBody, &g)
	if err != nil {
		return CredentialSetRequest{}, errors.New(fmt.Sprintf("error unmarshaling json request: %s", err.Error()))
	}

	var record CredentialRecordInterface

	switch g.Type {
	case CertificateType:
		record = &CertificateRecord{}
	case PasswordType:
		// PasswordRecords are just fancy strings and can't be initialized like structs so we gotta do this
		var passRecord PasswordRecord
		record = &passRecord
	case SshKeypairType:
		record = &SshKeypairRecord{}
	case RsaKeypairType:
		record = &RsaKeypairRecord{}
	default:
		return CredentialSetRequest{}, errors.New(fmt.Sprintf("credential set request type: %s not supported! Must be one of: %s, %s, %s, %s", g.Type, CertificateType, PasswordType, SshKeypairType, RsaKeypairType))
	}

	err = json.Unmarshal(g.Value, &record)

	return CredentialSetRequest{
		Name:   g.Name,
		Type:   g.Type,
		Record: record,
	}, err
}

func ParseCredentialGenerationRequest(requestBody []byte) (req CredentialGenerationRequest, noOverwrite bool, err error) {
	var g GenericCredentialGenerationRequest
	err = json.Unmarshal([]byte(requestBody), &g)
	if err != nil {
		return nil, false, errors.New(fmt.Sprintf("error unmarshaling json request: %s", err.Error()))
	}

	noOverwrite = g.Mode == NoOverwriteMode

	switch g.Type {
	case CertificateType:
		req = &CertificateRequest{}
	case PasswordType:
		req = &PasswordRequest{}
	case SshKeypairType:
		req = &SshKeypairRequest{}
	case RsaKeypairType:
		req = &RsaKeypairRequest{}
	default:
		return nil, noOverwrite, errors.New(fmt.Sprintf("credential request type: %s not supported! Must be one of: %s, %s, %s, %s", g.Type, CertificateType, PasswordType, SshKeypairType, RsaKeypairType))
	}

	err = json.Unmarshal(requestBody, &req)
	return req, noOverwrite, err
}
