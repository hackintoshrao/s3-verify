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

// MakeBucketReq - hardcode the static portions of a new Make Bucket request.
var MakeBucketReq = &http.Request{
	Header: map[string][]string{
		"X-Amz-Content-Sha256": {hex.EncodeToString(signv4.Sum256([]byte{}))},
	},
	Method: "PUT",
	Body:   nil, // No Body sent for Make Bucket requests.(Need to verify)
}

// NewMakeBucketReq - Create a new Make bucket request.
func NewMakeBucketReq(config ServerConfig, bucketName string) (*http.Request, error) {
	targetURL, err := url.Parse(config.Endpoint)
	if err != nil {
		return nil, err
	}
	targetURL.Path = "/" + bucketName + "/"     // Default to path style.
	if isVirtualStyleHostSupported(targetURL) { // Virtual style supported, use virtual style.
		targetURL.Path = "/"
		targetURL.Host = bucketName + "." + targetURL.Host
	}
	MakeBucketReq.URL = targetURL
	MakeBucketReq = signv4.SignV4(*MakeBucketReq, config.Access, config.Secret, config.Region)
	return MakeBucketReq, nil
}

// MakeBucketCleanUp - Remove the bucket created by the test after success / failure.
func MakeBucketCleanUp(config ServerConfig, bucketName string) error {
	// Minio only needs host part of the URL.
	targetURL, err := url.Parse(config.Endpoint)
	if err != nil {
		return err
	}
	secure := true // Use HTTPS request.
	s3Client, err := minio.New(targetURL.Host, config.Access, config.Secret, secure)
	if err != nil {
		return err
	}
	// Delete the bucket made by the request.
	err = s3Client.RemoveBucket(bucketName)
	if err != nil {
		// Bucket may not have been created successfully.
		if minio.ToErrorResponse(err).Code == "NoSuchBucket" { // Only use codes for now, strings unreliable.
			return nil
		}
		return err
	}
	return nil
}

// TODO: These checks only work on well formatted requests. Need to add support for poorly formed tests designed to fail.

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
	if location != "/"+bucketName && location != "/"+bucketName+"/" { // Remove second part of the and statement after minio bug is fixed.
		fmt.Println(location, "/"+bucketName+"/")
		err := fmt.Errorf("Unexpected Location: got %v, wanted %v", location, "/"+bucketName)
		return err
	}
	return nil
}

// Test the MakeBucket API when no extra headers are set.
func mainMakeBucketNoHeader(config ServerConfig, message string) error {
	// Spin the scanBar
	scanBar(message)
	// Generate new random bucket name.
	bucketName := randString(60, rand.NewSource(time.Now().UnixNano()), "s3verify-mb")
	// Spin the scanBar
	scanBar(message)

	// Create a new Make bucket request.
	req, err := NewMakeBucketReq(config, bucketName)
	if err != nil {
		// Attempt clean up.
		if errC := MakeBucketCleanUp(config, bucketName); errC != nil {
			return errC
		}
		return err
	}
	// Spin the scanBar
	scanBar(message)

	// Execute the request.
	res, err := ExecRequest(req)
	if err != nil {
		// Attempt clean up.
		if errC := MakeBucketCleanUp(config, bucketName); errC != nil {
			return errC
		}
		return err
	}
	// Spin the scanBar
	scanBar(message)

	// Check the responses Body, Status, Header.
	if err := VerifyResponseMakeBucket(res, bucketName); err != nil {
		// Attempt clean up.
		if errC := MakeBucketCleanUp(config, bucketName); errC != nil {
			return errC
		}
		return err
	}
	// Spin the scanBar
	scanBar(message)

	// Clean up the test.
	if err := MakeBucketCleanUp(config, bucketName); err != nil {
		return err
	}
	return nil
}
