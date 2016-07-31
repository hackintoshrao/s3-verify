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
	"net/url"
	"time"

	"github.com/minio/s3verify/signv4"
)

// newCopyObjectIfModifiedSinceReq - Create a new HTTP request for CopyObject with the x-amz-copy-source-if-modified-since header set.
func newCopyObjectIfModifiedSinceReq(config ServerConfig, sourceBucketName, sourceObjectName, destBucketName, destObjectName string, lastModified time.Time) (*http.Request, error) {
	// Create a new HTTP request for a CopyObject.
	var copyObjectIfModifiedSinceReq = &http.Request{
		Header: map[string][]string{
		// X-Amz-Content-Sha256 will be set dynamically.
		// x-amz-copy-source will be set dynamically.
		// x-amz-copy-source-if-modified-since will be set dynamically.
		},
		Method: "PUT",
	}
	// Set req URL and Header.
	targetURL, err := makeTargetURL(config.Endpoint, destBucketName, destObjectName, config.Region, nil)
	if err != nil {
		return nil, err
	}
	// Copying is done by the server so no body is sent in the request.
	reader := bytes.NewReader([]byte(""))
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return nil, err
	}
	copyObjectIfModifiedSinceReq.URL = targetURL
	copyObjectIfModifiedSinceReq.Header.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	copyObjectIfModifiedSinceReq.Header.Set("x-amz-copy-source", url.QueryEscape(sourceBucketName+"/"+sourceObjectName))
	copyObjectIfModifiedSinceReq.Header.Set("x-amz-copy-source-if-modified-since", lastModified.Format(http.TimeFormat))
	copyObjectIfModifiedSinceReq.Header.Set("User-Agent", appUserAgent)

	copyObjectIfModifiedSinceReq = signv4.SignV4(*copyObjectIfModifiedSinceReq, config.Access, config.Secret, config.Region)
	return copyObjectIfModifiedSinceReq, nil
}

// copyObjectIfModifiedSinceVerify - verify the response returned matches what is expected.
func copyObjectIfModifiedSinceVerify(res *http.Response, expectedStatus string, expectedError ErrorResponse) error {
	if err := verifyStatusCopyObjectIfModifiedSince(res, expectedStatus); err != nil {
		return err
	}
	if err := verifyBodyCopyObjectIfModifiedSince(res, expectedError); err != nil {
		return err
	}
	if err := verifyHeaderCopyObjectIfModifiedSince(res); err != nil {
		return err
	}
	return nil
}

// verifyStatusCopyObjectIfModifiedSince - verify the status returned matches what is expected.
func verifyStatusCopyObjectIfModifiedSince(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus {
		err := fmt.Errorf("Unexpected Status Received: wanted %v, got %v", expectedStatus, res.Status)
		return err
	}
	return nil
}

// verifyBodyCopyObjectIfModifiedSince - verify the body returned matches what is expected.
func verifyBodyCopyObjectIfModifiedSince(res *http.Response, expectedError ErrorResponse) error {
	if expectedError.Message != "" {
		resError := ErrorResponse{}
		err := xmlDecoder(res.Body, &resError)
		if err != nil {
			return err
		}
		if expectedError.Message != resError.Message {
			err := fmt.Errorf("Unexpected Error Message: wanted %v, got %v", expectedError.Message, resError.Message)
			return err
		}
		return nil
	} else {
		copyObjResult := copyObjectResult{}
		err := xmlDecoder(res.Body, &copyObjResult)
		if err != nil {
			body, errR := ioutil.ReadAll(res.Body)
			if errR != nil {
				return errR
			}
			err = fmt.Errorf("Unexpected Body Received: %v", string(body))
			return err
		}
		return nil
	}
}

// verifyHeaderCopyObjectIfModifiedSince - verify the header returned matches what is expected.
func verifyHeaderCopyObjectIfModifiedSince(res *http.Response) error {
	if err := verifyStandardHeaders(res); err != nil {
		return err
	}
	return nil
}

func mainCopyObjectIfModifiedSince(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] CopyObject (If-Modified-Since):", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)
	// Set a date in the past.
	pastDate, err := time.Parse(http.TimeFormat, "Thu, 01 Jan 1970 00:00:00 GMT")
	if err != nil {
		printMessage(message, err)
		return false
	}
	sourceBucketName := validBuckets[0].Name
	destBucketName := validBuckets[1].Name
	sourceObject := objects[0]
	destObject := &ObjectInfo{
		Key: sourceObject.Key + "if-modified-since",
	}
	copyObjects = append(copyObjects, destObject)
	expectedError := ErrorResponse{
		Code:    "PreconditionFailed",
		Message: "At least one of the pre-conditions you specified did not hold",
	}
	// Spin scanBar
	scanBar(message)
	// Create a new request with a valid date.
	req, err := newCopyObjectIfModifiedSinceReq(config, sourceBucketName, sourceObject.Key, destBucketName, destObject.Key, pastDate)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Execute the request.
	res, err := execRequest(req, config.Client, destBucketName, destObject.Key)
	if err != nil {
		printMessage(message, err)
		return false
	}
	defer closeResponse(res)
	// Spin scanBar
	scanBar(message)
	// Verify the response is valid.
	if err := copyObjectIfModifiedSinceVerify(res, "200 OK", ErrorResponse{}); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Create a new request with an invalid date.
	badReq, err := newCopyObjectIfModifiedSinceReq(config, sourceBucketName, sourceObject.Key, destBucketName, destObject.Key, time.Now().UTC().Add(2*time.Hour))
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Execute the request.
	badRes, err := execRequest(badReq, config.Client, destBucketName, destObject.Key)
	if err != nil {
		printMessage(message, err)
		return false
	}
	defer closeResponse(badRes)
	// Spin scanBar
	scanBar(message)
	// Verify the bad request fails the right way.
	if err := copyObjectIfModifiedSinceVerify(badRes, "412 Precondition Failed", expectedError); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	printMessage(message, err)
	return true
}
