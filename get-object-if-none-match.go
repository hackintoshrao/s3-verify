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

// newGetObjectIfNoneMatchReq - Create a new HTTP request to perform.
func newGetObjectIfNoneMatchReq(config ServerConfig, bucketName, objectName, ETag string) (*http.Request, error) {
	var getObjectIfNoneMatchReq = &http.Request{
		Header: map[string][]string{
			// Set the Content SHA with empty body for GET requests because nothing is being uploaded.
			"X-Amz-Content-Sha256": {hex.EncodeToString(signv4.Sum256([]byte{}))},
		},
		Body:   nil, // There is no body for GET requests
		Method: "GET",
	}
	targetURL, err := makeTargetURL(config.Endpoint, bucketName, objectName, config.Region, nil)
	if err != nil {
		return nil, err
	}
	getObjectIfNoneMatchReq.Header.Set("If-None-Match", ETag)
	// Add the URL and sign
	getObjectIfNoneMatchReq.URL = targetURL
	getObjectIfNoneMatchReq = signv4.SignV4(*getObjectIfNoneMatchReq, config.Access, config.Secret, config.Region)
	return getObjectIfNoneMatchReq, nil
}

// getObjectIfNoneMatchVerify - Verify that the response matches with what is expected.
func getObjectIfNoneMatchVerify(res *http.Response, objectBody []byte, expectedStatus string) error {
	if err := verifyHeaderGetObjectIfNoneMatch(res); err != nil {
		return err
	}
	if err := verifyStatusGetObjectIfNoneMatch(res, expectedStatus); err != nil {
		return err
	}
	if err := verifyBodyGetObjectIfNoneMatch(res, objectBody); err != nil {
		return err
	}
	return nil
}

// verifyHeaderGetObjectIfNoneMatch - Verify that the header fields of the response match what is expected.
func verifyHeaderGetObjectIfNoneMatch(res *http.Response) error {
	if err := verifyStandardHeaders(res); err != nil {
		return err
	}
	return nil
}

// verifyStatusGetObjectIfNoneMatch - Verify that the response status matches what is expected.
func verifyStatusGetObjectIfNoneMatch(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus {
		err := fmt.Errorf("Unexpected Response Status Code: wanted %v, got %v", expectedStatus, res.Status)
		return err
	}
	return nil
}

// verifyBodyGetObjectIfNoneMatch - Verify that the response body matches what is expected.
func verifyBodyGetObjectIfNoneMatch(res *http.Response, expectedBody []byte) error {
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
		req, err := newGetObjectIfNoneMatchReq(config, bucket.Name, object.Key, object.ETag)
		if err != nil {
			return err
		}
		// Spin scanBar
		scanBar(message)
		// Execute the request.
		res, err := execRequest(req, config.Client)
		if err != nil {
			return err
		}
		// Spin scanBar
		scanBar(message)
		// Verify the response...these checks do not check the headers yet.
		if err := getObjectIfNoneMatchVerify(res, []byte(""), "304 Not Modified"); err != nil {
			return err
		}
		// Spin scanBar
		scanBar(message)
		// Create a bad GET object If-None-Match request with invalid ETag.
		badReq, err := newGetObjectIfNoneMatchReq(config, bucket.Name, object.Key, invalidETag)
		if err != nil {
			return err
		}
		// Spin scanBar
		scanBar(message)
		// Execute the request.
		badRes, err := execRequest(badReq, config.Client)
		if err != nil {
			return err
		}
		// Spin scanBar
		scanBar(message)
		// Verify the response returns the object since ETag != invalidETag
		if err := getObjectIfNoneMatchVerify(badRes, object.Body, "200 OK"); err != nil {
			return err
		}
		// Spin scanBar
		scanBar(message)
	}
	return nil
}
