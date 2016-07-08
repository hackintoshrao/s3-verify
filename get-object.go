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

// GetObjectReq - a new HTTP request for a GET object.
var GetObjectReq = &http.Request{
	Header: map[string][]string{
		// Set Content SHA with empty body for GET requests because no data is being uploaded.
		"X-Amz-Content-Sha256": {hex.EncodeToString(signv4.Sum256([]byte{}))},
	},
	Body:   nil, // There is no body for GET requests.
	Method: "GET",
}

// NewGetObjectReq - Create a new HTTP requests to perform.
func NewGetObjectReq(config ServerConfig, bucketName, objectName string) (*http.Request, error) {
	targetURL, err := makeTargetURL(config.Endpoint, bucketName, objectName, config.Region)
	if err != nil {
		return nil, err
	}
	// Fill request URL and sign.
	GetObjectReq.URL = targetURL
	GetObjectReq = signv4.SignV4(*GetObjectReq, config.Access, config.Secret, config.Region)
	return GetObjectReq, nil
}

// TODO: These checks only verify correctly formatted requests. There is no request that is made to fail / check failure yet.

// GetObjectVerify - Check a Response's Status, Headers, and Body for AWS S3 compliance.
func GetObjectVerify(res *http.Response, expectedBody []byte, expectedStatus string) error {
	if err := VerifyHeaderGetObject(res); err != nil {
		return err
	}
	if err := VerifyStatusGetObject(res, expectedStatus); err != nil {
		return err
	}
	if err := VerifyBodyGetObject(res, expectedBody); err != nil {
		return err
	}
	return nil
}

// VerifyHeaderGetObject - Verify that the header returned matches what is expected.
func VerifyHeaderGetObject(res *http.Response) error {
	if err := verifyStandardHeaders(res); err != nil {
		return err
	}
	return nil
}

// VerifyBodyGetObject - Verify that the body returned matches what is expected.
func VerifyBodyGetObject(res *http.Response, expectedBody []byte) error {
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

// VerifyStatusGetObject - Verify that the status returned matches what is expected.
func VerifyStatusGetObject(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus {
		err := fmt.Errorf("Unexpected Response Status Code: wanted %v, got %v", expectedStatus, res.Status)
		return err
	}
	return nil
}

// Test a GET object request with no special headers set.
func mainGetObjectNoHeader(config ServerConfig, message string) error {
	// TODO: should errors be returned to the top level or printed here.
	bucket := testBuckets[0]
	for _, object := range objects {
		// Spin scanBar
		scanBar(message)
		// Create new GET object request.
		req, err := NewGetObjectReq(config, bucket.Name, object.Key)
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

		// Verify the response...these checks do not check the header yet.
		if err := GetObjectVerify(res, object.Body, "200 OK"); err != nil {
			return err
		}
		// Spin scanBar
		scanBar(message)
	}
	return nil
}
