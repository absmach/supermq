// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// Package ui contains the domain concept definitions needed to support
// Mainflux ui adapter service functionality.
package ui

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"

	"github.com/mainflux/mainflux"
	sdk "github.com/mainflux/mainflux/pkg/sdk/go"
)

const (
	templateDir = "ui/web/template"
)

var (
	// ErrUnauthorizedAccess indicates missing or invalid credentials provided
	// when accessing a protected resource.
	ErrUnauthorizedAccess = errors.New("missing or invalid credentials provided")

	// ErrMalformedEntity indicates malformed entity specification (e.g.
	// invalid username or password).
	ErrMalformedEntity = errors.New("malformed entity specification")
)

// Service specifies coap service API.
type Service interface {
	Index(ctx context.Context, token string) ([]byte, error)
	CreateThings(ctx context.Context, token string, things ...sdk.Thing) ([]byte, error)
	ListThings(ctx context.Context, token string) ([]byte, error)
	UpdateThing(ctx context.Context, token string, thing sdk.Thing) ([]byte, error)
	CreateChannels(ctx context.Context, token string, channels ...sdk.Channel) ([]byte, error)
	ListChannels(ctx context.Context, token string) ([]byte, error)
}

var _ Service = (*uiService)(nil)

type uiService struct {
	things, channels mainflux.ThingsServiceClient
	sdk              sdk.SDK
}

// New instantiates the HTTP adapter implementation.
func New(things, channels mainflux.ThingsServiceClient, sdk sdk.SDK) Service {
	return &uiService{
		channels: channels,
		things:   things,
		sdk:      sdk,
	}
}

func (gs *uiService) Index(ctx context.Context, token string) ([]byte, error) {
	tpl, err := template.ParseGlob(templateDir + "/*")
	if err != nil {
		return []byte{}, err
	}

	data := struct {
		NavbarActive string
	}{
		"dashboard",
	}

	var btpl bytes.Buffer
	if err := tpl.ExecuteTemplate(&btpl, "index", data); err != nil {
		println(err.Error())
	}

	return btpl.Bytes(), nil
}

func (gs *uiService) CreateThings(ctx context.Context, token string, things ...sdk.Thing) ([]byte, error) {

	for i := range things {
		fmt.Println(things[i])
		_, err := gs.sdk.CreateThing(things[i], "123")
		if err != nil {
			return []byte{}, err
		}
	}

	return gs.ListThings(ctx, "123")
}

func (gs *uiService) ListThings(ctx context.Context, token string) ([]byte, error) {
	tpl, err := template.ParseGlob(templateDir + "/*")
	if err != nil {
		return []byte{}, err
	}

	thsPage, err := gs.sdk.Things("123", 0, 100, "")
	if err != nil {
		return []byte{}, err
	}
	fmt.Println(thsPage.Things)

	data := struct {
		NavbarActive string
		Things       []sdk.Thing
	}{
		"things",
		thsPage.Things,
	}

	var btpl bytes.Buffer
	if err := tpl.ExecuteTemplate(&btpl, "things", data); err != nil {
		println(err.Error())
	}

	return btpl.Bytes(), nil
}

func (gs *uiService) UpdateThing(ctx context.Context, token string, thing sdk.Thing) ([]byte, error) {
	_, err := template.ParseGlob(templateDir + "/*")
	if err != nil {
		return []byte{}, err
	}

	return gs.ListThings(ctx, "123")
}

func (gs *uiService) CreateChannels(ctx context.Context, token string, channels ...sdk.Channel) ([]byte, error) {
	fmt.Println("assss")
	for i := range channels {
		fmt.Println(channels[i])
		_, err := gs.sdk.CreateChannel(channels[i], "123")
		if err != nil {
			return []byte{}, err
		}
	}

	return gs.ListChannels(ctx, "123")
}

func (gs *uiService) ListChannels(ctx context.Context, token string) ([]byte, error) {
	tpl, err := template.ParseGlob(templateDir + "/*")
	if err != nil {
		return []byte{}, err
	}

	chsPage, err := gs.sdk.Channels("123", 0, 100, "")
	if err != nil {
		return []byte{}, err
	}
	fmt.Println(chsPage.Channels)

	data := struct {
		NavbarActive string
		Channels     []sdk.Channel
	}{
		"channels",
		chsPage.Channels,
	}

	var btpl bytes.Buffer
	if err := tpl.ExecuteTemplate(&btpl, "channels", data); err != nil {
		println(err.Error())
	}

	return btpl.Bytes(), nil
}
