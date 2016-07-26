/*
 * Minio S3verify Library for Amazon S3 Compatible Cloud Storage (C) 2016 Minio, Inc.
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

//
func newHeadBucketReq(config ServerConfig, bucketName string) (*http.Request, error) {
	//
	var headBucketReq = &http.Request{
		Header: map[string][]string{
		// X-Amz-Content-Sha256 will be set below.
		},
		Body:   nil, // There is no body sent with HEAD requests.
		Method: "HEAD",
	}
	targetURL, err := makeTargetURL(config.Endpoint, bucketName, "", config.Region, nil)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader([]byte{}) // Compute hash using empty body because HEAD requests do not send a body.
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return nil, err
	}
	// Set the URL and Header of the request.
	headBucketReq.URL = targetURL
	headBucketReq.Header.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))

	headBucketReq = signv4.SignV4(*headBucketReq, config.Access, config.Secret, config.Region)
	return headBucketReq, nil
}

// headBucketVerify - verify the response returned matches what is expected.
func headBucketVerify(res *http.Response, expectedStatus string) error {
	if err := verifyBodyHeadBucket(res); err != nil {
		return err
	}
	if err := verifyStatusHeadBucket(res, expectedStatus); err != nil {
		return err
	}
	if err := verifyHeaderHeadBucket(res); err != nil {
		return err
	}
	return nil
}

// verifyBodyHeadBucket - verify the body returned matches what is expected.
func verifyBodyHeadBucket(res *http.Response) error {
	// Verify that the body returned is empty.
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

// verifyHeaderHeadBucket - verify the header returned matches what is expected.
func verifyHeaderHeadBucket(res *http.Response) error {
	if err := verifyStandardHeaders(res); err != nil {
		return err
	}
	return nil
}

// verifyStatusHeadBucket - verify the status returned matches what is expected.
func verifyStatusHeadBucket(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus {
		err := fmt.Errorf("Unexpected Status Received: wanted %v, got %v", expectedStatus, res.Status)
		return err
	}
	return nil
}

// mainHeadBucket - Entry point for the HeadBucket API test.
func mainHeadBucket(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] HeadBucket:", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)
	bucketName := validBuckets[0].Name
	// Create a new request for one of the validBuckets.
	req, err := newHeadBucketReq(config, bucketName)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Execute the request.
	res, err := execRequest(req, config.Client)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Verify the response.
	if err := headBucketVerify(res, "200 OK"); err != nil {
		printMessage(message, err)
		return false
	}
	// Test passed.
	printMessage(message, nil)
	return true
}
