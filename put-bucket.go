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
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/minio/s3verify/signv4"
)

var (
	testBuckets = [][]BucketInfo{
		validBuckets,
		invalidBuckets,
	}
	validBuckets = []BucketInfo{
		BucketInfo{
			Name: "s3verify-put-bucket-test",
		},
		BucketInfo{
			Name: "s3verify-put-bucket-copy-test",
		},
	}
	// See http://docs.aws.amazon.com/AmazonS3/latest/dev/BucketRestrictions.html for all bucket naming restrictions.
	invalidBuckets = []BucketInfo{
		BucketInfo{
			Name: "s3", // Bucket names must be at least 3 chars long.
		},
		BucketInfo{
			Name: "babcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwzyz", // Bucket names must be less than 63 chars long. This is only on regions other than us-east-1.
		},
		BucketInfo{
			Name: "S3verify", // Bucket names must start with a lowercase letter or a number.
		},
		BucketInfo{
			Name: "192.168.5.4", // Bucket names must not be formatted as an IP address.
		},
		BucketInfo{
			Name: "s3..verify", // Bucket names can not have adjacent periods in them.
		},
		BucketInfo{
			Name: ".s3verify", // Bucket names can not start with periods.
		},
		BucketInfo{
			Name: "s3verify.", // Bucket names can not end with periods.
		},
	}
)

// newPutBucketReq - Create a new Make bucket request.
func newPutBucketReq(config ServerConfig, bucketName string) (*http.Request, error) {
	// putBucketReq - hardcode the static portions of a new Make Bucket request.
	var putBucketReq = &http.Request{
		Header: map[string][]string{
			"X-Amz-Content-Sha256": {hex.EncodeToString(signv4.Sum256([]byte{}))},
		},
		Method: "PUT",
		// Body: will be set if the region is different from us-east-1.
	}

	targetURL, err := makeTargetURL(config.Endpoint, bucketName, "", config.Region, nil)
	if err != nil {
		return nil, err
	}
	putBucketReq.URL = targetURL
	if config.Region != globalDefaultRegion { // Must set the request elements for non us-east-1 regions.
		bucketConfig := createBucketConfiguration{}
		bucketConfig.Location = config.Region
		bucketConfigBytes, err := xml.Marshal(bucketConfig)
		if err != nil {
			return nil, err
		}
		bucketConfigBuffer := bytes.NewReader(bucketConfigBytes)
		_, sha256Sum, contentLength, err := computeHash(bucketConfigBuffer)
		if err != nil {
			return nil, err
		}
		// Set the body.
		putBucketReq.Body = ioutil.NopCloser(bucketConfigBuffer)
		putBucketReq.ContentLength = contentLength
		// Fill request headers and URL.
		putBucketReq.Header.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	}
	putBucketReq = signv4.SignV4(*putBucketReq, config.Access, config.Secret, config.Region)
	return putBucketReq, nil
}

// putBucketVerify - Check the response Body, Header, Status for AWS S3 compliance.
func putBucketVerify(res *http.Response, bucketName, expectedStatus string, expectedError ErrorResponse) error {
	if err := verifyStatusPutBucket(res, expectedStatus); err != nil {
		return err
	}
	if err := verifyHeaderPutBucket(res, bucketName, expectedStatus); err != nil {
		return err
	}
	if err := verifyBodyPutBucket(res, expectedError); err != nil {
		return err
	}
	return nil
}

// verifyStatusPutBucket - Check the response status for AWS S3 compliance.
func verifyStatusPutBucket(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus {
		err := fmt.Errorf("Unexpected Response Status Code: wanted %v, got %v", expectedStatus, res.StatusCode)
		return err
	}
	return nil
}

// verifyBodyPutBucket - Check the response body for AWS S3 compliance.
func verifyBodyPutBucket(res *http.Response, expectedError ErrorResponse) error {
	if expectedError.Message != "" {
		resError := ErrorResponse{}
		err := xmlDecoder(res.Body, &resError)
		if err != nil {
			return err
		}
		if resError.Message != expectedError.Message {
			err := fmt.Errorf("Unexpected Error Message: wanted %v, got %v", expectedError.Message, resError.Message)
			return err
		}
		return nil
	} else {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		// There is no body returned by a Put Bucket request.
		if string(body) != "" {
			err := fmt.Errorf("Unexpected Body: %v", string(body))
			return err
		}
	}
	return nil
}

// verifyHeaderPutBucket - Check the response header for AWS S3 compliance.
func verifyHeaderPutBucket(res *http.Response, bucketName, expectedStatus string) error {
	if expectedStatus == "200 OK" {
		location := res.Header["Location"][0]
		if location != "http://"+bucketName+".s3.amazonaws.com/" && location != "/"+bucketName {
			// TODO: wait for Minio server to fix endpoint detection.
			err := fmt.Errorf("Unexpected Location: got %v", location)
			return err
		}
	}
	if err := verifyStandardHeaders(res); err != nil {
		return err
	}
	return nil
}

// Test the PutBucket API when no extra headers are set. This creates three new buckets and leaves them for the next tests to use.
func mainPutBucket(config ServerConfig, message string) error {
	// Spin the scanBar
	scanBar(message)
	// Test that valid names are PUT correctly.
	for _, bucket := range validBuckets {
		// Spin the scanBar
		scanBar(message)
		// Create a new Make bucket request.
		req, err := newPutBucketReq(config, bucket.Name)
		if err != nil {
			return err
		}
		// Spin the scanBar
		scanBar(message)
		// Execute the request.
		res, err := execRequest(req, config.Client)
		if err != nil {
			return err
		}
		// Spin the scanBar
		scanBar(message)
		// Check the responses Body, Status, Header.
		if err := putBucketVerify(res, bucket.Name, "200 OK", ErrorResponse{}); err != nil {
			return err
		}
		// Spin the scanBar
		scanBar(message)
	}
	expectedError := ErrorResponse{
		Message: "The specified bucket is not valid.",
	}
	// Test that all invalid names fail correctly.
	for _, bucket := range invalidBuckets {
		// Spin scanBar
		scanBar(message)
		// Create a new PUT bucket request.
		req, err := newPutBucketReq(config, bucket.Name)
		if err != nil {
			return err
		}
		// Spin scanBar
		scanBar(message)
		// Execute the request.
		res, err := execRequest(req, config.Client)
		if err != nil {
			return err
		}
		// Spin scanBar
		scanBar(message)
		// Verify that the request failed as predicted.
		if err := putBucketVerify(res, bucket.Name, "400 Bad Request", expectedError); err != nil {
			return err
		}
		// Spin scanBar
		scanBar(message)
	}
	return nil
}
