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
	"net/http"
	"net/url"

	"github.com/minio/s3verify/signv4"
)

// newPutObjectCopyIfNoneMatchReq - Create a new HTTP request for a CopyObject with the if-none-match header set.
func newCopyObjectIfNoneMatchReq(config ServerConfig, sourceBucketName, sourceObjectName, destBucketName, destObjectName, ETag string) (*http.Request, error) {
	var copyObjectIfNoneMatchReq = &http.Request{
		Header: map[string][]string{
		// X-Amz-Content-Sha256 will be set dynamically.
		// x-amz-copy-source will be set dynamically.
		// x-amz-copy-source-if-match will be set dynamically.
		},
		Method: "PUT",
	}
	targetURL, err := makeTargetURL(config.Endpoint, destBucketName, destObjectName, config.Region, nil)
	if err != nil {
		return nil, err
	}
	copyObjectIfNoneMatchReq.URL = targetURL
	// Body will be calculated by the server so no body needs to be sent in the request.
	reader := bytes.NewReader([]byte(""))
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return nil, err
	}
	// Fill in the request header.
	copyObjectIfNoneMatchReq.Header.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	copyObjectIfNoneMatchReq.Header.Set("x-amz-copy-source", url.QueryEscape(sourceBucketName+"/"+sourceObjectName))
	copyObjectIfNoneMatchReq.Header.Set("x-amz-copy-source-if-none-match", ETag)

	copyObjectIfNoneMatchReq = signv4.SignV4(*copyObjectIfNoneMatchReq, config.Access, config.Secret, config.Region)
	return copyObjectIfNoneMatchReq, nil
}

// Verify that the response returned matches what is expected.
func copyObjectIfNoneMatchVerify(res *http.Response, expectedStatus string, expectedError ErrorResponse) error {
	if err := verifyStatusCopyObjectIfNoneMatch(res, expectedStatus); err != nil {
		return err
	}
	if err := verifyHeaderCopyObjectIfNoneMatch(res); err != nil {
		return err
	}
	if err := verifyBodyCopyObjectIfNoneMatch(res, expectedError); err != nil {
		return err
	}
	return nil
}

// verifyStatusCopyIfNoneMatch - Verify that the response status matches what is expected.
func verifyStatusCopyObjectIfNoneMatch(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus {
		err := fmt.Errorf("Unexpected Status Received: wanted %v, got %v", expectedStatus, res.Status)
		return err
	}
	return nil
}

// verifyBodyCopyIfNoneMatch - Verify the body returned matches what is expected.
func verifyBodyCopyObjectIfNoneMatch(res *http.Response, expectedError ErrorResponse) error {
	if expectedError.Message != "" { // Error is expected.
		errResponse := ErrorResponse{}
		if err := xmlDecoder(res.Body, &errResponse); err != nil {
			return err
		}
		if errResponse.Message != expectedError.Message {
			err := fmt.Errorf("Unexpected Error Response: wanted %v, got %v", expectedError.Message, errResponse.Message)
			return err
		}
		return nil
	} else { // Successful copy expected.
		copyObjRes := copyObjectResult{}
		if err := xmlDecoder(res.Body, &copyObjRes); err != nil {
			return err
		}
		return nil
	}
}

// verifyHeaderCopyIfNoneMatch - Verify that the header returned matches what is expected.
func verifyHeaderCopyObjectIfNoneMatch(res *http.Response) error {
	if err := verifyStandardHeaders(res); err != nil {
		return err
	}
	return nil
}

// Test the CopyObject API with the if-none-match header set.
func mainCopyObjectIfNoneMatch(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] CopyObject (If-None-Match):", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)
	// Create unmatchable ETag.
	goodETag := "1234567890"

	sourceBucketName := validBuckets[0].Name
	destBucketName := validBuckets[1].Name
	sourceObject := objects[0]
	destObject := &ObjectInfo{
		Key: sourceObject.Key + "if-none-match",
	}
	copyObjects = append(copyObjects, destObject)
	// Create an error for the case that is expected to fail.
	expectedError := ErrorResponse{
		Code:    "PreconditionFailed",
		Message: "At least one of the pre-conditions you specified did not hold",
	}
	// Create a successful copy request.
	req, err := newCopyObjectIfNoneMatchReq(config, sourceBucketName, sourceObject.Key, destBucketName, destObject.Key, goodETag)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Execute the response.
	res, err := execRequest(req, config.Client)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Verify the response.
	if err = copyObjectIfNoneMatchVerify(res, "200 OK", ErrorResponse{}); err != nil {
		printMessage(message, err)
		return false
	}
	// Create a bad copy request.
	badReq, err := newCopyObjectIfNoneMatchReq(config, sourceBucketName, sourceObject.Key, destBucketName, destObject.Key, sourceObject.ETag)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Execute the response.
	badRes, err := execRequest(badReq, config.Client)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Verify the response errors out as it should.
	if err = copyObjectIfNoneMatchVerify(badRes, "412 Precondition Failed", expectedError); err != nil {
		printMessage(message, err)
		return false
	}
	printMessage(message, nil)
	return true
}
