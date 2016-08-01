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
	// Set req URL and Header.
	targetURL, err := makeTargetURL(config.Endpoint, bucketName, objectName, config.Region, nil)
	if err != nil {
		return nil, err
	}
	getObjectIfNoneMatchReq.Header.Set("If-None-Match", ETag)
	getObjectIfNoneMatchReq.Header.Set("User-Agent", appUserAgent)
	// Add the URL and sign
	getObjectIfNoneMatchReq.URL = targetURL
	getObjectIfNoneMatchReq = signv4.SignV4(*getObjectIfNoneMatchReq, config.Access, config.Secret, config.Region)
	return getObjectIfNoneMatchReq, nil
}

// getObjectIfNoneMatchVerify - Verify that the response matches with what is expected.
func getObjectIfNoneMatchVerify(res *http.Response, objectBody []byte, expectedStatusCode int) error {
	if err := verifyHeaderGetObjectIfNoneMatch(res.Header); err != nil {
		return err
	}
	if err := verifyStatusGetObjectIfNoneMatch(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	if err := verifyBodyGetObjectIfNoneMatch(res.Body, objectBody); err != nil {
		return err
	}
	return nil
}

// verifyHeaderGetObjectIfNoneMatch - Verify that the header fields of the response match what is expected.
func verifyHeaderGetObjectIfNoneMatch(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// verifyStatusGetObjectIfNoneMatch - Verify that the response status matches what is expected.
func verifyStatusGetObjectIfNoneMatch(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Response Status Code: wanted %v, got %v", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// verifyBodyGetObjectIfNoneMatch - Verify that the response body matches what is expected.
func verifyBodyGetObjectIfNoneMatch(resBody io.Reader, expectedBody []byte) error {
	// The body should be returned in full.
	body, err := ioutil.ReadAll(resBody)
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
func mainGetObjectIfNoneMatch(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] GetObject (If-None-Match):", curTest, globalTotalNumTest)
	// Set up an invalid ETag to test failed requests responses.
	invalidETag := "1234567890"
	bucket := validBuckets[0]
	// Spin scanBar
	scanBar(message)
	errCh := make(chan error, globalTotalNumTest)
	for _, object := range objects {
		// Spin scanBar
		scanBar(message)
		// Test with If-None-Match Header set.
		go func(objectKey, objectETag string, objectBody []byte) {
			// Create new GET object If-None-Match request.
			req, err := newGetObjectIfNoneMatchReq(config, bucket.Name, objectKey, objectETag)
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
			if err := getObjectIfNoneMatchVerify(res, []byte(""), 304); err != nil {
				errCh <- err
				return
			}
			// Create a bad GET object If-None-Match request with invalid ETag.
			badReq, err := newGetObjectIfNoneMatchReq(config, bucket.Name, objectKey, invalidETag)
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
			// Verify the response returns the object since ETag != invalidETag
			if err := getObjectIfNoneMatchVerify(badRes, objectBody, http.StatusOK); err != nil {
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
