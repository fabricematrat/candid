// Copyright 2015 Canonical Ltd.

// Package idputil contains utility routines common to many identity
// providers.
package idputil

import (
	"net/http"
	"net/url"
	"time"

	"github.com/juju/httprequest"
	"github.com/juju/idmclient/params"
	"github.com/juju/loggo"
	"golang.org/x/net/context"
	"gopkg.in/errgo.v1"

	"github.com/CanonicalLtd/blues-identity/idp"
)

var logger = loggo.GetLogger("identity.idp.idputil")

const (
	// identityMacaroonDuration is the length of time for which an
	// identity macaroon is valid.
	identityMacaroonDuration = 28 * 24 * time.Hour
)

// LoginUser completes a successful login for the specified user. A new
// identity macaroon is generated for the user and an appropriate message
// will be returned for the login request.
func LoginUser(ctx idp.RequestContext, waitid string, w http.ResponseWriter, u *params.User) {
	if !ctx.LoginSuccess(waitid, params.Username(u.Username), time.Now().Add(identityMacaroonDuration)) {
		return
	}
	t := ctx.Template("login")
	if t == nil {
		return
	}
	if err := t.Execute(w, u); err != nil {
		logger.Errorf("error processing login template: %s", err)
	}
}

// GetLoginMethods uses c to perform a request to get the list of
// available login methods from u. The result is unmarshalled into v.
func GetLoginMethods(ctx context.Context, c *httprequest.Client, u *url.URL, v interface{}) error {
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return errgo.Mask(err)
	}
	req.Header.Set("Accept", "application/json")
	if err := c.Do(ctx, req, v); err != nil {
		return errgo.Mask(err)
	}
	return nil
}

// RequestParams creates an httprequest.Params object from the given fields.
func RequestParams(ctx context.Context, w http.ResponseWriter, req *http.Request) httprequest.Params {
	return httprequest.Params{
		Response: w,
		Request:  req,
		Context:  ctx,
	}
}

// WaitID gets the wait ID from the given request using the standard form value.
func WaitID(req *http.Request) string {
	return req.Form.Get("waitid")
}

// URL creates a URL addressed to the fgiven path within the IDP handler
// and adds the given waitid (when specified).
func URL(ctx idp.Context, path, waitid string) string {
	callback := ctx.URL(path)
	if waitid != "" {
		callback += "?waitid=" + waitid
	}
	return callback
}
