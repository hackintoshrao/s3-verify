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
	"time"

	"github.com/minio/s3verify/signv4"
)

// newGetObjectIfUnModifiedSinceReq - Create a new HTTP GET request with the If-Unmodified-Since header set to perform.
func newGetObjectIfUnModifiedSinceReq(config ServerConfig, bucketName, objectName string, lastModified time.Time) (*http.Request, error) {
	// An HTTP GET request with the If-Unmodified-Since header set.
	var getObjectIfUnModifiedSinceReq = &http.Request{
		Header: map[string][]string{
			// Set Content SHA with empty body for GET requests because no data is being uploaded.
			"X-Amz-Content-Sha256": {hex.EncodeToString(signv4.Sum256([]byte{}))},
		},
		Body:   nil, // There is no body for GET requests.
		Method: "GET",
	}
	targetURL, err := makeTargetURL(config.Endpoint, bucketName, objectName, config.Region)
	if err != nil {
		return nil, err
	}
	getObjectIfUnModifiedSinceReq.Header.Set("If-Unmodified-Since", lastModified.Format(http.TimeFormat))

	// Fill request URL and sign.
	getObjectIfUnModifiedSinceReq.URL = targetURL
	getObjectIfUnModifiedSinceReq = signv4.SignV4(*getObjectIfUnModifiedSinceReq, config.Access, config.Secret, config.Region)
	return getObjectIfUnModifiedSinceReq, nil
}

// verifyGetObjectIfUnModifiedSince - Verify the response matches what is expected.
func verifyGetObjectIfUnModifiedSince(res *http.Response, expectedBody []byte, expectedStatus string, shouldFail bool) error {
	if err := verifyBodyGetObjectIfUnModifiedSince(res, expectedBody, shouldFail); err != nil {
		return err
	}
	if err := verifyStatusGetObjectIfUnModifiedSince(res, expectedStatus); err != nil {
		return err
	}
	if err := verifyHeaderGetObjectIfUnModifiedSince(res); err != nil {
		return err
	}
	return nil
}

// verifyGetObjectIfUnModifiedSinceBody - Verify that the response body matches what is expected.
func verifyBodyGetObjectIfUnModifiedSince(res *http.Response, expectedBody []byte, shouldFail bool) error {
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
		if !bytes.Equal(body, expectedBody) {
			err := fmt.Errorf("Unexpected Body Received: wanted %v, got %v", string(expectedBody), string(body))
			return err
		}
	}
	// Otherwise test failed / passed as expected.
	return nil
}

// verifyStatusGetObjectIfUnModifiedSince - Verify that the response status matches what is expected.
func verifyStatusGetObjectIfUnModifiedSince(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus {
		err := fmt.Errorf("Unexpected Response Status: wanted %v, got %v", expectedStatus, res.Status)
		return err
	}
	return nil
}

// verifyHeaderGetObjectIfUnModifiedSince - Verify that the header returned matches what is expected.
func verifyHeaderGetObjectIfUnModifiedSince(res *http.Response) error {
	if err := verifyStandardHeaders(res); err != nil {
		return err
	}
	return nil
}

// Test the GET object API with the If-Unmodified-Since header set.
func mainGetObjectIfUnModifiedSince(config ServerConfig, message string) error {
	// Set up past date.
	pastDate, err := time.Parse(http.TimeFormat, "Thu, 01 Jan 1970 00:00:00 GMT")
	if err != nil {
		return err
	}
	bucket := testBuckets[0]
	for _, object := range objects {
		// Spin scanBar
		scanBar(message)
		// Form a request with a pastDate to make sure the object is not returned.
		req, err := newGetObjectIfUnModifiedSinceReq(config, bucket.Name, object.Key, pastDate)
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
		// Verify that the response returns an error.
		if err := verifyGetObjectIfUnModifiedSince(res, []byte(""), "412 Precondition Failed", true); err != nil {
			return err
		}
		// Spin scanBar
		scanBar(message)
		// Form a request with a date in the past.
		curReq, err := newGetObjectIfUnModifiedSinceReq(config, bucket.Name, object.Key, object.LastModified)
		if err != nil {
			return err
		}
		// Spin scanBar
		scanBar(message)
		// Execute current request.
		curRes, err := execRequest(curReq, config.Client)
		if err != nil {
			return err
		}
		// Spin scanBar
		scanBar(message)
		// Verify that the lastModified date in a request returns the object.
		if err := verifyGetObjectIfUnModifiedSince(curRes, object.Body, "200 OK", false); err != nil {
			return err
		}
		// Spin scanBar
		scanBar(message)

	}
	return nil
}
