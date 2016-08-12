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
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// newCopyObjectReq - Create a new HTTP request for PUT object with copy-
func newCopyObjectReq(config ServerConfig, sourceBucketName, sourceObjectName, destBucketName, destObjectName string) (Request, error) {
	var copyObjectReq = Request{
		customHeader: http.Header{},
	}

	// Set the bucketName and objectName
	copyObjectReq.bucketName = destBucketName
	copyObjectReq.objectName = destObjectName

	// Body will be set by the server so don't upload any body here.
	reader := bytes.NewReader([]byte(""))
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return Request{}, err
	}
	// Fill request headers.
	// Content-MD5 should never be set for CopyObject API.
	copyObjectReq.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	copyObjectReq.customHeader.Set("x-amz-copy-source", url.QueryEscape(sourceBucketName+"/"+sourceObjectName))
	copyObjectReq.customHeader.Set("User-Agent", appUserAgent)

	return copyObjectReq, nil
}

// copyObjectVerify - Verify that the response returned matches what is expected.
func copyObjectVerify(res *http.Response, expectedStatusCode int) error {
	if err := verifyHeaderCopyObject(res.Header); err != nil {
		return err
	}
	if err := verifyBodyCopyObject(res.Body); err != nil {
		return err
	}
	if err := verifyStatusCopyObject(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	return nil
}

// verifyHeaderscopyObject - verify that the header returned matches what is expected.
func verifyHeaderCopyObject(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// verifyBodycopyObject - verify that the body returned is a valid CopyObject Result.
func verifyBodyCopyObject(resBody io.Reader) error {
	copyObjRes := copyObjectResult{}
	decoder := xml.NewDecoder(resBody)
	err := decoder.Decode(&copyObjRes)
	if err != nil {
		return err
	}
	return nil
}

// verifyStatusCopyObject - verify that the status returned matches what is expected.
func verifyStatusCopyObject(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Response Status Code: wanted %v, got %v", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// Test a PUT object request with the copy header set.
func mainCopyObject(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] CopyObject:", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)
	// All copy-object tests happen in s3verify created buckets
	// on s3verify created objects.
	sourceBucketName := s3verifyBuckets[0].Name
	destBucketName := s3verifyBuckets[1].Name
	sourceObject := s3verifyObjects[0]

	// TODO: create tests designed to fail.
	destObject := &ObjectInfo{
		Key: sourceObject.Key,
	}
	copyObjects = append(copyObjects, destObject)
	// Spin scanBar
	scanBar(message)
	// Create a new request.
	req, err := newCopyObjectReq(config, sourceBucketName, sourceObject.Key, destBucketName, destObject.Key)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Execute the request.
	res, err := config.execRequest("PUT", req)
	if err != nil {
		printMessage(message, err)
		return false
	}
	defer closeResponse(res)
	// Spin scanBar
	scanBar(message)
	// Verify the response.
	if err = copyObjectVerify(res, http.StatusOK); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	printMessage(message, nil)
	return true
}
