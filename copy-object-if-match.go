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

	"github.com/minio/s3verify/signv4"
)

// newCopyObjectIfMatchReq - Create a new HTTP request for a PUT copy object.
func newCopyObjectIfMatchReq(config ServerConfig, sourceBucketName, sourceObjectName, destBucketName, destObjectName, ETag string) (*http.Request, error) {
	var copyObjectIfMatchReq = &http.Request{
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
	copyObjectIfMatchReq.URL = targetURL
	// The body will be set by the server so calculate SHA from an empty body.
	reader := bytes.NewReader([]byte(""))
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return nil, err
	}
	// Fill in the request header.
	copyObjectIfMatchReq.Header.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	// Content-MD5 should not be set for CopyObject request.
	// Content-Length should not be set for CopyObject request.
	copyObjectIfMatchReq.Header.Set("x-amz-copy-source", url.QueryEscape(sourceBucketName+"/"+sourceObjectName))
	copyObjectIfMatchReq.Header.Set("x-amz-copy-source-if-match", ETag)

	copyObjectIfMatchReq = signv4.SignV4(*copyObjectIfMatchReq, config.Access, config.Secret, config.Region)
	return copyObjectIfMatchReq, nil
}

// copyObjectIfMatchVerify - Verify that the response returned matches what is expected.
func copyObjectIfMatchVerify(res *http.Response, expectedStatus string, expectedError ErrorResponse) error {
	if err := verifyBodyCopyObjectIfMatch(res, expectedError); err != nil {
		return err
	}
	if err := verifyHeaderCopyObjectIfMatch(res); err != nil {
		return err
	}
	if err := verifyStatusCopyObjectIfMatch(res, expectedStatus); err != nil {
		return err
	}
	return nil
}

// verifyHeaderCopyObjectIfMatch - Verify that the header returned matches what is expected.
func verifyHeaderCopyObjectIfMatch(res *http.Response) error {
	if err := verifyStandardHeaders(res); err != nil {
		return err
	}
	return nil
}

// verifyBodyCopyObjectIfMatch - Verify that the body returned matches what is expected.jK;
func verifyBodyCopyObjectIfMatch(res *http.Response, expectedError ErrorResponse) error {
	if expectedError.Message != "" { // Error is expected. Verify error returned matches.
		errResponse := ErrorResponse{}
		err := xmlDecoder(res.Body, &errResponse)
		if err != nil {
			return err
		}
		if errResponse.Message != expectedError.Message {
			err = fmt.Errorf("Unexpected error message: wanted %v, got %v", expectedError.Message, errResponse.Message)
			return err
		}
	} else { // Error unexpected. Body should be a copyobjectresult.
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

	}
	return nil
}

// verifyStatusCopyObjectIfMatch - Verify that the status returned matches what is expected.
func verifyStatusCopyObjectIfMatch(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus {
		err := fmt.Errorf("Unexpected Status Recieved: wanted %v, got %v", expectedStatus, res.Status)
		return err
	}
	return nil
}

// Test the PUT Object Copy with If-Match header is set.
func mainCopyObjectIfMatch(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] CopyObject (If-Match)", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)
	// Create bad ETag.
	badETag := "1234567890"

	sourceBucketName := validBuckets[0].Name
	destBucketName := validBuckets[1].Name
	sourceObject := objects[0]
	destObject := &ObjectInfo{
		Key: sourceObject.Key + "if-match",
	}
	copyObjects = append(copyObjects, destObject)
	expectedError := ErrorResponse{
		Code:    "PreconditionFailed",
		Message: "At least one of the pre-conditions you specified did not hold",
	}
	// Spin scanBar
	scanBar(message)
	// Create a new valid PUT object copy request.
	req, err := newCopyObjectIfMatchReq(config, sourceBucketName, sourceObject.Key, destBucketName, destObject.Key, sourceObject.ETag)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Execute the response.
	res, err := execRequest(req, config.Client, destBucketName, destObject.Key)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Verify the response.
	if err := copyObjectIfMatchVerify(res, "200 OK", ErrorResponse{}); err != nil {
		printMessage(message, err)
		return false
	}

	// Create a new invalid PUT object copy request.
	badReq, err := newCopyObjectIfMatchReq(config, sourceBucketName, sourceObject.Key, destBucketName, destObject.Key, badETag)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Execute the request.
	badRes, err := execRequest(badReq, config.Client, destBucketName, destObject.Key)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Verify the request failed as expected.
	if err := copyObjectIfMatchVerify(badRes, "412 Precondition Failed", expectedError); err != nil {
		printMessage(message, err)
		return false
	}
	printMessage(message, nil)
	return true
}
