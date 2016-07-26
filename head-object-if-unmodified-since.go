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

// newHeadObjectIfUnModifiedReq - Create a new HTTP request for HEAD object with if-unmodified-since header set.
func newHeadObjectIfUnModifiedSinceReq(config ServerConfig, bucketName, objectName string, lastModified time.Time) (*http.Request, error) {
	// headObjectIfUnModifiedReq - a new HTTP request for HEAD object with if-unmodified-since header set.
	var headObjectIfUnModifiedSinceReq = &http.Request{
		Header: map[string][]string{
		// X-Amz-Content-Sha256 will be set below.
		// If-Unmodified-Since will be set below.
		},
		Body:   nil, // No body is sent in HEAD object requests.
		Method: "HEAD",
	}

	targetURL, err := makeTargetURL(config.Endpoint, bucketName, objectName, config.Region, nil)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader([]byte{})
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return nil, err
	}
	headObjectIfUnModifiedSinceReq.Header.Set("If-Unmodified-Since", lastModified.Format(http.TimeFormat))
	headObjectIfUnModifiedSinceReq.Header.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	headObjectIfUnModifiedSinceReq.URL = targetURL
	headObjectIfUnModifiedSinceReq = signv4.SignV4(*headObjectIfUnModifiedSinceReq, config.Access, config.Secret, config.Region)

	return headObjectIfUnModifiedSinceReq, nil
}

// headObjectIfUnModifiedSinceVerify - verify the response returned matches what is expected.
func headObjectIfUnModifiedSinceVerify(res *http.Response, expectedStatus string) error {
	if err := verifyStatusHeadObjectIfUnModifiedSince(res, expectedStatus); err != nil {
		return err
	}
	if err := verifyBodyHeadObjectIfUnModifiedSince(res); err != nil {
		return err
	}
	if err := verifyHeaderHeadObjectIfUnModifiedSince(res); err != nil {
		return err
	}
	return nil
}

// verifyStatusHeadObjectIfUnModifiedSince - verify the status returned matches what is expected.
func verifyStatusHeadObjectIfUnModifiedSince(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus {
		err := fmt.Errorf("Unexpected Status Received: wanted %v, got %v", expectedStatus, res.Status)
		return err
	}
	return nil
}

// verifyBodyHeadObjectIfUnModifiedSince - verify the body returned is emtpy.
func verifyBodyHeadObjectIfUnModifiedSince(res *http.Response) error {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if !bytes.Equal(body, []byte{}) {
		err := fmt.Errorf("Unexpected Body Received: %v", string(body))
		return err
	}
	return nil
}

// verifyHeaderHeadObjectIfUnModifiedSince - verify that the header returned matches what is expected.
func verifyHeaderHeadObjectIfUnModifiedSince(res *http.Response) error {
	if err := verifyStandardHeaders(res); err != nil {
		return err
	}
	return nil
}

// mainHeadObjectIfUnModifiedSince - Entry point for the HEAD object with if-unmodified-since header set test.
func mainHeadObjectIfUnModifiedSince(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] HeadObject (If-Unmodified-Since):", curTest, globalTotalNumTest)
	scanBar(message)
	bucket := validBuckets[0]
	object := objects[0]
	// Create a date in the past to use.
	lastModified, err := time.Parse(http.TimeFormat, "Thu, 01 Jan 1970 00:00:00 GMT")
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Create a new request.
	req, err := newHeadObjectIfUnModifiedSinceReq(config, bucket.Name, object.Key, object.LastModified)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Perform the request.
	res, err := execRequest(req, config.Client)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Verify the request succeeds as expected.
	if err := headObjectIfUnModifiedSinceVerify(res, "200 OK"); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Create a bad request.
	badReq, err := newHeadObjectIfUnModifiedSinceReq(config, bucket.Name, object.Key, lastModified)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Perform the bad request.
	badRes, err := execRequest(badReq, config.Client)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Verify the response failed.
	if err := headObjectIfUnModifiedSinceVerify(badRes, "412 Precondition Failed"); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, nil)
	return true
}
