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
	"io"
	"net/http"
	"net/url"
)

// newListObjectsV2Req - Create a new HTTP request for ListObjects V2 API.
func newListObjectsV2Req(config ServerConfig, bucketName string) (Request, error) {
	// listObjectsV2Req - a new HTTP request for ListObjects V2 API.
	var listObjectsV2Req = Request{
		customHeader: http.Header{},
	}

	// Set the bucketName.
	listObjectsV2Req.bucketName = bucketName

	// Set URL query values.
	urlValues := make(url.Values)
	urlValues.Set("list-type", "2")
	listObjectsV2Req.queryValues = urlValues

	// No body is sent with GET requests.
	reader := bytes.NewReader([]byte{})
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return Request{}, err
	}

	listObjectsV2Req.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	listObjectsV2Req.customHeader.Set("User-Agent", appUserAgent)

	return listObjectsV2Req, nil
}

// listObjectsV2Verify - verify the response returned matches what is expected.
func listObjectsV2Verify(res *http.Response, expectedStatusCode int, expectedList listBucketV2Result) error {
	if err := verifyStatusListObjectsV2(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	if err := verifyBodyListObjectsV2(res.Body, expectedList); err != nil {
		return err
	}
	if err := verifyHeaderListObjectsV2(res.Header); err != nil {
		return err
	}
	return nil
}

// verifyHeaderListObjectsV2 - verify the heaer returned matches what is expected.
func verifyHeaderListObjectsV2(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// verifyStatusListObjectsV2 - verify the status returned matches what is expected.
func verifyStatusListObjectsV2(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Status Received: wanted %v, got %v", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// verifyBodyListObjectsV2 - verify the objects listed match what is expected.
func verifyBodyListObjectsV2(resBody io.Reader, expectedList listBucketV2Result) error {
	receivedList := listBucketV2Result{}
	if err := xmlDecoder(resBody, &receivedList); err != nil {
		return err
	}
	if receivedList.Name != expectedList.Name {
		err := fmt.Errorf("Unexpected Bucket Listed: wanted %v, got %v", expectedList.Name, receivedList.Name)
		return err
	}
	if len(receivedList.Contents)+len(receivedList.CommonPrefixes) != len(expectedList.Contents)+len(expectedList.CommonPrefixes) {
		err := fmt.Errorf("Unexpected Number of Objects Listed: wanted %d, got %d", len(expectedList.Contents)+len(expectedList.CommonPrefixes), len(receivedList.CommonPrefixes)+len(receivedList.Contents))
		return err
	}
	return nil
}

// mainListObjectsV2 - Entry point for the ListObjects V2 API test. This test is the same for --prepared environments and non --prepared.
func mainListObjectsV2(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] ListObjects V2:", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)
	bucketName := unpreparedBuckets[0].Name
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
	res, err := config.execRequest("GET", req)
	if err != nil {
		printMessage(message, err)
		return false
	}
	defer closeResponse(res)
	// Verify the response.
	if err := listObjectsV2Verify(res, http.StatusOK, expectedList); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, nil)
	return true
}
