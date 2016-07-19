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
	"encoding/hex"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/minio/s3verify/signv4"
)

// newRemoveBucketReq - Fill in the dynamic fields of a DELETE request here.
func newRemoveBucketReq(config ServerConfig, bucketName string) (*http.Request, error) {
	// removeBucketReq is a new DELETE bucket request.
	var removeBucketReq = &http.Request{
		Header: map[string][]string{
			// Set Content SHA with empty body for GET / DELETE requests because no data is being uploaded.
			"X-Amz-Content-Sha256": {hex.EncodeToString(signv4.Sum256([]byte{}))},
		},
		Method: "DELETE",
		Body:   nil, // There is no body for GET / DELETE requests.
	}

	// Set the DELETE req URL.
	targetURL, err := makeTargetURL(config.Endpoint, bucketName, "", config.Region, nil)
	if err != nil {
		return nil, err
	}
	removeBucketReq.URL = targetURL
	// Sign the necessary headers.
	removeBucketReq = signv4.SignV4(*removeBucketReq, config.Access, config.Secret, config.Region)
	return removeBucketReq, nil
}

// removeBucketVerify - Check a Response's Status, Headers, and Body for AWS S3 compliance.
func removeBucketVerify(res *http.Response, expectedStatus string, errorResponse ErrorResponse) error {
	if err := verifyHeaderRemoveBucket(res); err != nil {
		return err
	}
	if err := verifyStatusRemoveBucket(res, expectedStatus); err != nil {
		return err
	}
	if err := verifyBodyRemoveBucket(res, errorResponse); err != nil {
		return err
	}
	return nil
}

// TODO: right now only checks for correctly deleted buckets...need to add in checks for 'failed' tests.

// verifyHeaderRemoveBucket - Check that the responses headers match the expected headers for a given DELETE Bucket request.
func verifyHeaderRemoveBucket(res *http.Response) error {
	if err := verifyStandardHeaders(res); err != nil {
		return err
	}
	return nil
}

// verifyBodyRemoveBucket - Check that the body of the response matches the expected body for a given DELETE Bucket request.
func verifyBodyRemoveBucket(res *http.Response, expectedError ErrorResponse) error {
	if expectedError.Message != "" { // Error is expected.
		errResponse := ErrorResponse{}
		err := xmlDecoder(res.Body, &errResponse)
		if err != nil {
			return err
		}
		if errResponse.Message != expectedError.Message {
			err := fmt.Errorf("Unexpected Error: %v", errResponse.Message)
			return err
		}
	}
	return nil
}

// verifyStatusRemoveBucket - Check that the status of the response matches the expected status for a given DELETE Bucket request.
func verifyStatusRemoveBucket(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus { // Successful DELETE request will result in 204 No Content.
		err := fmt.Errorf("Unexpected Status: wanted %v, got %v", expectedStatus, res.StatusCode)
		return err
	}
	return nil
}

// Test the RemoveBucket API when the bucket exists.
func mainRemoveBucketExists(config ServerConfig, curTest int, printFunc func(string, error)) {
	message := fmt.Sprintf("[%02d/%d] RemoveBucket (Bucket Exists):", curTest, globalTotalNumTest)
	for _, bucket := range validBuckets {
		// Spin the scanBar
		scanBar(message)

		// Generate the new DELETE bucket request.
		req, err := newRemoveBucketReq(config, bucket.Name)
		if err != nil {
			printFunc(message, err)
			return
		}
		// Spin the scanBar
		scanBar(message)

		// Perform the request.
		res, err := execRequest(req, config.Client)
		if err != nil {
			printFunc(message, err)
			return
		}
		// Spin the scanBar
		scanBar(message)

		if err = removeBucketVerify(res, "204 No Content", ErrorResponse{}); err != nil {
			printFunc(message, err)
			return
		}
		// Spin the scanBar
		scanBar(message)
	}
	printFunc(message, nil)
	return
}

// Test the RemoveBucket API when the bucket does not exist.
func mainRemoveBucketDNE(config ServerConfig, curTest int, printFunc func(string, error)) {
	message := fmt.Sprintf("[%02d/%d] RemoveBucket (Bucket DNE):", curTest, globalTotalNumTest)
	// Generate a random bucketName.
	bucketName := randString(60, rand.NewSource(time.Now().UnixNano()), "")
	// Hardcode the expected error response.
	errResponse := ErrorResponse{
		Code:       "NoSuchBucket",
		Message:    "The specified bucket does not exist",
		BucketName: bucketName,
		Key:        "",
	}
	// Spin scanBar
	scanBar(message)
	// Generate a new DELETE bucket request for a bucket that does not exist.
	req, err := newRemoveBucketReq(config, bucketName)
	if err != nil {
		printFunc(message, err)
		return
	}
	// Spin scanBar
	scanBar(message)
	// Perform the request.
	res, err := execRequest(req, config.Client)
	if err != nil {
		printFunc(message, err)
		return
	}
	// Spin scanBar
	scanBar(message)
	if err = removeBucketVerify(res, "404 Not Found", errResponse); err != nil {
		printFunc(message, err)
		return
	}
	// Spin scanBar
	scanBar(message)
	printFunc(message, nil)
	return
}
