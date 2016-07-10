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

var CopyObjectIfNoneMatchReq = &http.Request{
	Header: map[string][]string{
	// X-Amz-Content-Sha256 will be set dynamically.
	// x-amz-copy-source will be set dynamically.
	// x-amz-copy-source-if-match will be set dynamically.
	},
	Method: "PUT",
}

// NewPutObjectCopyIfNoneMatchReq - Create a new HTTP request for a CopyObject with the if-none-match header set.
func NewCopyObjectIfNoneMatchReq(config ServerConfig, sourceBucketName, sourceObjectName, destBucketName, destObjectName, ETag string, objectData []byte) (*http.Request, error) {
	targetURL, err := makeTargetURL(config.Endpoint, destBucketName, destObjectName, config.Region)
	if err != nil {
		return nil, err
	}
	CopyObjectIfNoneMatchReq.URL = targetURL
	// Compute the md5sum and sha256sum from the data to be uploaded.
	reader := bytes.NewReader(objectData)
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return nil, err
	}
	// Fill in the request header.
	CopyObjectIfNoneMatchReq.Header.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	CopyObjectIfNoneMatchReq.Header.Set("x-amz-copy-source", url.QueryEscape(sourceBucketName+"/"+sourceObjectName))
	CopyObjectIfNoneMatchReq.Header.Set("x-amz-copy-source-if-none-match", ETag)

	CopyObjectIfNoneMatchReq = signv4.SignV4(*CopyObjectIfNoneMatchReq, config.Access, config.Secret, config.Region)
	return CopyObjectIfNoneMatchReq, nil
}

// Verify that the response returned matches what is expected.
func CopyObjectIfNoneMatchVerify(res *http.Response, expectedStatus string, expectedError ErrorResponse) error {
	if err := VerifyStatusCopyObjectIfNoneMatch(res, expectedStatus); err != nil {
		return err
	}
	if err := VerifyHeaderCopyObjectIfNoneMatch(res); err != nil {
		return err
	}
	if err := VerifyBodyCopyObjectIfNoneMatch(res, expectedError); err != nil {
		return err
	}
	return nil
}

// VerifyStatusCopyIfNoneMatch - Verify that the response status matches what is expected.
func VerifyStatusCopyObjectIfNoneMatch(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus {
		err := fmt.Errorf("Unexpected Status Received: wanted %v, got %v", expectedStatus, res.Status)
		return err
	}
	return nil
}

// VerifyBodyCopyIfNoneMatch - Verify the body returned matches what is expected.
func VerifyBodyCopyObjectIfNoneMatch(res *http.Response, expectedError ErrorResponse) error {
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

// VerifyHeaderCopyIfNoneMatch - Verify that the header returned matches what is expected.
func VerifyHeaderCopyObjectIfNoneMatch(res *http.Response) error {
	if err := verifyStandardHeaders(res); err != nil {
		return err
	}
	return nil
}

// Test the CopyObject API with the if-none-match header set.
func mainCopyObjectIfNoneMatch(config ServerConfig, message string) error {
	// Spin scanBar
	scanBar(message)
	// Create unmatchable ETag.
	goodETag := "1234567890"

	sourceBucketName := testBuckets[0].Name
	destBucketName := testBuckets[1].Name
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
	// Copy object copies data on the server, there is nothing to be set in the body.
	body := []byte("")
	// Create a successful copy request.
	req, err := NewCopyObjectIfNoneMatchReq(config, sourceBucketName, sourceObject.Key, destBucketName, destObject.Key, goodETag, body)
	if err != nil {
		return err
	}
	// Execute the response.
	res, err := ExecRequest(req, config.Client)
	if err != nil {
		return err
	}
	// Verify the response.
	if err = CopyObjectIfNoneMatchVerify(res, "200 OK", ErrorResponse{}); err != nil {
		return err
	}
	// Create a bad copy request.
	badReq, err := NewCopyObjectIfNoneMatchReq(config, sourceBucketName, sourceObject.Key, destBucketName, destObject.Key, sourceObject.ETag, body)
	if err != nil {
		return err
	}
	// Execute the response.
	badRes, err := ExecRequest(badReq, config.Client)
	if err != nil {
		return err
	}
	// Verify the response errors out as it should.
	if err = CopyObjectIfNoneMatchVerify(badRes, "412 Precondition Failed", expectedError); err != nil {
		return err
	}
	return nil
}
