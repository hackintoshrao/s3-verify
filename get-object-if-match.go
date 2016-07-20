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

// newGetObjectIfMatchReq - Create a new HTTP request to perform.
func newGetObjectIfMatchReq(config ServerConfig, bucketName, objectName, ETag string) (*http.Request, error) {
	var getObjectIfMatchReq = &http.Request{
		Header: map[string][]string{
			// Set Content SHA with empty body for GET requests because no data is being uploaded.
			"X-Amz-Content-Sha256": {hex.EncodeToString(signv4.Sum256([]byte{}))},
		},
		Body:   nil, // There is no body for GET requests.
		Method: "GET",
	}
	targetURL, err := makeTargetURL(config.Endpoint, bucketName, objectName, config.Region, nil)
	if err != nil {
		return nil, err
	}
	getObjectIfMatchReq.Header.Set("If-Match", ETag)
	// Fill request URL and sign
	getObjectIfMatchReq.URL = targetURL
	getObjectIfMatchReq = signv4.SignV4(*getObjectIfMatchReq, config.Access, config.Secret, config.Region)
	return getObjectIfMatchReq, nil
}

// getObjectIfMatchVerify - Verify that the response matches what is expected.
func getObjectIfMatchVerify(res *http.Response, objectBody []byte, expectedStatus string, shouldFail bool) error {
	if err := verifyHeaderGetObjectIfMatch(res); err != nil {
		return err
	}
	if err := verifyBodyGetObjectIfMatch(res, objectBody, shouldFail); err != nil {
		return err
	}
	if err := verifyStatusGetObjectIfMatch(res, expectedStatus); err != nil {
		return err
	}
	return nil
}

// verifyHeaderGetObjectIfMatch - Verify that the response header matches what is expected.
func verifyHeaderGetObjectIfMatch(res *http.Response) error {
	if err := verifyStandardHeaders(res); err != nil {
		return err
	}
	return nil
}

// verifyBodyGetObjectIfMatch - Verify that the response body matches what is expected.
func verifyBodyGetObjectIfMatch(res *http.Response, objectBody []byte, shouldFail bool) error {
	if shouldFail {
		// Decode the supposed error response.
		errBody := ErrorResponse{}
		err := xmlDecoder(res.Body, &errBody)
		if err != nil {
			return err
		}
		if errBody.Code != "PreconditionFailed" {
			err := fmt.Errorf("Unexpected Error Response: wanted PreconditionFailed, got %v", errBody.Code)
			return err
		}
	} else {
		// The body should be returned in full.
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		if !shouldFail && !bytes.Equal(body, objectBody) { // Test should pass ensure body is what was uploaded.
			err := fmt.Errorf("Unexpected Body Recieved: wanted %v, got %v", string(objectBody), string(body))
			return err
		}
	}
	// Otherwise test failed / passed as expected.
	return nil
}

// verifyStatusGetObjectIfMatch - Verify that the response status matches what is expected.
func verifyStatusGetObjectIfMatch(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus {
		err := fmt.Errorf("Unexpected Response Status Code: wanted %v, got %v", expectedStatus, res.Status)
		return err
	}
	return nil
}

// Test the compatibility of the GET object API when using the If-Match header.
func mainGetObjectIfMatch(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] GetObject (If-Match):", curTest, globalTotalNumTest)
	// Run the test on every object in every bucket.
	// Set up an invalid ETag to test failed requests responses.
	invalidETag := "1234567890"

	bucket := validBuckets[0]
	for _, object := range objects {
		// Test with If-Match Header set.
		// Spin scanBar
		scanBar(message)
		// Create new GET object If-Match request.
		req, err := newGetObjectIfMatchReq(config, bucket.Name, object.Key, object.ETag)
		if err != nil {
			printMessage(message, err)
			return false
		}
		// Spin scanBar
		scanBar(message)
		// Execute the request.
		res, err := execRequest(req, config.Client)
		if err != nil {
			printMessage(message, err)
			return false
		}
		// Spin scanBar
		scanBar(message)
		// Verify the response...these checks do not check the headers yet.
		if err := getObjectIfMatchVerify(res, object.Body, "200 OK", false); err != nil {
			printMessage(message, err)
			return false
		}
		// Spin scanBar
		scanBar(message)
		// Create a bad GET object If-Match request.
		badReq, err := newGetObjectIfMatchReq(config, bucket.Name, object.Key, invalidETag)
		if err != nil {
			printMessage(message, err)
			return false
		}
		// Spin scanBar
		scanBar(message)
		// Execute the request.
		badRes, err := execRequest(badReq, config.Client)
		if err != nil {
			printMessage(message, err)
			return false
		}
		// Spin scanBar
		scanBar(message)
		// Verify the request fails as expected.
		if err := getObjectIfMatchVerify(badRes, []byte(""), "412 Precondition Failed", true); err != nil {
			printMessage(message, err)
			return false
		}
		// Spin scanBar
		scanBar(message)

	}
	printMessage(message, nil)
	return true
}
