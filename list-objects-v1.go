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

// newListObjectsV1Req - Create a new HTTP request for ListObjects V1.
func newListObjectsV1Req(config ServerConfig, bucketName string, parameters map[string]string) (Request, error) {
	// listObjectsV1Req - a new HTTP request for ListObjects V1.
	var listObjectsV1Req = Request{
		customHeader: http.Header{},
	}

	// Set the bucketName.
	listObjectsV1Req.bucketName = bucketName

	// Set the query values.
	urlValues := make(url.Values)
	for k, v := range parameters {
		urlValues.Set(k, v)
	}
	listObjectsV1Req.queryValues = urlValues

	// No body is sent with GET requests.
	reader := bytes.NewReader([]byte{})
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return Request{}, err
	}

	// Set the headers.
	listObjectsV1Req.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	listObjectsV1Req.customHeader.Set("User-Agent", appUserAgent)

	return listObjectsV1Req, nil
}

// listObjectsV1Verify - verify the response returned matches what is expected.
func listObjectsV1Verify(res *http.Response, expectedStatusCode int, expectedList listBucketResult) error {
	if err := verifyStatusListObjectsV1(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	if err := verifyBodyListObjectsV1(res.Body, expectedList); err != nil {
		return err
	}
	if err := verifyHeaderListObjectsV1(res.Header); err != nil {
		return err
	}
	return nil
}

// verifyStatusListObjectsV1 - verify the status returned matches what is expected.
func verifyStatusListObjectsV1(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Status Received: wanted %v, got %v", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// verifyBodyListObjectsV1 - verify the body returned matches what is expected.
func verifyBodyListObjectsV1(resBody io.Reader, expectedList listBucketResult) error {
	receivedList := listBucketResult{}
	err := xmlDecoder(resBody, &receivedList)
	if err != nil {
		return err
	}
	if receivedList.Name != expectedList.Name {
		err := fmt.Errorf("Unexpected Bucket Listed: wanted %v, got %v", expectedList.Name, receivedList.Name)
		return err
	}
	if len(receivedList.Contents) != len(expectedList.Contents) {
		err := fmt.Errorf("Incorrect Number of Objects Listed: wanted %v, got %v", len(expectedList.Contents), len(receivedList.Contents))
		return err
	}
	return nil
}

// verifyHeaderListObjectsV1 - verify the header returned matches what is expected.
func verifyHeaderListObjectsV1(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// mainListObjectsV1 - ListObjects V1 API test. This test is the same for both --prepared and non --prepared environments.
func mainListObjectsV1(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] ListObjects V1:", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)
	bucketName := unpreparedBuckets[0].Name
	objectInfo := ObjectInfos{}
	for _, object := range objects {
		objectInfo = append(objectInfo, *object)
	}
	// Test for listobjects with no extra parameters.
	expectedList := listBucketResult{
		Name:     bucketName, // Listing from the first bucket created that houses all objects.
		Contents: objectInfo, // The first bucket created will house all the objects created by the PUT object test.
		// Currently the ListObjects V1 test does not test with extra 'parameters' set:
		// Prefix
		// Max-Keys
		// Marker
		// Delimiter
	}
	// Test for listobjects with maxkeys parameter set.
	expectedListMaxKeys := listBucketResult{
		Name:     bucketName,
		Contents: objectInfo[:30], // Only return the first 30 objects.
		MaxKeys:  30,              // Only return the first 30 objects.
	}
	// Store the parameters to be set by the request.
	maxKeysMap := map[string]string{
		"max-keys": "30", // 30 objects.
	}
	// Create a new request.
	noParamReq, err := newListObjectsV1Req(config, bucketName, nil) // No extra parameters for the first test.
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Execute the request.
	noParamRes, err := config.execRequest("GET", noParamReq)
	if err != nil {
		printMessage(message, err)
		return false
	}
	defer closeResponse(noParamRes)
	// Spin scanBar
	scanBar(message)
	// Verify the response.
	if err := listObjectsV1Verify(noParamRes, http.StatusOK, expectedList); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Create a new request with max-keys set to 30.
	maxKeysReq, err := newListObjectsV1Req(config, bucketName, maxKeysMap) // MaxKeys set to 30.
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Execute the request.
	maxKeysRes, err := config.execRequest("GET", maxKeysReq)
	if err != nil {
		printMessage(message, err)
		return false
	}
	defer closeResponse(maxKeysRes)
	// Spin scanBar
	scanBar(message)
	// Verify the max-keys parameter is respected.
	if err := listObjectsV1Verify(maxKeysRes, http.StatusOK, expectedListMaxKeys); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, nil)
	return true
}
