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

// newListPartsReq - Create a new HTTP request for the ListParts API.
func newListPartsReq(config ServerConfig, bucketName, objectName, uploadID string) (*http.Request, error) {
	// listPartsReq - a new HTTP request for ListParts.
	var listPartsReq = &http.Request{
		Header: map[string][]string{
		// X-Amz-Content-Sha256 will be set below.
		},
		Body:   nil, // There is no body sent for GET requests.
		Method: "GET",
	}
	// Create new url queries.
	urlValues := make(url.Values)
	urlValues.Set("uploadId", uploadID)

	targetURL, err := makeTargetURL(config.Endpoint, bucketName, objectName, config.Region, urlValues)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader([]byte{})
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return nil, err
	}
	// Set the requests URL and Header values.
	listPartsReq.URL = targetURL
	listPartsReq.Header.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	listPartsReq = signv4.SignV4(*listPartsReq, config.Access, config.Secret, config.Region)
	return listPartsReq, nil
}

// listPartsVerify - verify that the returned response matches what is expected.
func listPartsVerify(res *http.Response, expectedStatus string, expectedList listObjectPartsResult) error {
	if err := verifyStatusListParts(res, expectedStatus); err != nil {
		return err
	}
	if err := verifyBodyListParts(res, expectedList); err != nil {
		return err
	}
	if err := verifyHeaderListParts(res); err != nil {
		return err
	}
	return nil
}

// verifyStatusListParts - verify that the status returned matches what is expected.
func verifyStatusListParts(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus {
		err := fmt.Errorf("Unexpected Status Received: wanted %v, got %v", expectedStatus, res.Status)
		return err
	}
	return nil
}

// verifyBodyListParts - verify that the returned body matches whats expected.
func verifyBodyListParts(res *http.Response, expectedList listObjectPartsResult) error {
	result := listObjectPartsResult{}
	err := xmlDecoder(res.Body, &result)
	if err != nil {
		return err
	}
	totalParts := 0
	for _, part := range expectedList.ObjectParts {
		for _, resPart := range result.ObjectParts {
			if part.PartNumber == resPart.PartNumber && "\""+part.ETag+"\"" == resPart.ETag {
				totalParts++
			}
		}
	}
	if totalParts != 1 {
		err := fmt.Errorf("Incorrect number of parts listed: wanted 1, got %v", totalParts)
		return err
	}
	return nil
}

// verifyHeaderListParts - verify the header returned matches what is expected.
func verifyHeaderListParts(res *http.Response) error {
	if err := verifyStandardHeaders(res); err != nil {
		return err
	}
	return nil
}

// mainListParts - Entry point for the ListParts API test.
func mainListParts(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] Multipart (List-Parts):", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)
	bucket := validBuckets[0]
	object := multipartObjects[0]
	// Create a handcrafted ListObjectsPartsResult
	expectedList := listObjectPartsResult{
		Bucket:      bucket.Name,
		Key:         object.Key,
		UploadID:    object.UploadID,
		ObjectParts: objectParts,
	}
	// Create a new ListParts request.
	req, err := newListPartsReq(config, bucket.Name, object.Key, object.UploadID)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Execute the request.
	res, err := execRequest(req, config.Client)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Verify the response.
	if err := listPartsVerify(res, "200 OK", expectedList); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, err)
	return true
}
