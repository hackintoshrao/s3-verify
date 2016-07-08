/*
 * Minio S3Verify Library for Amazon S3 Compatible Cloud Storage (C) 2016 Minio, Inc.
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
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/minio/s3verify/signv4"
)

var GetObjectIfNoneMatchReq = &http.Request{
	Header: map[string][]string{
		// Set the Content SHA with empty body for GET requests because nothing is being uploaded.
		"X-Amz-Content-Sha256": {hex.EncodeToString(signv4.Sum256([]byte{}))},
		"If-None-Match":        {""}, // To be filled by the request.
	},
	Body:   nil, // There is no body for GET requests
	Method: "GET",
}

// NewGetObjectIfNoneMatchReq - Create a new HTTP request to perform.
func NewGetObjectIfNoneMatchReq(config ServerConfig, bucketName, objectName, ETag string) (*http.Request, error) {
	targetURL, err := makeTargetURL(config.Endpoint, bucketName, objectName, config.Region)
	if err != nil {
		return nil, err
	}
	GetObjectIfNoneMatchReq.Header.Set("If-None-Match", ETag)
	// Add the URL and sign
	GetObjectIfNoneMatchReq.URL = targetURL
	GetObjectIfNoneMatchReq = signv4.SignV4(*GetObjectIfNoneMatchReq, config.Access, config.Secret, config.Region)
	return GetObjectIfNoneMatchReq, nil
}

// GetObjectIfNoneMatchVerify - Verify that the response matches with what is expected.
func GetObjectIfNoneMatchVerify(res *http.Response, objectBody []byte, expectedStatus string) error {
	if err := VerifyHeaderGetObjectIfNoneMatch(res); err != nil {
		return err
	}
	if err := VerifyStatusGetObjectIfNoneMatch(res, expectedStatus); err != nil {
		return err
	}
	if err := VerifyBodyGetObjectIfNoneMatch(res, objectBody); err != nil {
		return err
	}
	return nil
}

// VerifyHeaderGetObjectIfNoneMatch - Verify that the header fields of the response match what is expected.
func VerifyHeaderGetObjectIfNoneMatch(res *http.Response) error {
	if err := verifyStandardHeaders(res); err != nil {
		return err
	}
	return nil
}

// VerifyStatusGetObjectIfNoneMatch - Verify that the response status matches what is expected.
func VerifyStatusGetObjectIfNoneMatch(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus {
		err := fmt.Errorf("Unexpected Response Status Code: wanted %v, got %v", expectedStatus, res.Status)
		return err
	}
	return nil
}

// VerifyBodyGetObjectIfNoneMatch - Verify that the response body matches what is expected.
func VerifyBodyGetObjectIfNoneMatch(res *http.Response, expectedBody []byte) error {
	// The body should be returned in full.
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if !bytes.Equal(body, expectedBody) { // If the request does not go through an empty body is recieved.
		err := fmt.Errorf("Unexpected Body Recieved: wanted %v, got %v", string(expectedBody), string(body))
		return err
	}
	return nil
}

// Test the compatibility of the GetObject API when using the If-None-Match header.
func mainGetObjectIfNoneMatch(config ServerConfig, message string) error {
	// Set up an invalid ETag to test failed requests responses.
	invalidETag := "1234567890"
	bucket := testBuckets[0]
	for _, object := range objects {
		// Test with If-None-Match Header set.
		// Spin scanBar
		scanBar(message)
		// Create new GET object If-None-Match request.
		req, err := NewGetObjectIfNoneMatchReq(config, bucket.Name, object.Key, object.ETag)
		if err != nil {
			return err
		}
		// Spin scanBar
		scanBar(message)
		// Execute the request.
		res, err := ExecRequest(req, config.Client)
		if err != nil {
			return err
		}
		// Spin scanBar
		scanBar(message)
		// Verify the response...these checks do not check the headers yet.
		if err := GetObjectIfNoneMatchVerify(res, []byte(""), "304 Not Modified"); err != nil {
			return err
		}
		// Spin scanBar
		scanBar(message)
		// Create a bad GET object If-None-Match request with invalid ETag.
		badReq, err := NewGetObjectIfNoneMatchReq(config, bucket.Name, object.Key, invalidETag)
		if err != nil {
			return err
		}
		// Spin scanBar
		scanBar(message)
		// Execute the request.
		badRes, err := ExecRequest(badReq, config.Client)
		if err != nil {
			return err
		}
		// Spin scanBar
		scanBar(message)
		// Verify the response returns the object since ETag != invalidETag
		if err := GetObjectIfNoneMatchVerify(badRes, object.Body, "200 OK"); err != nil {
			return err
		}
		// Spin scanBar
		scanBar(message)
	}
	return nil
}
