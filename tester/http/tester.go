/*
Copyright (c) Facebook, Inc. and its affiliates.
All rights reserved.

This source code is licensed under the BSD-style license found in the
LICENSE file in the root directory of this source tree.
*/

package http

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pinterest/bender"
	protocol "github.com/pinterest/bender/http"
)

// Tester is a load tester for HTTP.
type Tester struct {
	Timeout   time.Duration
	Validator protocol.ResponseValidator
	client    *http.Client
}

// httpStatusOK is the HTTP correct response status code.
const httpStatusOK = 200

// ErrInvalidResponse is raised when a request returns a code different than 200.
var ErrInvalidResponse = errors.New("invalid response status")

// Before is called before the first test.
func (t *Tester) Before(options interface{}) error {
	t.client = &http.Client{
		Timeout: t.Timeout,
	}

	return nil
}

// After is called after all tests are finished.
func (t *Tester) After(_ interface{}) {}

// BeforeEach is called before every test.
func (t *Tester) BeforeEach(_ interface{}) error {
	return nil
}

// AfterEach is called after every test.
func (t *Tester) AfterEach(_ interface{}) {}

// A default validator checks if response type is 200 OK, reads the whole body
// to force download.
func validator(request interface{}, response *http.Response) error {
	if response.StatusCode != httpStatusOK {
		return fmt.Errorf("%w, want: \"200 OK\", got: \"%s\"", ErrInvalidResponse, response.Status)
	}

	_, err := io.Copy(ioutil.Discard, response.Body)

	//nolint:wrapcheck
	return err
}

// RequestExecutor returns a request executor.
func (t *Tester) RequestExecutor(options interface{}) (bender.RequestExecutor, error) {
	if t.Validator == nil {
		return protocol.CreateExecutor(nil, t.client, validator), nil
	}

	return protocol.CreateExecutor(nil, t.client, t.Validator), nil
}
