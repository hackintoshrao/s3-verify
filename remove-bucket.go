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
	"reflect"
	"time"

	"github.com/minio/s3verify/signv4"
)

// RemoveBucketReq is a new DELETE bucket request.
var RemoveBucketReq = &http.Request{
	Header: map[string][]string{
		// Set Content SHA with empty body for GET / DELETE requests because no data is being uploaded.
		"X-Amz-Content-Sha256": {hex.EncodeToString(signv4.Sum256([]byte{}))},
	},
	Method: "DELETE",
	Body:   nil, // There is no body for GET / DELETE requests.
}

// NewRemoveBucketReq - Fill in the dynamic fields of a DELETE request here.
func NewRemoveBucketReq(config ServerConfig, bucketName string) (*http.Request, error) {
	// Set the DELETE req URL.
	targetURL, err := makeTargetURL(config.Endpoint, bucketName, "", config.Region)
	if err != nil {
		return nil, err
	}
	RemoveBucketReq.URL = targetURL
	// Sign the necessary headers.
	RemoveBucketReq = signv4.SignV4(*RemoveBucketReq, config.Access, config.Secret, config.Region)

	return RemoveBucketReq, nil
}

// RemoveBucketVerify - Check a Response's Status, Headers, and Body for AWS S3 compliance.
func RemoveBucketVerify(res *http.Response, expectedStatus string, errorResponse ErrorResponse) error {
	if err := VerifyHeaderRemoveBucket(res); err != nil {
		return err
	}
	if err := VerifyStatusRemoveBucket(res, expectedStatus); err != nil {
		return err
	}
	if err := VerifyBodyRemoveBucket(res, errorResponse); err != nil {
		return err
	}
	return nil
}

// TODO: right now only checks for correctly deleted buckets...need to add in checks for 'failed' tests.

// VerifyHeaderRemoveBucket - Check that the responses headers match the expected headers for a given DELETE Bucket request.
func VerifyHeaderRemoveBucket(res *http.Response) error {
	if err := verifyStandardHeaders(res); err != nil {
		return err
	}
	return nil
}

// VerifyBodyRemoveBucket - Check that the body of the response matches the expected body for a given DELETE Bucket request.
func VerifyBodyRemoveBucket(res *http.Response, expectedError ErrorResponse) error {
	if !reflect.DeepEqual(expectedError, ErrorResponse{}) {
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

// VerifyStatusRemoveBucket - Check that the status of the response matches the expected status for a given DELETE Bucket request.
func VerifyStatusRemoveBucket(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus { // Successful DELETE request will result in 204 No Content.
		err := fmt.Errorf("Unexpected Status: wanted %v, got %v", expectedStatus, res.StatusCode)
		return err
	}
	return nil
}

// Test the RemoveBucket API when the bucket exists.
func mainRemoveBucketExists(config ServerConfig, message string) error {
	for _, bucket := range testBuckets {
		// Spin the scanBar
		scanBar(message)

		// Generate the new DELETE bucket request.
		req, err := NewRemoveBucketReq(config, bucket.Name)
		if err != nil {
			return err
		}
		// Spin the scanBar
		scanBar(message)

		// Perform the request.
		res, err := ExecRequest(req, config.Client)
		if err != nil {
			return err
		}
		// Spin the scanBar
		scanBar(message)

		if err = RemoveBucketVerify(res, "204 No Content", ErrorResponse{}); err != nil {
			return err
		}
		// Spin the scanBar
		scanBar(message)
	}
	return nil
}

// Test the RemoveBucket API when the bucket exists.
func mainRemoveBucketDNE(config ServerConfig, message string) error {
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
	req, err := NewRemoveBucketReq(config, bucketName)
	if err != nil {
		return err
	}
	// Spin scanBar
	scanBar(message)
	// Perform the request.
	res, err := ExecRequest(req, config.Client)
	if err != nil {
		return err
	}
	// Spin scanBar
	scanBar(message)
	if err = RemoveBucketVerify(res, "404 Not Found", errResponse); err != nil {
		return err
	}
	// Spin scanBar
	scanBar(message)
	return nil
}
