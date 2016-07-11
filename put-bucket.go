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
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/minio/s3verify/signv4"
)

var testBuckets = []BucketInfo{
	BucketInfo{
		Name: "s3verify-put-bucket-test",
	},
	BucketInfo{
		Name: "s3verify-put-bucket-test1",
	},
}

// newPutBucketReq - Create a new Make bucket request.
func newPutBucketReq(config ServerConfig, bucketName string) (*http.Request, error) {

	// putBucketReq - hardcode the static portions of a new Make Bucket request.
	var putBucketReq = &http.Request{
		Header: map[string][]string{
			"X-Amz-Content-Sha256": {hex.EncodeToString(signv4.Sum256([]byte{}))},
		},
		Method: "PUT",
		Body:   nil, // No Body sent for Make Bucket requests.(Need to verify)
	}

	targetURL, err := makeTargetURL(config.Endpoint, bucketName, "", config.Region)
	if err != nil {
		return nil, err
	}
	putBucketReq.URL = targetURL
	if config.Region != "us-east-1" { // Must set the request elements for non us-east-1 regions.
		bucketConfig := createBucketConfiguration{}
		bucketConfig.Location = config.Region
		bucketConfigBytes, err := xml.Marshal(bucketConfig)
		if err != nil {
			return nil, err
		}
		// TODO: use hash function here.
		bucketConfigBuffer := bytes.NewReader(bucketConfigBytes)
		putBucketReq.ContentLength = int64(bucketConfigBuffer.Size())
		// Reset X-Amz-Content-Sha256
		var hashSHA256 hash.Hash
		// Create a reader from the data.
		hashSHA256 = sha256.New()
		_, err = io.Copy(hashSHA256, bucketConfigBuffer)
		if err != nil {
			return nil, err
		}
		// Move back to beginning of data.
		bucketConfigBuffer.Seek(0, 0)
		// Set the body.
		putBucketReq.Body = ioutil.NopCloser(bucketConfigBuffer)
		// Finalize the SHA calculation.
		sha256Sum := hashSHA256.Sum(nil)
		// Fill request headers and URL.
		putBucketReq.Header.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	}
	putBucketReq = signv4.SignV4(*putBucketReq, config.Access, config.Secret, config.Region)
	return putBucketReq, nil
}

// verifyResponsePutBucket - Check the response Body, Header, Status for AWS S3 compliance.
func verifyResponsePutBucket(res *http.Response, bucketName string) error {
	if err := verifyStatusPutBucket(res); err != nil {
		return err
	}
	if err := verifyHeaderPutBucket(res, bucketName); err != nil {
		return err
	}
	if err := verifyBodyPutBucket(res); err != nil {
		return err
	}
	return nil
}

// verifyStatusPutBucket - Check the response status for AWS S3 compliance.
func verifyStatusPutBucket(res *http.Response) error {
	if res.StatusCode != http.StatusOK {
		err := fmt.Errorf("Unexpected Response Status Code: %v", res.StatusCode)
		return err
	}
	return nil
}

// verifyBodyPutBucket - Check the response body for AWS S3 compliance.
func verifyBodyPutBucket(res *http.Response) error {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	// There is no body returned by a Put Bucket request.
	if string(body) != "" {
		err := fmt.Errorf("Unexpected Body: %v", string(body))
		return err
	}
	return nil
}

// verifyHeaderPutBucket - Check the response header for AWS S3 compliance.
func verifyHeaderPutBucket(res *http.Response, bucketName string) error {
	location := res.Header["Location"][0]
	if location != "http://"+bucketName+".s3.amazonaws.com/" && location != "/"+bucketName {
		// TODO: wait for Minio server to fix endpoint detection.
		err := fmt.Errorf("Unexpected Location: got %v", location)
		return err
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
	for _, bucket := range testBuckets { // Test the creation of 3 new buckets. All valid for now. TODO: test invalid names/input.
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
		if err := verifyResponsePutBucket(res, bucket.Name); err != nil {
			return err
		}
		// Spin the scanBar
		scanBar(message)
	}

	return nil
}
