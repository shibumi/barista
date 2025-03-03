// Copyright 2018 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package httpclient provides a testable wrapper around an existing *http.Client.
package httpclient

import (
	"net/http"
	"net/url"

	l "github.com/shibumi/barista/logging"

	"golang.org/x/oauth2"
)

type rewritingTransport struct {
	newURL    *url.URL
	transport http.RoundTripper
}

func (r rewritingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	newReq := &http.Request{}
	*newReq = *req
	newReq.URL = &url.URL{}
	*newReq.URL = *req.URL
	newReq.URL.Scheme = r.newURL.Scheme
	newReq.URL.Host = r.newURL.Host
	return r.transport.RoundTrip(newReq)
}

// Wrap redirects all calls from the original *http.Client to the given host.
// Typical usage would be httpclient.Wrap(client, server.URL), where server
// is a httptest.Server or equivalent.
func Wrap(client *http.Client, newURL string) {
	u, _ := url.Parse(newURL)
	client.Transport = rewritingTransport{
		newURL:    u,
		transport: client.Transport,
	}
}

// FreezeOauthToken sets the client's token source to a static token source that
// always provides the given access token.
func FreezeOauthToken(client *http.Client, accessToken string) {
	if t, ok := client.Transport.(*oauth2.Transport); ok {
		t.Source = oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: accessToken})
	} else {
		l.Log("Client %v does not use an oauth transport (%T)",
			client, client.Transport)
	}
}
