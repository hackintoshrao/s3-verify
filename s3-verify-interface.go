/*
 * Minio Go Library for Amazon S3 Compatible Cloud Storage (C) 2015, 2016 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"net/http"
	"net/url"
)

type S3Verify interface {
	// construct HTTP request.
	MakePlainRequest(endPointStr string) (*http.Request, error)
	// construct the URL for bucket operation.
	MakeURLPath(endPointStr string) (*url.URL, error)
	// interface containing methods for setting body, signature and headers of the request.
	FillHTTPRequest
	// Executes the HTTP request and returns the response.
	ExecRequest(*http.Request) (*http.Response, error)
	// Verifies the response for S3 Compatibility.
	VerifyResponse(*http.Response) error
}

// FillHTTPRequest is an interface with utilities for adding v4-Signature, Headers and setting Body of HTTP request.
// Used for contructing http requests for given object storage operation.
type FillHTTPRequest interface {
	SetHeaders(req *http.Request) *http.Request
	SetBody(req *http.Request) *http.Request
	SignRequest(req *http.Request, accessKeyID, secretAccessKey string) *http.Request
}
