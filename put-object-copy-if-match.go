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
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"

	"github.com/minio/s3verify/signv4"
)

var PutObjectCopyIfMatchReq = &http.Request{
	Header: map[string][]string{
	// X-Amz-Content-Sha256 will be set dynamically.
	// Content-MD5 will be set dynamically.
	// Content-Length will be set dynamically.
	// x-amz-copy-source will be set dynamically.
	// x-amz-copy-source-if-match will be set dynamically.
	},
	Method: "PUT",
}

// NewPutObjectCopyIfMatchReq - Create a new HTTP request for a PUT copy object.
func NewPutObjectCopyIfMatchReq(config ServerConfig, sourceBucketName, sourceObjectName, destBucketName, destObjectName, ETag string, objectData []byte) (*http.Request, error) {
	targetURL, err := makeTargetURL(config.Endpoint, destBucketName, destObjectName, config.Region)
	if err != nil {
		return nil, err
	}
	PutObjectCopyIfMatchReq.URL = targetURL
	// Compute the md5sum and sha256sum from the data to be uploaded.
	reader := bytes.NewReader(objectData)
	md5Sum, sha256Sum, contentLength, err := computeHash(reader)
	if err != nil {
		return nil, err
	}
	// Fill in the request header.
	PutObjectCopyIfMatchReq.Header.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	PutObjectCopyIfMatchReq.Header.Set("Content-MD5", base64.StdEncoding.EncodeToString(md5Sum))
	PutObjectCopyIfMatchReq.Header.Set("Content-Length", strconv.FormatInt(contentLength, 10))
	PutObjectCopyIfMatchReq.Header.Set("x-amz-copy-source", url.QueryEscape(sourceBucketName+"/"+sourceObjectName))
	PutObjectCopyIfMatchReq.Header.Set("x-amz-copy-source-if-match", ETag)

	PutObjectCopyIfMatchReq = signv4.SignV4(*PutObjectCopyIfMatchReq, config.Access, config.Secret, config.Region)
	return PutObjectCopyIfMatchReq, nil
}

// PutObjectCopyIfMatchVerify - Verify that the response returned matches what is expected.
func PutObjectCopyIfMatchVerify(res *http.Response, expectedStatus string, expectedError ErrorResponse) error {
	if err := VerifyBodyPutObjectCopyIfMatch(res, expectedError); err != nil {
		return err
	}
	if err := VerifyHeaderPutObjectCopyIfMatch(res); err != nil {
		return err
	}
	if err := VerifyStatusPutObjectCopyIfMatch(res, expectedStatus); err != nil {
		return err
	}
	return nil
}

// VerifyHeaderPutObjectCopyIfMatch - Verify that the header returned matches what is expected.
func VerifyHeaderPutObjectCopyIfMatch(res *http.Response) error {
	if err := verifyStandardHeaders(res); err != nil {
		return err
	}
	return nil
}

// VerifyBodyPutObjectCopyIfMatch - Verify that the body returned matches what is expected.jK;
func VerifyBodyPutObjectCopyIfMatch(res *http.Response, expectedError ErrorResponse) error {
	if !reflect.DeepEqual(expectedError, ErrorResponse{}) { // Error is expected. Verify error returned matches.
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

// VerifyStatusPutObjectCopyIfMatch - Verify that the status returned matches what is expected.
func VerifyStatusPutObjectCopyIfMatch(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus {
		err := fmt.Errorf("Unexpected Status Recieved: wanted %v, got %v", expectedStatus, res.Status)
		return err
	}
	return nil
}

// Test the PUT Object Copy with If-Match header is set.
func mainPutObjectCopyIfMatch(config ServerConfig, message string) error {
	// Spin scanBar
	scanBar(message)
	// Create bad ETag.
	badETag := "1234567890"

	sourceBucketName := testBuckets[0].Name
	destBucketName := testBuckets[1].Name
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
	req, err := NewPutObjectCopyIfMatchReq(config, sourceBucketName, sourceObject.Key, destBucketName, destObject.Key, sourceObject.ETag, sourceObject.Body)
	if err != nil {
		return err
	}
	// Execute the response.
	res, err := ExecRequest(req, config.Client)
	if err != nil {
		return err
	}
	// Verify the response.
	if err := PutObjectCopyIfMatchVerify(res, "200 OK", ErrorResponse{}); err != nil {
		return err
	}

	// Create a new invalid PUT object copy request.
	badReq, err := NewPutObjectCopyIfMatchReq(config, sourceBucketName, sourceObject.Key, destBucketName, destObject.Key, badETag, sourceObject.Body)
	if err != nil {
		return err
	}
	// Execute the request.
	badRes, err := ExecRequest(badReq, config.Client)
	if err != nil {
		return err
	}
	// Verify the request failed as expected.
	if err := PutObjectCopyIfMatchVerify(badRes, "412 Precondition Failed", expectedError); err != nil {
		return err
	}
	return nil
}
