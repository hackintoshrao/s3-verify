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

// newHeadObjectIfModifiedSinceReq - Create a new HTTP request for HEAD object with if-modified-since header set.
func newHeadObjectIfModifiedSinceReq(config ServerConfig, bucketName, objectName string, lastModified time.Time) (*http.Request, error) {
	// headObjectIfModifiedSinceReq - a new HTTP request for HEAD object with if-modified-since header set.
	var headObjectIfModifiedSinceReq = &http.Request{
		Header: map[string][]string{
		// X-Amz-Content-Sha256 will be set below.
		// If-Modified-Since will be set below.
		},
		Body:   nil, // No body is sent in a HEAD request.
		Method: "HEAD",
	}
	reader := bytes.NewReader([]byte{})
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return nil, err
	}
	targetURL, err := makeTargetURL(config.Endpoint, bucketName, objectName, config.Region, nil)
	if err != nil {
		return nil, err
	}
	// Set the URL and Header.
	headObjectIfModifiedSinceReq.URL = targetURL
	headObjectIfModifiedSinceReq.Header.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	headObjectIfModifiedSinceReq.Header.Set("If-Modified-Since", lastModified.Format(http.TimeFormat))
	headObjectIfModifiedSinceReq = signv4.SignV4(*headObjectIfModifiedSinceReq, config.Access, config.Secret, config.Region)

	return headObjectIfModifiedSinceReq, nil
}

// headObjectIfModifiedSinceVerify - verify the response returned matches what is expected.
func headObjectIfModifiedSinceVerify(res *http.Response, expectedStatus string) error {
	if err := verifyStatusHeadObjectIfModifiedSince(res, expectedStatus); err != nil {
		return err
	}
	if err := verifyHeaderHeadObjectIfModifiedSince(res); err != nil {
		return err
	}
	if err := verifyBodyHeadObjectIfModifiedSince(res); err != nil {
		return err
	}
	return nil
}

// verifyStatusHeadObjectIfModifiedSince - verify the status returned matches what is expected.
func verifyStatusHeadObjectIfModifiedSince(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus {
		err := fmt.Errorf("Unexpected Status Received: wanted %v, got %v", expectedStatus, res.Status)
		return err
	}
	return nil
}

// verifyBodyHeadObjectIfModifiedSince - verify the body returned is empty.
func verifyBodyHeadObjectIfModifiedSince(res *http.Response) error {
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

// verifyHeaderHeadObjectIfModifiedSince - verify the header returned matches what is expected.
func verifyHeaderHeadObjectIfModifiedSince(res *http.Response) error {
	if err := verifyStandardHeaders(res); err != nil {
		return err
	}
	return nil
}

// mainHeadObjectIfModifiedSince - Entry point for the HEAD object with if-modified-since header set.
func mainHeadObjectIfModifiedSince(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] HeadObject (If-Modified-Since):", curTest, globalTotalNumTest)
	bucket := validBuckets[0]
	object := objects[0]
	lastModified, err := time.Parse(http.TimeFormat, "Thu, 01 Jan 1970 00:00:00 GMT")
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Create a new request.
	req, err := newHeadObjectIfModifiedSinceReq(config, bucket.Name, object.Key, lastModified)
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
	// Spin scanBar
	scanBar(message)
	// Verify the response.
	if err := headObjectIfModifiedSinceVerify(res, "200 OK"); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Create a bad request.
	badReq, err := newHeadObjectIfModifiedSinceReq(config, bucket.Name, object.Key, object.LastModified.Add(time.Hour*2))
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Execute the bad request.
	badRes, err := execRequest(badReq, config.Client, bucket.Name, object.Key)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Verify the bad request failed as expected.
	if err := headObjectIfModifiedSinceVerify(badRes, "304 Not Modified"); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	printMessage(message, nil)
	return true
}
