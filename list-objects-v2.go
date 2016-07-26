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

// newListObjectsV2Req - Create a new HTTP request for ListObjects V2 API.
func newListObjectsV2Req(config ServerConfig, bucketName string) (*http.Request, error) {
	// listObjectsV2Req - a new HTTP request for ListObjects V2 API.
	var listObjectsV2Req = &http.Request{
		Header: map[string][]string{
		// X-Amz-Content-Sha256 will be set below.
		},
		Body:   nil, // No body is sent with GET requests.
		Method: "GET",
	}
	// Set URL query values
	urlValues := make(url.Values)
	urlValues.Set("list-type", "2")
	targetURL, err := makeTargetURL(config.Endpoint, bucketName, "", config.Region, urlValues)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader([]byte{})
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return nil, err
	}
	listObjectsV2Req.Header.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	listObjectsV2Req.URL = targetURL
	listObjectsV2Req = signv4.SignV4(*listObjectsV2Req, config.Access, config.Secret, config.Region)

	return listObjectsV2Req, nil
}

// listObjectsV2Verify - verify the response returned matches what is expected.
func listObjectsV2Verify(res *http.Response, expectedStatus string, expectedList listBucketV2Result) error {
	if err := verifyStatusListObjectsV2(res, expectedStatus); err != nil {
		return err
	}
	if err := verifyBodyListObjectsV2(res, expectedList); err != nil {
		return err
	}
	if err := verifyHeaderListObjectsV2(res); err != nil {
		return err
	}
	return nil
}

// verifyHeaderListObjectsV2 - verify the heaer returned matches what is expected.
func verifyHeaderListObjectsV2(res *http.Response) error {
	if err := verifyStandardHeaders(res); err != nil {
		return err
	}
	return nil
}

// verifyStatusListObjectsV2 - verify the status returned matches what is expected.
func verifyStatusListObjectsV2(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus {
		err := fmt.Errorf("Unexpected Status Received: wanted %v, got %v", expectedStatus, res.Status)
		return err
	}
	return nil
}

// verifyBodyListObjectsV2 - verify the objects listed match what is expected.
func verifyBodyListObjectsV2(res *http.Response, expectedList listBucketV2Result) error {
	receivedList := listBucketV2Result{}
	if err := xmlDecoder(res.Body, &receivedList); err != nil {
		return err
	}
	if receivedList.Name != expectedList.Name {
		err := fmt.Errorf("Unexpected Bucket Listed: wanted %v, got %v", expectedList.Name, receivedList.Name)
		return err
	}
	listedObjects := 0
	for _, receivedObject := range receivedList.Contents {
		for _, expectedObject := range expectedList.Contents {
			if receivedObject.Key == expectedObject.Key &&
				receivedObject.ETag == "\""+expectedObject.ETag+"\"" &&
				receivedObject.Size == expectedObject.Size {
				listedObjects++
			}
		}
	}
	if listedObjects != len(objects) {
		err := fmt.Errorf("Unexpected Number of Objects Listed: wanted %d, got %d", len(objects), listedObjects)
		return err
	}
	return nil
}

// mainListObjectsV2 - Entry point for the ListObjects V2 API test.
func mainListObjectsV2(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] ListObjects V2:", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)
	bucketName := validBuckets[0].Name
	objectInfo := []ObjectInfo{}
	for _, object := range objects {
		objectInfo = append(objectInfo, *object)
	}
	expectedList := listBucketV2Result{
		Name:     bucketName, // List only from the first bucket created because that is the bucket holding the objects.
		Contents: objectInfo,
	}
	// Create a new request.
	req, err := newListObjectsV2Req(config, bucketName)
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
	// Verify the response.
	if err := listObjectsV2Verify(res, "200 OK", expectedList); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, nil)
	return true
}
