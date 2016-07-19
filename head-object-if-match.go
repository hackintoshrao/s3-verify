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

// newHeadObjectIfMatchReq - Create a new HTTP request for HEAD object with if-match header set.
func newHeadObjectIfMatchReq(config ServerConfig, bucketName, objectName, ETag string) (*http.Request, error) {
	// headObjectIfMatchReq - an HTTP request for HEAD with if-match header set.
	var headObjectIfMatchReq = &http.Request{
		Header: map[string][]string{
		// Set Content SHA with an empty body for HEAD requests because no data is being uploaded.
		// Set If-Match header dynamically.
		},
		Body:   nil, // No body is sent with HEAD requests.
		Method: "HEAD",
	}
	targetURL, err := makeTargetURL(config.Endpoint, bucketName, objectName, config.Region, nil)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader([]byte(""))
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return nil, err
	}
	headObjectIfMatchReq.Header.Set("If-Match", ETag)
	headObjectIfMatchReq.Header.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	headObjectIfMatchReq.URL = targetURL
	headObjectIfMatchReq = signv4.SignV4(*headObjectIfMatchReq, config.Access, config.Secret, config.Region)
	return headObjectIfMatchReq, nil
}

// headObjectIfMatchVerify - verify that the returned response matches what is expected.
func headObjectIfMatchVerify(res *http.Response, expectedStatus string) error {
	if err := verifyStatusHeadObjectIfMatch(res, expectedStatus); err != nil {
		return err
	}
	if err := verifyBodyHeadObjectIfMatch(res); err != nil {
		return err
	}
	if err := verifyHeaderHeadObjectIfMatch(res); err != nil {
		return err
	}
	return nil
}

// verifyStatusHeadObjectIfMatch - verify the status returned matches what is expected.
func verifyStatusHeadObjectIfMatch(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus {
		err := fmt.Errorf("Unexpected Status Received: wanted %v, got %v", expectedStatus, res.Status)
		return err
	}
	return nil
}

// verifyBodyHeadObjectIfMatch - verify that the body returned matches what is expected.
func verifyBodyHeadObjectIfMatch(res *http.Response) error {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if !bytes.Equal(body, []byte{}) {
		err := fmt.Errorf("Unexpected Body Recieved: HEAD requests should not return a body, but got back: %v", string(body))
		return err
	}
	return nil
}

// verifyHeaderHeadObjectIfMatch - verify that the header returned matches what is expected.
func verifyHeaderHeadObjectIfMatch(res *http.Response) error {
	if err := verifyStandardHeaders(res); err != nil {
		return err
	}
	return nil
}

// mainHeadObjectIfMatch - Entry point for the HEAD object with if-match header set test.
func mainHeadObjectIfMatch(config ServerConfig, curTest int, printFunc func(string, error)) {
	message := fmt.Sprintf("[%02d/%d] HeadObject (If-Match):", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)
	// Create a bad ETag.
	invalidETag := "1234567890"
	bucket := validBuckets[0]
	object := objects[0]
	// Create a new valid request for HEAD object with if-match header set.
	req, err := newHeadObjectIfMatchReq(config, bucket.Name, object.Key, object.ETag)
	if err != nil {
		printFunc(message, err)
		return
	}
	// Spin scanBar
	scanBar(message)
	// Execute the request.
	res, err := execRequest(req, config.Client)
	if err != nil {
		printFunc(message, err)
		return
	}
	// Spin scanBar
	scanBar(message)
	// Verify the response.
	if err := headObjectIfMatchVerify(res, "200 OK"); err != nil {
		printFunc(message, err)
		return
	}
	// Spin scanBar
	scanBar(message)
	// Create a new invalid request for HEAD object with if-match header set.
	badReq, err := newHeadObjectIfMatchReq(config, bucket.Name, object.Key, invalidETag)
	if err != nil {
		printFunc(message, err)
		return
	}
	// Spin scanBar
	scanBar(message)
	// Execute the invalid request.
	badRes, err := execRequest(badReq, config.Client)
	if err != nil {
		printFunc(message, err)
		return
	}
	// Spin scanBar
	scanBar(message)
	// Verify the request sends back the right error.
	if err := headObjectIfMatchVerify(badRes, "412 Precondition Failed"); err != nil {
		printFunc(message, err)
		return
	}
	// Spin scanBar
	scanBar(message)
	printFunc(message, nil)
	return
}
