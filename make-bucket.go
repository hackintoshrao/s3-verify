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

// MakeBucketReq - hardcode the static portions of a new Make Bucket request.
var MakeBucketReq = &http.Request{
	Header: map[string][]string{
		"X-Amz-Content-Sha256": {hex.EncodeToString(signv4.Sum256([]byte{}))},
	},
	Method: "PUT",
	Body:   nil, // No Body sent for Make Bucket requests.(Need to verify)
}

var testBuckets = []BucketInfo{
	BucketInfo{
		Name: "s3verify-put-bucket-test",
	},
	BucketInfo{
		Name: "s3verify-put-bucket-test1",
	},
}

// NewMakeBucketReq - Create a new Make bucket request.
func NewMakeBucketReq(config ServerConfig, bucketName string) (*http.Request, error) {
	targetURL, err := makeTargetURL(config.Endpoint, bucketName, "", config.Region)
	if err != nil {
		return nil, err
	}
	MakeBucketReq.URL = targetURL
	if config.Region != "us-east-1" { // Must set the request elements for non us-east-1 regions.
		bucketConfig := createBucketConfiguration{}
		bucketConfig.Location = config.Region
		bucketConfigBytes, err := xml.Marshal(bucketConfig)
		if err != nil {
			return nil, err
		}
		// TODO: use hash function here.
		bucketConfigBuffer := bytes.NewReader(bucketConfigBytes)
		MakeBucketReq.ContentLength = int64(bucketConfigBuffer.Size())
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
		MakeBucketReq.Body = ioutil.NopCloser(bucketConfigBuffer)
		// Finalize the SHA calculation.
		sha256Sum := hashSHA256.Sum(nil)
		// Fill request headers and URL.
		MakeBucketReq.Header.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	}
	MakeBucketReq = signv4.SignV4(*MakeBucketReq, config.Access, config.Secret, config.Region)
	return MakeBucketReq, nil
}

// VerifyResponseMakeBucket - Check the response Body, Header, Status for AWS S3 compliance.
func VerifyResponseMakeBucket(res *http.Response, bucketName string) error {
	if err := VerifyStatusMakeBucket(res); err != nil {
		return err
	}
	if err := VerifyHeaderMakeBucket(res, bucketName); err != nil {
		return err
	}
	if err := VerifyBodyMakeBucket(res); err != nil {
		return err
	}
	return nil
}

// VerifyStatusMakeBucket - Check the response status for AWS S3 compliance.
func VerifyStatusMakeBucket(res *http.Response) error {
	if res.StatusCode != http.StatusOK {
		err := fmt.Errorf("Unexpected Response Status Code: %v", res.StatusCode)
		return err
	}
	return nil
}

// VerifyBodyMakeBucket - Check the response body for AWS S3 compliance.
func VerifyBodyMakeBucket(res *http.Response) error {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	// There is no body returned by a Make bucket request.
	if string(body) != "" {
		err := fmt.Errorf("Unexpected Body: %v", string(body))
		return err
	}
	return nil
}

// VerifyHeaderMakeBucket - Check the response header for AWS S3 compliance.
func VerifyHeaderMakeBucket(res *http.Response, bucketName string) error {
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

// Test the MakeBucket API when no extra headers are set. This creates three new buckets and leaves them for the next tests to use.
func mainMakeBucketNoHeader(config ServerConfig, message string) error {
	// Spin the scanBar
	scanBar(message)
	for _, bucket := range testBuckets { // Test the creation of 3 new buckets. All valid for now. TODO: test invalid names/input.
		// Spin the scanBar
		scanBar(message)

		// Create a new Make bucket request.
		req, err := NewMakeBucketReq(config, bucket.Name)
		if err != nil {
			return err
		}
		// Spin the scanBar
		scanBar(message)

		// Execute the request.
		res, err := ExecRequest(req, config.Client)
		if err != nil {
			return err
		}
		// Spin the scanBar
		scanBar(message)

		// Check the responses Body, Status, Header.
		if err := VerifyResponseMakeBucket(res, bucket.Name); err != nil {
			return err
		}
		// Spin the scanBar
		scanBar(message)
	}

	return nil
}
