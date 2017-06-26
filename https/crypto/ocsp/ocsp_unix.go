// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

// fork from golang.org/x/crypto/ocsp
// Package ocsp parses OCSP responses as specified in RFC 2560. OCSP responses
// are signed messages attesting to the validity of a certificate for a small
// period of time. This is used to manage revocation for X.509 certificates.
package ocsp

import (
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"errors"
	"fmt"
	"time"
)

// CreateResponse returns a DER-encoded OCSP response with the specified contents.
// The fields in the response are populated as follows:
//
// The responder cert is used to populate the responder's name field, and the
// certificate itself is provided alongside the OCSP response signature.
//
// The issuer cert is used to puplate the IssuerNameHash and IssuerKeyHash fields.
//
// The template is used to populate the SerialNumber, RevocationStatus, RevokedAt,
// RevocationReason, ThisUpdate, and NextUpdate fields.
//
// If template.IssuerHash is not set, SHA1 will be used.
//
// The ProducedAt date is automatically set to the current date, to the nearest minute.
func CreateResponse(issuer, responderCert *x509.Certificate, template Response, priv crypto.Signer) ([]byte, error) {
	var publicKeyInfo struct {
		Algorithm pkix.AlgorithmIdentifier
		PublicKey asn1.BitString
	}
	if _, err := asn1.Unmarshal(issuer.RawSubjectPublicKeyInfo, &publicKeyInfo); err != nil {
		return nil, err
	}

	if template.IssuerHash == 0 {
		template.IssuerHash = crypto.SHA1
	}
	hashOID := getOIDFromHashAlgorithm(template.IssuerHash)
	if hashOID == nil {
		return nil, errors.New("unsupported issuer hash algorithm")
	}

	if !template.IssuerHash.Available() {
		return nil, fmt.Errorf("issuer hash algorithm %v not linked into binary", template.IssuerHash)
	}
	h := template.IssuerHash.New()
	h.Write(publicKeyInfo.PublicKey.RightAlign())
	issuerKeyHash := h.Sum(nil)

	h.Reset()
	h.Write(issuer.RawSubject)
	issuerNameHash := h.Sum(nil)

	innerResponse := singleResponse{
		CertID: certID{
			HashAlgorithm: pkix.AlgorithmIdentifier{
				Algorithm:  hashOID,
				Parameters: asn1.RawValue{Tag: 5 /* ASN.1 NULL */},
			},
			NameHash:      issuerNameHash,
			IssuerKeyHash: issuerKeyHash,
			SerialNumber:  template.SerialNumber,
		},
		ThisUpdate:       template.ThisUpdate.UTC(),
		NextUpdate:       template.NextUpdate.UTC(),
		SingleExtensions: template.ExtraExtensions,
	}

	switch template.Status {
	case Good:
		innerResponse.Good = true
	case Unknown:
		innerResponse.Unknown = true
	case Revoked:
		innerResponse.Revoked = revokedInfo{
			RevocationTime: template.RevokedAt.UTC(),
			Reason:         asn1.Enumerated(template.RevocationReason),
		}
	}

	rawResponderID := asn1.RawValue{
		Class:      2, // context-specific
		Tag:        1, // Name (explicit tag)
		IsCompound: true,
		Bytes:      responderCert.RawSubject,
	}
	tbsResponseData := responseData{
		Version:        0,
		RawResponderID: rawResponderID,
		ProducedAt:     time.Now().Truncate(time.Minute).UTC(),
		Responses:      []singleResponse{innerResponse},
	}

	tbsResponseDataDER, err := asn1.Marshal(tbsResponseData)
	if err != nil {
		return nil, err
	}

	hashFunc, signatureAlgorithm, err := signingParamsForPublicKey(priv.Public(), template.SignatureAlgorithm)
	if err != nil {
		return nil, err
	}

	responseHash := hashFunc.New()
	responseHash.Write(tbsResponseDataDER)
	signature, err := priv.Sign(rand.Reader, responseHash.Sum(nil), hashFunc)
	if err != nil {
		return nil, err
	}

	response := basicResponse{
		TBSResponseData:    tbsResponseData,
		SignatureAlgorithm: signatureAlgorithm,
		Signature: asn1.BitString{
			Bytes:     signature,
			BitLength: 8 * len(signature),
		},
	}
	if template.Certificate != nil {
		response.Certificates = []asn1.RawValue{
			asn1.RawValue{FullBytes: template.Certificate.Raw},
		}
	}
	responseDER, err := asn1.Marshal(response)
	if err != nil {
		return nil, err
	}

	return asn1.Marshal(responseASN1{
		Status: asn1.Enumerated(Success),
		Response: responseBytes{
			ResponseType: idPKIXOCSPBasic,
			Response:     responseDER,
		},
	})
}
