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
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"

	"github.com/minio/s3verify/signv4"
)

var PutObjectCopyReq = &http.Request{
	Header: map[string][]string{
	// X-Amz-Content-Sha256 will be set dynamically.
	// Content-MD5 will be set dynamically.
	// Content-Length will be set dynamically.
	// x-amz-copy-source will be set dynamically.
	},
	// Body will be set dynamically.
	// Body:
	Method: "PUT",
}

// NewPutObjectCopyReq - Create a new HTTP request for PUT object with copy-
func NewPutObjectCopyReq(config ServerConfig, sourceBucketName, sourceObjectName, destBucketName, destObjectName string, objectData []byte) (*http.Request, error) {
	targetURL, err := makeTargetURL(config.Endpoint, destBucketName, destObjectName, config.Region)
	if err != nil {
		return nil, err
	}
	// Fill request URL.
	PutObjectCopyReq.URL = targetURL

	// Compute md5Sum and sha256Sum from the input data.
	reader := bytes.NewReader(objectData)
	md5Sum, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return nil, err
	}
	// Fill request headers.
	PutObjectCopyReq.Header.Set("Content-MD5", base64.StdEncoding.EncodeToString(md5Sum))
	PutObjectCopyReq.Header.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	PutObjectCopyReq.Header.Set("x-amz-copy-source", url.QueryEscape(sourceBucketName+"/"+sourceObjectName))

	PutObjectCopyReq = signv4.SignV4(*PutObjectCopyReq, config.Access, config.Secret, config.Region)

	return PutObjectCopyReq, nil
}

//
func PutObjectCopyVerify(res *http.Response, expectedStatus string) error {
	if err := VerifyHeaderPutObjectCopy(res); err != nil {
		return err
	}
	if err := VerifyBodyPutObjectCopy(res); err != nil {
		return err
	}
	if err := VerifyStatusPutObjectCopy(res, expectedStatus); err != nil {
		return err
	}
	return nil
}

// VerifyHeadersPutObjectCopy - Verify that the header returned matches what is expected.
func VerifyHeaderPutObjectCopy(res *http.Response) error {
	if err := verifyStandardHeaders(res); err != nil {
		return err
	}
	return nil
}

// VerifyBodyPutObjectCopy - Verify that the body returned is empty.
func VerifyBodyPutObjectCopy(res *http.Response) error {
	copyObjRes := copyObjectResult{}
	decoder := xml.NewDecoder(res.Body)
	err := decoder.Decode(&copyObjRes)
	if err != nil {
		return err
	}
	return nil
}

// VerifyStatusPutObjectCopy - Verify that the status returned matches what is expected.
func VerifyStatusPutObjectCopy(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus {
		err := fmt.Errorf("Unexpected Response Status Code: wanted %v, got %v", expectedStatus, res.Status)
		return err
	}
	return nil
}

// Test a PUT object request with the copy header set.
func mainPutObjectCopy(config ServerConfig, message string) error {
	// Spin scanBar
	scanBar(message)
	// TODO: create tests designed to fail.
	sourceBucketName := testBuckets[0].Name
	destBucketName := testBuckets[1].Name
	sourceObject := objects[0]
	destObject := &ObjectInfo{
		Key: sourceObject.Key,
	}
	copyObjects = append(copyObjects, destObject)
	// Spin scanBar
	scanBar(message)
	// Create a new request.
	req, err := NewPutObjectCopyReq(config, sourceBucketName, sourceObject.Key, destBucketName, destObject.Key, sourceObject.Body)
	if err != nil {
		return err
	}
	// Spin scanBar
	scanBar(message)
	// Execute the request.
	res, err := ExecRequest(req, config.Client)
	if err != nil {
		return err
	}
	// Spin scanBar
	scanBar(message)
	// Verify the response.
	if err = PutObjectCopyVerify(res, "200 OK"); err != nil {
		return err
	}

	// Spin scanBar
	scanBar(message)
	return nil
}
