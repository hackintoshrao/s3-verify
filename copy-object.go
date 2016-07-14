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
	"net/http"
	"net/url"

	"github.com/minio/s3verify/signv4"
)

// newCopyObjectReq - Create a new HTTP request for PUT object with copy-
func newCopyObjectReq(config ServerConfig, sourceBucketName, sourceObjectName, destBucketName, destObjectName string) (*http.Request, error) {
	var copyObjectReq = &http.Request{
		Header: map[string][]string{
		// X-Amz-Content-Sha256 will be set dynamically.
		// x-amz-copy-source will be set dynamically.
		},
		Method: "PUT",
	}
	targetURL, err := makeTargetURL(config.Endpoint, destBucketName, destObjectName, config.Region, nil)
	if err != nil {
		return nil, err
	}
	// Fill request URL.
	copyObjectReq.URL = targetURL
	// Body will be set by the server so don't upload any body here.
	reader := bytes.NewReader([]byte(""))
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return nil, err
	}
	// Fill request headers.
	// Content-MD5 should never be set for CopyObject API.
	copyObjectReq.Header.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	copyObjectReq.Header.Set("x-amz-copy-source", url.QueryEscape(sourceBucketName+"/"+sourceObjectName))

	copyObjectReq = signv4.SignV4(*copyObjectReq, config.Access, config.Secret, config.Region)

	return copyObjectReq, nil
}

// copyObjectVerify - Verify that the response returned matches what is expected.
func copyObjectVerify(res *http.Response, expectedStatus string) error {
	if err := verifyHeaderCopyObject(res); err != nil {
		return err
	}
	if err := verifyBodyCopyObject(res); err != nil {
		return err
	}
	if err := verifyStatusCopyObject(res, expectedStatus); err != nil {
		return err
	}
	return nil
}

// verifyHeaderscopyObject - verify that the header returned matches what is expected.
func verifyHeaderCopyObject(res *http.Response) error {
	if err := verifyStandardHeaders(res); err != nil {
		return err
	}
	return nil
}

// verifyBodycopyObject - verify that the body returned is empty.
func verifyBodyCopyObject(res *http.Response) error {
	copyObjRes := copyObjectResult{}
	decoder := xml.NewDecoder(res.Body)
	err := decoder.Decode(&copyObjRes)
	if err != nil {
		return err
	}
	return nil
}

// verifyStatusCopyObject - verify that the status returned matches what is expected.
func verifyStatusCopyObject(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus {
		err := fmt.Errorf("Unexpected Response Status Code: wanted %v, got %v", expectedStatus, res.Status)
		return err
	}
	return nil
}

// Test a PUT object request with the copy header set.
func mainCopyObject(config ServerConfig, message string) error {
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
	req, err := newCopyObjectReq(config, sourceBucketName, sourceObject.Key, destBucketName, destObject.Key)
	if err != nil {
		return err
	}
	// Spin scanBar
	scanBar(message)
	// Execute the request.
	res, err := execRequest(req, config.Client)
	if err != nil {
		return err
	}
	// Spin scanBar
	scanBar(message)
	// Verify the response.
	if err = copyObjectVerify(res, "200 OK"); err != nil {
		return err
	}
	// Spin scanBar
	scanBar(message)
	return nil
}
