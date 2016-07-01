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
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/minio/minio-go"
	"github.com/minio/s3verify/signv4"
)

// TODO: add support for additional GET object req. Headers: (Create individual funcs for each case)
// Range
// If-Modified-Since
// If-Unmodified-Since
// If-Match
// If-Non-Match
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

// TODO: Implement setups for different style of tests. See Above about headers. (individual funcs or switch?)
// OR can these tests all be run on the same object? Why not...

// GetObjectInit - Setup for a GET object test. NEED TO GIVE BACK A RANGE OF BYTES FOR SECOND TEST AS WELL.
func GetObjectInit(s3Client minio.Client, config ServerConfig) (bucketName string, objectName string, buf []byte, err error) {
	// Create random bucket and object names.
	bucketName = randString(60, rand.NewSource(time.Now().UnixNano()), "s3verify-get")
	objectName = randString(60, rand.NewSource(time.Now().UnixNano()), "s3verify-get")

	// Create random data more than 32K.
	buf = make([]byte, rand.Intn(1<<20)+32*1024)
	_, err = io.ReadFull(crand.Reader, buf)
	if err != nil {
		return bucketName, objectName, buf, err
	}
	// Create the test bucket and object.
	err = s3Client.MakeBucket(bucketName, config.Region)
	if err != nil {
		return bucketName, objectName, buf, err
	}
	// TODO: Do we need to test different content-types?
	_, err = s3Client.PutObject(bucketName, objectName, bytes.NewReader(buf), "binary/octet-stream")
	if err != nil {
		return bucketName, objectName, buf, err
	}
	return bucketName, objectName, buf, nil
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
func mainGetObjectNoHeader(config ServerConfig, s3Client minio.Client, message string) error {
	// TODO: should errors be returned to the top level or printed here.
	// Spin scanBar
	scanBar(message)
	// Set up an new Bucket and Object to GET
	bucketName, objectName, buf, err := GetObjectInit(s3Client, config)
	if err != nil {
		// Attempt a clean up of created object and bucket.
		if errC := cleanUpTest(s3Client, []string{bucketName}, []string{objectName}); errC != nil {
			return errC
		}
		return err
	}
	// Spin scanBar
	scanBar(message)

	// Create new GET object request...testing standard.
	req, err := NewGetObjectReq(config, bucketName, objectName)
	if err != nil {
		// Attempt a clean up of created object and bucket.
		if errC := cleanUpTest(s3Client, []string{bucketName}, []string{objectName}); errC != nil {
			return errC
		}
		return err
	}
	// Spin scanBar
	scanBar(message)

	// Execute the request.
	res, err := ExecRequest(req, config.Client)
	if err != nil {
		// Attempt a clean up of created object and bucket.
		if errC := cleanUpTest(s3Client, []string{bucketName}, []string{objectName}); errC != nil {
			return errC
		}
		return err
	}
	// Spin scanBar
	scanBar(message)

	// Verify the response...these checks do not check the header yet.
	if err := GetObjectVerify(res, buf, "200 OK"); err != nil {
		// Attempt a clean up of created object and bucket.
		if errC := cleanUpTest(s3Client, []string{bucketName}, []string{objectName}); errC != nil {
			return errC
		}
		return err
	}
	// Spin scanBar
	scanBar(message)

	// Clean up after the test.
	if err := cleanUpTest(s3Client, []string{bucketName}, []string{objectName}); err != nil {
		return err
	}
	// Spin scanBar
	scanBar(message)

	return nil
}
