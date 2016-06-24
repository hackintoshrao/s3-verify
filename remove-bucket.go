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
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/minio/minio-go"
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
	targetURL, err := url.Parse(config.Endpoint)
	if err != nil {
		return nil, err
	}
	targetURL.Path = "/" + bucketName + "/"     // Default to path style.
	if isVirtualStyleHostSupported(targetURL) { // Virtual style supported, use virtual style.
		targetURL.Path = "/"
		targetURL.Host = bucketName + "." + targetURL.Host
	}
	RemoveBucketReq.URL = targetURL
	// Sign the necessary headers.
	RemoveBucketReq = signv4.SignV4(*RemoveBucketReq, config.Access, config.Secret, config.Region)

	return RemoveBucketReq, nil
}

// RemoveBucketVerify - Check a Response's Status, Headers, and Body for AWS S3 compliance.
func RemoveBucketVerify(res *http.Response) error {
	if err := VerifyHeaderRemoveBucket(res); err != nil {
		return err
	}
	if err := VerifyStatusRemoveBucket(res); err != nil {
		return err
	}
	if err := VerifyBodyRemoveBucket(res); err != nil {
		return err
	}
	return nil
}

// RemoveBucketInit - Set up the RemoveBucket test.
func RemoveBucketInit(config ServerConfig) (string, error) {
	// Create random bucket name.
	bucketName := randString(60, rand.NewSource(time.Now().UnixNano()), "s3verify-rb")

	// Only need host part of endpoint for Minio.
	hostURL, err := url.Parse(config.Endpoint)
	if err != nil {
		return bucketName, err
	}
	secure := true // Use HTTPS request.
	s3Client, err := minio.New(hostURL.Host, config.Access, config.Secret, secure)
	if err != nil {
		return bucketName, err
	}
	// Use Minio to create a test bucket.
	err = s3Client.MakeBucket(bucketName, config.Region)
	if err != nil {
		return bucketName, err
	}
	return bucketName, nil
}

// RemoveBucketCleanUp - Clean up after any successful and failed tests.
func RemoveBucketCleanUp(config ServerConfig, bucketName string) error {
	// Only need host part of endpoint for Minio.
	hostURL, err := url.Parse(config.Endpoint)
	if err != nil {
		return err
	}
	secure := true // Use HTTPS request.
	s3Client, err := minio.New(hostURL.Host, config.Access, config.Secret, secure)
	if err != nil {
		return err
	}
	// Explicitly remove the Minio created test bucket.
	err = s3Client.RemoveBucket(bucketName)
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchBucket" {
			return nil
		}
		return err
	}
	return nil
}

// TODO: right now only checks for correctly deleted buckets...need to add in checks for 'failed' tests.

// VerifyHeaderRemoveBucket - Check that the responses headers match the expected headers for a given DELETE Bucket request.
func VerifyHeaderRemoveBucket(res *http.Response) error {
	// TODO: Needs discussion on what to actually check here....
	return nil
}

// VerifyBodyRemoveBucket - Check that the body of the response matches the expected body for a given DELETE Bucket request.
func VerifyBodyRemoveBucket(res *http.Response) error {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	// A successful DELETE request will return an empty body.
	if string(body) != "" {
		err = fmt.Errorf("Unexpected text in body: %v", body)
		return err
	}
	return nil
}

// VerifyStatusRemoveBucket - Check that the status of the response matches the expected status for a given DELETE Bucket request.
func VerifyStatusRemoveBucket(res *http.Response) error {
	if res.StatusCode != http.StatusNoContent { // Successful DELETE request will result in 204 No Content.
		err := fmt.Errorf("Remove Bucket Failed with %v", res.StatusCode)
		return err
	}
	return nil
}
