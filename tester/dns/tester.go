/*
Copyright (c) Facebook, Inc. and its affiliates.
All rights reserved.

This source code is licensed under the BSD-style license found in the
LICENSE file in the root directory of this source tree.
*/

package dns

import (
	"errors"
	"fmt"
	"time"

	"github.com/miekg/dns"
	"github.com/pinterest/bender"
	protocol "github.com/pinterest/bender/dns"
)

// ExtendedMsg wraps a dns.Msg with expectations.
type ExtendedMsg struct {
	dns.Msg
	Rcode int
}

// Tester is a load tester for DNS.
type Tester struct {
	Target   string
	Timeout  time.Duration
	Protocol string
	client   *dns.Client
}

// ErrInvalidRequest is an error raised when the request is invalid.
var ErrInvalidRequest = errors.New("invalid request")

// ErrInvalidResponse is raised when the response is invalid.
var ErrInvalidResponse = errors.New("invalid response")

// Before is called before the first test.
func (t *Tester) Before(options interface{}) error {
	//nolint:exhaustivestruct
	t.client = &dns.Client{
		ReadTimeout:  t.Timeout,
		DialTimeout:  t.Timeout,
		WriteTimeout: t.Timeout,
		Net:          t.Protocol,
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

func validator(request, response *dns.Msg) error {
	if request.Id != response.Id {
		return fmt.Errorf("%w: %d, want: %d", ErrInvalidResponse, request.Id, response.Id)
	}

	return nil
}

// RequestExecutor returns a request executor.
func (t *Tester) RequestExecutor(options interface{}) (bender.RequestExecutor, error) {
	innerExecutor := protocol.CreateExecutor(t.client, validator, t.Target)

	return func(n int64, request interface{}) (interface{}, error) {
		asExtended, ok := request.(*ExtendedMsg)
		if !ok {
			return nil, fmt.Errorf("%w: invalid type, want: *ExtendedMsg, got: %T", ErrInvalidRequest, request)
		}

		resp, err := innerExecutor(n, &asExtended.Msg)
		if err != nil {
			return resp, err
		}

		asMsg, ok := resp.(*dns.Msg)
		if !ok {
			return nil, fmt.Errorf("%w: invalid type, want: *dns.Msg, got: %T", ErrInvalidResponse, resp)
		}

		if asExtended.Rcode != -1 && asExtended.Rcode != asMsg.Rcode {
			return resp, fmt.Errorf(
				"%w: invalid rcode want: %q, got: %q", ErrInvalidResponse,
				dns.RcodeToString[asExtended.Rcode], dns.RcodeToString[asMsg.Rcode])
		}

		return resp, nil
	}, nil
}
