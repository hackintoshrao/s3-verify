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
	"strconv"
	"strings"
	"time"

	"github.com/minio/s3verify/signv4"
)

// HeadObjectReq - an HTTP request for HEAD with no headers set.
var HeadObjectReq = &http.Request{
	Header: map[string][]string{
		// Set Content SHA with an empty for HEAD requests because no data is being uploaded.
		"X-Amz-Content-Sha256": {hex.EncodeToString(signv4.Sum256([]byte{}))},
	},
	Body:   nil, // No body is sent with HEAD requests.
	Method: "HEAD",
}

// NewHeadObjectReq - Create a new HTTP request for a HEAD object.
func NewHeadObjectReq(config ServerConfig, bucketName, objectName string) (*http.Request, error) {
	targetURL, err := makeTargetURL(config.Endpoint, bucketName, objectName, config.Region)
	if err != nil {
		return nil, err
	}
	// Fill request URL and sign.
	HeadObjectReq.URL = targetURL
	HeadObjectReq = signv4.SignV4(*HeadObjectReq, config.Access, config.Secret, config.Region)
	return HeadObjectReq, nil
}

// HeadObjectVerify - Verify that the response received matches what is expected.
func HeadObjectVerify(res *http.Response, expectedStatus string) error {
	if err := VerifyStatusHeadObject(res, expectedStatus); err != nil {
		return err
	}
	if err := VerifyHeaderHeadObject(res); err != nil {
		return err
	}
	if err := VerifyBodyHeadObject(res); err != nil {
		return err
	}
	return nil
}

// VerifyStatusHeadObject - Verify that the status received matches what is expected.
func VerifyStatusHeadObject(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus {
		err := fmt.Errorf("Unexpected Response Status Code: wanted %v, got %v", expectedStatus, res.Status)
		return err
	}
	return nil
}

// VerifyBodyHeadObject - Verify that the body recieved is empty.
func VerifyBodyHeadObject(res *http.Response) error {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if !bytes.Equal(body, []byte{}) {
		err := fmt.Errorf("Unexpected Body Recieved: HEAD requests should not return a body, but got back: %v", string(body))
		return err
	}
	return nil
}

// VerifyHeaderHeadObject - Verify that the header received matches what is exepected.
func VerifyHeaderHeadObject(res *http.Response) error {
	if err := verifyStandardHeaders(res); err != nil {
		return err
	}
	// TODO: add verification for ETag formation.
	return nil
}

// Test the HeadObject API with no header set.
func mainHeadObjectNoHeader(config ServerConfig, message string) error {
	// Spin scanBar
	scanBar(message)
	bucket := testBuckets[0]
	for _, object := range objects {
		// Create a new HEAD object with no headers.
		req, err := NewHeadObjectReq(config, bucket.Name, object.Key)
		if err != nil {
			return err
		}
		// Spin scanBar
		scanBar(message)
		res, err := ExecRequest(req, config.Client)
		if err != nil {
			return err
		}
		// Spin scanBar
		scanBar(message)

		// Verify the response.
		if err := HeadObjectVerify(res, "200 OK"); err != nil {
			return err
		}
		// If the verification is good then set the ETag, Size, and LastModified.
		// Remove the odd double quotes from ETag in the beginning and end.
		ETag := strings.TrimPrefix(res.Header.Get("ETag"), "\"")
		ETag = strings.TrimSuffix(ETag, "\"")
		object.ETag = ETag
		date, err := time.Parse(http.TimeFormat, res.Header.Get("Last-Modified")) // This will never error out because it has already been verified.
		if err != nil {
			return err
		}
		object.LastModified = date
		size, err := strconv.ParseInt(res.Header.Get("Content-Length"), 10, 64)
		if err != nil {
			return err
		}
		object.Size = size
		// Spin scanBar
		scanBar(message)
	}

	return nil
}
