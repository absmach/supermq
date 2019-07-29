//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package bootstrap

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"net/http"

	"github.com/mainflux/mainflux"
)

// bootstrapRes represent Mainflux Response to the Bootatrap request.
// This is used as a response from ConfigReader and can easily be
// replace with any other response format.
type bootstrapRes struct {
	MFThing    string       `json:"mainflux_id"`
	MFKey      string       `json:"mainflux_key"`
	MFChannels []channelRes `json:"mainflux_channels"`
	ClientCert string       `json:"client_cert,omitempty"`
	ClientKey  string       `json:"client_key,omitempty"`
	CaCert     string       `json:"ca_cert,omitempty"`
	Content    string       `json:"content,omitempty"`
}

type channelRes struct {
	ID       string      `json:"id"`
	Name     string      `json:"name,omitempty"`
	Metadata interface{} `json:"metadata,omitempty"`
}

func (res bootstrapRes) Code() int {
	return http.StatusOK
}

func (res bootstrapRes) Headers() map[string]string {
	return map[string]string{}
}

func (res bootstrapRes) Empty() bool {
	return false
}

type reader struct {
	encKey []byte
}

// NewConfigReader return new reader which is used to generate response
// from the config.
func NewConfigReader(encKey []byte) ConfigReader {
	return reader{encKey: encKey}
}

func (r reader) ReadConfig(cfg Config) (mainflux.Response, error) {
	var channels []channelRes
	for _, ch := range cfg.MFChannels {
		channels = append(channels, channelRes{ID: ch.ID, Name: ch.Name, Metadata: ch.Metadata})
	}

	res := bootstrapRes{
		MFKey:      cfg.MFKey,
		MFThing:    cfg.MFThing,
		MFChannels: channels,
		ClientCert: cfg.ClientCert,
		ClientKey:  cfg.ClientKey,
		CaCert:     cfg.CACert,
		Content:    cfg.Content,
	}

	return res, nil
}

func (r reader) Encrypt(in []byte) ([]byte, error) {
	block, err := aes.NewCipher(r.encKey)
	if err != nil {
		return nil, err
	}
	ciphertext := make([]byte, aes.BlockSize+len(in))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], in)
	return ciphertext, nil
}
