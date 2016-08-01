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
	"io"
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
	// Set req URL and Header.
	targetURL, err := makeTargetURL(config.Endpoint, bucketName, objectName, config.Region, nil)
	if err != nil {
		return nil, err
	}
	getObjectIfMatchReq.Header.Set("If-Match", ETag)
	getObjectIfMatchReq.Header.Set("User-Agent", appUserAgent)
	// Fill request URL and sign
	getObjectIfMatchReq.URL = targetURL
	getObjectIfMatchReq = signv4.SignV4(*getObjectIfMatchReq, config.Access, config.Secret, config.Region)
	return getObjectIfMatchReq, nil
}

// getObjectIfMatchVerify - Verify that the response matches what is expected.
func getObjectIfMatchVerify(res *http.Response, objectBody []byte, expectedStatusCode int, shouldFail bool) error {
	if err := verifyHeaderGetObjectIfMatch(res.Header); err != nil {
		return err
	}
	if err := verifyBodyGetObjectIfMatch(res.Body, objectBody, shouldFail); err != nil {
		return err
	}
	if err := verifyStatusGetObjectIfMatch(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	return nil
}

// verifyHeaderGetObjectIfMatch - Verify that the response header matches what is expected.
func verifyHeaderGetObjectIfMatch(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// verifyBodyGetObjectIfMatch - Verify that the response body matches what is expected.
func verifyBodyGetObjectIfMatch(resBody io.Reader, objectBody []byte, shouldFail bool) error {
	if shouldFail {
		// Decode the supposed error response.
		errBody := ErrorResponse{}
		err := xmlDecoder(resBody, &errBody)
		if err != nil {
			return err
		}
		if errBody.Code != "PreconditionFailed" {
			err := fmt.Errorf("Unexpected Error Response: wanted PreconditionFailed, got %v", errBody.Code)
			return err
		}
	} else {
		// The body should be returned in full.
		body, err := ioutil.ReadAll(resBody)
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
func verifyStatusGetObjectIfMatch(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Response Status Code: wanted %v, got %v", expectedStatusCode, respStatusCode)
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
	errCh := make(chan error, globalTotalNumTest)
	bucket := validBuckets[0]
	// Spin scanBar
	scanBar(message)
	for _, object := range objects {
		// Test with If-Match Header set.
		// Spin scanBar
		scanBar(message)
		go func(objectKey, objectETag string, objectBody []byte) {
			// Create new GET object If-Match request.
			req, err := newGetObjectIfMatchReq(config, bucket.Name, objectKey, objectETag)
			if err != nil {
				errCh <- err
				return
			}
			// Execute the request.
			res, err := execRequest(req, config.Client, bucket.Name, objectKey)
			if err != nil {
				errCh <- err
				return
			}
			defer closeResponse(res)
			// Verify the response...these checks do not check the headers yet.
			if err := getObjectIfMatchVerify(res, objectBody, http.StatusOK, false); err != nil {
				errCh <- err
				return
			}
			// Create a bad GET object If-Match request.
			badReq, err := newGetObjectIfMatchReq(config, bucket.Name, objectKey, invalidETag)
			if err != nil {
				errCh <- err
				return
			}
			// Execute the request.
			badRes, err := execRequest(badReq, config.Client, bucket.Name, objectKey)
			if err != nil {
				errCh <- err
				return
			}
			defer closeResponse(badRes)
			// Verify the request fails as expected.
			if err := getObjectIfMatchVerify(badRes, []byte(""), http.StatusPreconditionFailed, true); err != nil {
				errCh <- err
				return
			}
			errCh <- nil
		}(object.Key, object.ETag, object.Body)
		// Spin scanBar
		scanBar(message)
	}
	count := len(objects)
	for count > 0 {
		count--
		// Spin scanBar
		scanBar(message)
		err, ok := <-errCh
		if !ok {
			return false
		}
		if err != nil {
			printMessage(message, err)
			return false
		}
		// Spin scanBar
		scanBar(message)
	}
	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, nil)
	return true
}
