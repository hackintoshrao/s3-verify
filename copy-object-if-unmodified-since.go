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
	"net/url"
	"time"

	"github.com/minio/s3verify/signv4"
)

// newCopyObjectIfUnModifiedSinceReq - Create a new HTTP request for CopyObject with if-unmodified-since header set.
func newCopyObjectIfUnModifiedSinceReq(config ServerConfig, sourceBucketName, sourceObjectName, destBucketName, destObjectName string, lastModified time.Time) (*http.Request, error) {
	// copyObjectIfUnModifiedSinceReq - A new HTTP request for CopyObject with if-unmodified-since header set.
	var copyObjectIfUnModifiedSinceReq = &http.Request{
		Header: map[string][]string{
		// X-Amz-Content-Sha256 will be set dynamically.
		// x-amz-copy-source will be set dynamically.
		// x-amz-copy-source-if-unmodified-since will be set dynamically.
		},
		Body:   nil, // For CopyObject requests the body is filled in on the server side.
		Method: "PUT",
	}
	targetURL, err := makeTargetURL(config.Endpoint, destBucketName, destObjectName, config.Region, nil)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader([]byte{})
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return nil, err
	}
	copyObjectIfUnModifiedSinceReq.URL = targetURL
	copyObjectIfUnModifiedSinceReq.Header.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	copyObjectIfUnModifiedSinceReq.Header.Set("x-amz-copy-source", url.QueryEscape(sourceBucketName+"/"+sourceObjectName))
	copyObjectIfUnModifiedSinceReq.Header.Set("x-amz-copy-if-unmodified-since", lastModified.Format(http.TimeFormat))

	copyObjectIfUnModifiedSinceReq = signv4.SignV4(*copyObjectIfUnModifiedSinceReq, config.Access, config.Secret, config.Region)
	return copyObjectIfUnModifiedSinceReq, nil
}

// copyObjectIfUnModifiedSinceVerify - verify the returned response matches what is expected.
func copyObjectIfUnModifiedSinceVerify(res *http.Response, expectedStatus string, expectedError ErrorResponse) error {
	if err := verifyStatusCopyObjectIfUnModifiedSince(res, expectedStatus); err != nil {
		return err
	}
	if err := verifyBodyCopyObjectIfUnModifiedSince(res, expectedError); err != nil {
		return err
	}
	if err := verifyHeaderCopyObjectIfUnModifiedSince(res); err != nil {
		return err
	}
	return nil
}

// verifyStatusCopyObjectIfUnModifiedSince - verify the status returned matches what is expected.
func verifyStatusCopyObjectIfUnModifiedSince(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus {
		err := fmt.Errorf("Unexpected Status Received: wanted %v, got %v", expectedStatus, res.Status)
		return err
	}
	return nil
}

// verifyBodyCopyObjectIfUnModifiedSince - verify the body returned matches what is expected.
func verifyBodyCopyObjectIfUnModifiedSince(res *http.Response, expectedError ErrorResponse) error {
	if expectedError.Message != "" {
		responseError := ErrorResponse{}
		err := xmlDecoder(res.Body, &responseError)
		if err != nil {
			return err
		}
		if responseError.Message != expectedError.Message {
			err := fmt.Errorf("Unexpected Error Message Received: wanted %v, got %v", expectedError.Message, responseError.Message)
			return err
		}
		return nil
	} else {
		// Verify the body returned is a copyobject result.
		copyObjResult := copyObjectResult{}
		err := xmlDecoder(res.Body, &copyObjResult)
		if err != nil {
			body, errR := ioutil.ReadAll(res.Body)
			if errR != nil {
				return errR
			}
			err = fmt.Errorf("Unexpected Body Received: %v", string(body))

			return err
		}
		return nil
	}
}

// verifyHeaderCopyObjectIfUnModifiedSince - verify the header returned matches what is expected.
func verifyHeaderCopyObjectIfUnModifiedSince(res *http.Response) error {
	if err := verifyStandardHeaders(res); err != nil {
		return err
	}
	return nil
}

// mainCopyObjectIfUnModifiedSince - Entry point for the CopyObject if-unmodified-since test.
func mainCopyObjectIfUnModifiedSince(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] CopyObject (If-Unmodified-Since): ", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)
	// Set a date in the past.
	pastDate, err := time.Parse(http.TimeFormat, "Thu, 01 Jan 1970 00:00:00 GMT")
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Source bucket and destination bucket.
	sourceBucket := validBuckets[0]
	destBucket := validBuckets[1]
	// Object to be copied.
	sourceObject := objects[0]
	destObject := &ObjectInfo{
		Key: sourceObject.Key + "if-unmodified-since",
	}
	// Expected error on failure.
	expectedError := ErrorResponse{
		Code:    "PreconditionFailed",
		Message: "At least one of the pre-conditions you specified did not hold",
	}
	// Create a new valid request.
	req, err := newCopyObjectIfUnModifiedSinceReq(config, sourceBucket.Name, sourceObject.Key, destBucket.Name, destObject.Key, sourceObject.LastModified.Add(time.Hour*2))
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Execute the request.
	res, err := execRequest(req, config.Client, destBucket.Name, destObject.Key)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Verify the response.
	if err := copyObjectIfUnModifiedSinceVerify(res, "200 OK", ErrorResponse{}); err != nil {
		printMessage(message, err)
		return false
	}
	// Add the copied object to the copyObjects slice.
	copyObjects = append(copyObjects, destObject)
	// Spin scanBar
	scanBar(message)

	// Create a new invalid request.
	badReq, err := newCopyObjectIfUnModifiedSinceReq(config, sourceBucket.Name, sourceObject.Key, destBucket.Name, destObject.Key, pastDate)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Execute the bad request.
	badRes, err := execRequest(badReq, config.Client, destBucket.Name, destObject.Key)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Verify the bad request fails with the proper error.
	if err := copyObjectIfUnModifiedSinceVerify(badRes, "412 Precondition Failed", expectedError); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, nil)
	return true
}
