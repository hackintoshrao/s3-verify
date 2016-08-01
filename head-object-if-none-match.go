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

// newHeadObjectIfNoneMatch - Create a new HTTP request for HEAD object with if-none-match header set.
func newHeadObjectIfNoneMatchReq(config ServerConfig, bucketName, objectName, ETag string) (*http.Request, error) {
	//
	var headObjectIfNoneMatchReq = &http.Request{
		Header: map[string][]string{
		// X-Amz-Content-Sha256 will be set below.
		// If-None-Match will be set below.
		},
		Body:   nil, // There is no body sent by HEAD requests.
		Method: "HEAD",
	}
	targetURL, err := makeTargetURL(config.Endpoint, bucketName, objectName, config.Region, nil)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader([]byte{})
	_, sha256Sum, _, err := computeHash(reader)

	// Set requests URL and Header.
	headObjectIfNoneMatchReq.URL = targetURL
	headObjectIfNoneMatchReq.Header.Set("If-None-Match", ETag)
	headObjectIfNoneMatchReq.Header.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	headObjectIfNoneMatchReq.Header.Set("User-Agent", appUserAgent)

	headObjectIfNoneMatchReq = signv4.SignV4(*headObjectIfNoneMatchReq, config.Access, config.Secret, config.Region)
	return headObjectIfNoneMatchReq, nil
}

// headObjectIfNoneMatchVerify - verify the returned response matches what is expected.
func headObjectIfNoneMatchVerify(res *http.Response, expectedStatusCode int) error {
	if err := verifyStatusHeadObjectIfNoneMatch(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	if err := verifyBodyHeadObjectIfNoneMatch(res.Body); err != nil {
		return err
	}
	if err := verifyHeaderHeadObjectIfNoneMatch(res.Header); err != nil {
		return err
	}
	return nil
}

// verifyStatusHeadObjectIfNoneMatch - verify the returned status matches what is expected.
func verifyStatusHeadObjectIfNoneMatch(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Status Received: wanted %v, got %v", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// verifyBodyHeadObjectIfNoneMatch - verify the body returned is empty.
func verifyBodyHeadObjectIfNoneMatch(resBody io.Reader) error {
	body, err := ioutil.ReadAll(resBody)
	if err != nil {
		return err
	}
	if !bytes.Equal(body, []byte{}) {
		err := fmt.Errorf("Unexpected Body Recieved: HEAD requests should not return a body, but got back: %v", string(body))
		return err
	}
	return nil
}

// verifyHeaderHeadObjectIfNoneMatch - verify the header returned matches what is expected.
func verifyHeaderHeadObjectIfNoneMatch(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// mainHeadObjectIfNoneMatch - Entry point for the HEAD object with if-none-match header set test.
func mainHeadObjectIfNoneMatch(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] HeadObject (If-None-Match):", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)
	// Create an ETag that won't match any already created.
	validETag := "1234567890"
	bucket := validBuckets[0]
	object := objects[0]
	// Create a new request for a HEAD object with if-none-match header set.
	req, err := newHeadObjectIfNoneMatchReq(config, bucket.Name, object.Key, validETag)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Execute the request.
	res, err := execRequest(req, config.Client, bucket.Name, object.Key)
	if err != nil {
		printMessage(message, err)
		return false
	}
	defer closeResponse(res)
	// Spin scanBar
	scanBar(message)
	// Verify the response.
	if err := headObjectIfNoneMatchVerify(res, http.StatusOK); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Create a new invalid request for a HEAD object with if-none-match header set.
	badReq, err := newHeadObjectIfNoneMatchReq(config, bucket.Name, object.Key, object.ETag)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Execute the request.
	badRes, err := execRequest(badReq, config.Client, bucket.Name, object.Key)
	if err != nil {
		printMessage(message, err)
		return false
	}
	defer closeResponse(badRes)
	// Spin scanBar
	scanBar(message)
	// Verify the response.
	if err := headObjectIfNoneMatchVerify(badRes, 304); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	printMessage(message, nil)
	return true
}
