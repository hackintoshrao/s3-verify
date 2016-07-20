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

	"github.com/minio/s3verify/signv4"
)

// newGetObjectReq - Create a new HTTP requests to perform.
func newGetObjectReq(config ServerConfig, bucketName, objectName string) (*http.Request, error) {
	// getObjectReq - a new HTTP request for a GET object.
	var getObjectReq = &http.Request{
		Header: map[string][]string{
			// Set Content SHA with empty body for GET requests because no data is being uploaded.
			"X-Amz-Content-Sha256": {hex.EncodeToString(signv4.Sum256([]byte{}))},
		},
		Body:   nil, // There is no body for GET requests.
		Method: "GET",
	}
	targetURL, err := makeTargetURL(config.Endpoint, bucketName, objectName, config.Region, nil)
	if err != nil {
		return nil, err
	}
	// Fill request URL and sign.
	getObjectReq.URL = targetURL
	getObjectReq = signv4.SignV4(*getObjectReq, config.Access, config.Secret, config.Region)
	return getObjectReq, nil
}

// TODO: These checks only verify correctly formatted requests. There is no request that is made to fail / check failure yet.

// getObjectVerify - Check a Response's Status, Headers, and Body for AWS S3 compliance.
func getObjectVerify(res *http.Response, expectedBody []byte, expectedStatus string) error {
	if err := verifyHeaderGetObject(res); err != nil {
		return err
	}
	if err := verifyStatusGetObject(res, expectedStatus); err != nil {
		return err
	}
	if err := verifyBodyGetObject(res, expectedBody); err != nil {
		return err
	}
	return nil
}

// verifyHeaderGetObject - Verify that the header returned matches what is expected.
func verifyHeaderGetObject(res *http.Response) error {
	if err := verifyStandardHeaders(res); err != nil {
		return err
	}
	return nil
}

// verifyBodyGetObject - Verify that the body returned matches what is expected.
func verifyBodyGetObject(res *http.Response, expectedBody []byte) error {
	body, err := ioutil.ReadAll(res.Body)
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
func verifyStatusGetObject(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus {
		err := fmt.Errorf("Unexpected Response Status Code: wanted %v, got %v", expectedStatus, res.Status)
		return err
	}
	return nil
}

// Test a GET object request with no special headers set.
func mainGetObject(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] GetObject:", curTest, globalTotalNumTest)
	// TODO: should errors be returned to the top level or printed here.
	bucket := validBuckets[0]
	for _, object := range objects {
		// Spin scanBar
		scanBar(message)
		// Create new GET object request.
		req, err := newGetObjectReq(config, bucket.Name, object.Key)
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
		if err := getObjectVerify(res, object.Body, "200 OK"); err != nil {
			printMessage(message, err)
			return false
		}
		// Spin scanBar
		scanBar(message)
	}
	printMessage(message, nil)
	return true
}
