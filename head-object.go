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

// newHeadObjectReq - Create a new HTTP request for a HEAD object.
func newHeadObjectReq(config ServerConfig, bucketName, objectName string) (*http.Request, error) {
	// headObjectReq - an HTTP request for HEAD with no headers set.
	var headObjectReq = &http.Request{
		Header: map[string][]string{
			// Set Content SHA with an empty for HEAD requests because no data is being uploaded.
			"X-Amz-Content-Sha256": {hex.EncodeToString(signv4.Sum256([]byte{}))},
		},
		Body:   nil, // No body is sent with HEAD requests.
		Method: "HEAD",
	}
	targetURL, err := makeTargetURL(config.Endpoint, bucketName, objectName, config.Region, nil)
	if err != nil {
		return nil, err
	}
	// Fill request URL and sign.
	headObjectReq.URL = targetURL
	headObjectReq = signv4.SignV4(*headObjectReq, config.Access, config.Secret, config.Region)
	return headObjectReq, nil
}

// headObjectVerify - Verify that the response received matches what is expected.
func headObjectVerify(res *http.Response, expectedStatus string) error {
	if err := verifyStatusHeadObject(res, expectedStatus); err != nil {
		return err
	}
	if err := verifyHeaderHeadObject(res); err != nil {
		return err
	}
	if err := verifyBodyHeadObject(res); err != nil {
		return err
	}
	return nil
}

// verifyStatusHeadObject - Verify that the status received matches what is expected.
func verifyStatusHeadObject(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus {
		err := fmt.Errorf("Unexpected Response Status Code: wanted %v, got %v", expectedStatus, res.Status)
		return err
	}
	return nil
}

// verifyBodyHeadObject - Verify that the body recieved is empty.
func verifyBodyHeadObject(res *http.Response) error {
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

// verifyHeaderHeadObject - Verify that the header received matches what is exepected.
func verifyHeaderHeadObject(res *http.Response) error {
	if err := verifyStandardHeaders(res); err != nil {
		return err
	}
	// TODO: add verification for ETag formation.
	return nil
}

// Test the HeadObject API with no header set.
func mainHeadObject(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] HeadObject:", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)
	bucket := validBuckets[0]
	for _, object := range objects {
		// Create a new HEAD object with no headers.
		req, err := newHeadObjectReq(config, bucket.Name, object.Key)
		if err != nil {
			printMessage(message, err)
			return false
		}
		// Spin scanBar
		scanBar(message)
		res, err := execRequest(req, config.Client)
		if err != nil {
			printMessage(message, err)
			return false
		}
		// Spin scanBar
		scanBar(message)

		// Verify the response.
		if err := headObjectVerify(res, "200 OK"); err != nil {
			printMessage(message, err)
			return false
		}
		// If the verification is good then set the ETag, Size, and LastModified.
		// Remove the odd double quotes from ETag in the beginning and end.
		ETag := strings.TrimPrefix(res.Header.Get("ETag"), "\"")
		ETag = strings.TrimSuffix(ETag, "\"")
		object.ETag = ETag
		date, err := time.Parse(http.TimeFormat, res.Header.Get("Last-Modified")) // This will never error out because it has already been verified.
		if err != nil {
			printMessage(message, err)
			return false
		}
		object.LastModified = date
		size, err := strconv.ParseInt(res.Header.Get("Content-Length"), 10, 64)
		if err != nil {
			printMessage(message, err)
			return false
		}
		object.Size = size
		// Spin scanBar
		scanBar(message)
	}
	printMessage(message, nil)
	return true
}
