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
)

// newHeadObjectIfMatchReq - Create a new HTTP request for HEAD object with if-match header set.
func newHeadObjectIfMatchReq(config ServerConfig, bucketName, objectName, ETag string) (Request, error) {
	// headObjectIfMatchReq - an HTTP request for HEAD with if-match header set.
	var headObjectIfMatchReq = Request{
		customHeader: http.Header{},
	}

	// Set bucketName and objectName.
	headObjectIfMatchReq.bucketName = bucketName
	headObjectIfMatchReq.objectName = objectName

	// HEAD requests send no body.
	reader := bytes.NewReader([]byte{})
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return Request{}, err
	}
	headObjectIfMatchReq.customHeader.Set("User-Agent", appUserAgent)
	headObjectIfMatchReq.customHeader.Set("If-Match", ETag)
	headObjectIfMatchReq.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))

	return headObjectIfMatchReq, nil
}

// headObjectIfMatchVerify - verify that the returned response matches what is expected.
func headObjectIfMatchVerify(res *http.Response, expectedStatusCode int) error {
	if err := verifyStatusHeadObjectIfMatch(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	if err := verifyBodyHeadObjectIfMatch(res.Body); err != nil {
		return err
	}
	if err := verifyHeaderHeadObjectIfMatch(res.Header); err != nil {
		return err
	}
	return nil
}

// verifyStatusHeadObjectIfMatch - verify the status returned matches what is expected.
func verifyStatusHeadObjectIfMatch(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Status Received: wanted %v, got %v", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// verifyBodyHeadObjectIfMatch - verify that the body returned matches what is expected.
func verifyBodyHeadObjectIfMatch(resBody io.Reader) error {
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

// verifyHeaderHeadObjectIfMatch - verify that the header returned matches what is expected.
func verifyHeaderHeadObjectIfMatch(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// mainHeadObjectIfMatch - Entry point for the HEAD object with if-match header set test.
func mainHeadObjectIfMatch(config ServerConfig, curTest int) bool {
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
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Execute the request.
	res, err := config.execRequest("HEAD", req)
	if err != nil {
		printMessage(message, err)
		return false
	}
	defer closeResponse(res)
	// Spin scanBar
	scanBar(message)
	// Verify the response.
	if err := headObjectIfMatchVerify(res, http.StatusOK); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Create a new invalid request for HEAD object with if-match header set.
	badReq, err := newHeadObjectIfMatchReq(config, bucket.Name, object.Key, invalidETag)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Execute the invalid request.
	badRes, err := config.execRequest("HEAD", badReq)
	if err != nil {
		printMessage(message, err)
		return false
	}
	defer closeResponse(badRes)
	// Spin scanBar
	scanBar(message)
	// Verify the request sends back the right error.
	if err := headObjectIfMatchVerify(badRes, http.StatusPreconditionFailed); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	printMessage(message, nil)
	return true
}
