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
	"io/ioutil"
	"net/http"
)

// newGetObjectReq - Create a new HTTP requests to perform.
func newGetObjectReq(config ServerConfig, bucketName, objectName string) (Request, error) {
	// getObjectReq - a new HTTP request for a GET object.
	var getObjectReq = Request{
		customHeader: http.Header{},
	}

	// Set the bucketName and objectName.
	getObjectReq.bucketName = bucketName
	getObjectReq.objectName = objectName

	reader := bytes.NewReader([]byte{}) // Compute hash using empty body because GET requests do not send a body.
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return Request{}, err
	}

	// Set the headers.
	getObjectReq.customHeader.Set("User-Agent", appUserAgent)
	getObjectReq.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))

	return getObjectReq, nil
}

// TODO: These checks only verify correctly formatted requests. There is no request that is made to fail / check failure yet.

// getObjectVerify - Check a Response's Status, Headers, and Body for AWS S3 compliance.
func getObjectVerify(res *http.Response, expectedBody []byte, expectedStatusCode int) error {
	if err := verifyHeaderGetObject(res.Header); err != nil {
		return err
	}
	if err := verifyStatusGetObject(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	if err := verifyBodyGetObject(res.Body, expectedBody); err != nil {
		return err
	}
	return nil
}

// verifyHeaderGetObject - Verify that the header returned matches what is expected.
func verifyHeaderGetObject(header map[string][]string) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// verifyBodyGetObject - Verify that the body returned matches what is expected.
func verifyBodyGetObject(resBody io.Reader, expectedBody []byte) error {
	body, err := ioutil.ReadAll(resBody)
	if err != nil {
		return err
	}
	// Compare what was created to be uploaded and what is contained in the response body.
	if !bytes.Equal(body, expectedBody) {
		err := fmt.Errorf("Unexpected Body Recieved: wanted %v, got %v", string(expectedBody), string(body))
		return err
	}
	return nil
}

// verifyStatusGetObject - Verify that the status returned matches what is expected.
func verifyStatusGetObject(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Response Status Code: wanted %v, got %v", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// testGetObject - test a get object request.
func testGetObject(config ServerConfig, curTest int, bucketName string, testObjects []*ObjectInfo) bool {
	message := fmt.Sprintf("[%02d/%d] GetObject:", curTest, globalTotalNumTest)
	// Use the bucket created in the mainPutBucketPrepared Test.
	for _, object := range testObjects {
		// Spin scanBar
		scanBar(message)
		// Create new GET object request.
		req, err := newGetObjectReq(config, bucketName, object.Key)
		if err != nil {
			printMessage(message, err)
			return false
		}
		// Execute the request.
		res, err := config.execRequest("GET", req)
		if err != nil {
			printMessage(message, err)
			return false
		}
		defer closeResponse(res)
		// Verify the response.
		if err := getObjectVerify(res, object.Body, http.StatusOK); err != nil {
			printMessage(message, err)
			return false
		}
		// Spin scanBar
		scanBar(message)
	}
	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, nil)
	return true
}

// mainGetObjectUnPrepared - entry point for the GetObject test if --prepare was not used.
func mainGetObjectUnPrepared(config ServerConfig, curTest int) bool {
	bucketName := unpreparedBuckets[0].Name
	return testGetObject(config, curTest, bucketName, objects)
}

// mainGetObjectPrepared - entry point for the GetObject test if --prepare was used.
func mainGetObjectPrepared(config ServerConfig, curTest int) bool {
	bucketName := s3verifyBuckets[0].Name
	return testGetObject(config, curTest, bucketName, s3verifyObjects)
}
