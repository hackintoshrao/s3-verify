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
	"time"

	"github.com/minio/s3verify/signv4"
)

// newGetObjcetIfModifiedSinceReq - Create a new HTTP request to perform.
func newGetObjectIfModifiedSinceReq(config ServerConfig, bucketName, objectName string, lastModified time.Time) (*http.Request, error) {
	var getObjectIfModifiedReq = &http.Request{
		Header: map[string][]string{
			// Set Content SHA with empty body for GET requests because no data is being uploaded.
			"X-Amz-Content-Sha256": {hex.EncodeToString(signv4.Sum256([]byte{}))},
			"If-Modified-Since":    {""}, // To be added dynamically.
		},
		Body:   nil, // There is no body for GET requests.
		Method: "GET",
	}
	// Set req URL and Header.
	targetURL, err := makeTargetURL(config.Endpoint, bucketName, objectName, config.Region, nil)
	if err != nil {
		return nil, err
	}
	getObjectIfModifiedReq.Header.Set("If-Modified-Since", lastModified.Format(http.TimeFormat))
	getObjectIfModifiedReq.Header.Set("User-Agent", appUserAgent)

	// Fill request URL and sign.
	getObjectIfModifiedReq.URL = targetURL
	getObjectIfModifiedReq = signv4.SignV4(*getObjectIfModifiedReq, config.Access, config.Secret, config.Region)
	return getObjectIfModifiedReq, nil
}

// verifyGetObjectIfModifiedSince - Verify that the response matches what is expected.
func verifyGetObjectIfModifiedSince(res *http.Response, expectedBody []byte, expectedStatus string) error {
	if err := verifyHeaderGetObjectIfModifiedSince(res); err != nil {
		return err
	}
	if err := verifyBodyGetObjectIfModifiedSince(res, expectedBody); err != nil {
		return err
	}
	if err := verifyStatusGetObjectIfModifiedSince(res, expectedStatus); err != nil {
		return err
	}
	return nil
}

// verifyBodyGetObjectIfModifiedSince - Verify that the response body matches what is expected.
func verifyBodyGetObjectIfModifiedSince(res *http.Response, expectedBody []byte) error {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if !bytes.Equal(body, expectedBody) {
		err := fmt.Errorf("Unexpected Body Received: wanted %v, got %v", string(expectedBody), string(body))
		return err
	}
	return nil
}

// verifyStatusGetObjectIfModifiedSince - Verify that the response status matches what is expected.
func verifyStatusGetObjectIfModifiedSince(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus {
		err := fmt.Errorf("Unexpected Response Status Code: wanted %v, got %v", expectedStatus, res.Status)
		return err
	}
	return nil
}

// verifyHeaderGetObjectIfModifiedSince - Verify that the response header matches what is expected.
func verifyHeaderGetObjectIfModifiedSince(res *http.Response) error {
	if err := verifyStandardHeaders(res); err != nil {
		return err
	}
	return nil
}

// Test the compatibility of the GET object API when using the If-Modified-Since header.
func mainGetObjectIfModifiedSince(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] GetObject (If-Modified-Since):", curTest, globalTotalNumTest)
	// Set a date in the past.
	pastDate, err := time.Parse(http.TimeFormat, "Thu, 01 Jan 1970 00:00:00 GMT")
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	errCh := make(chan error, globalTotalNumTest)
	bucket := validBuckets[0]
	for _, object := range objects {
		// Spin scanBar
		scanBar(message)
		go func(objectKey string, objectLastModified time.Time, objectBody []byte) {
			// Create new GET object request.
			req, err := newGetObjectIfModifiedSinceReq(config, bucket.Name, objectKey, objectLastModified)
			if err != nil {
				errCh <- err
				return
			}
			// Perform the request.
			res, err := execRequest(req, config.Client, bucket.Name, objectKey)
			if err != nil {
				errCh <- err
				return
			}
			defer closeResponse(res)
			// Verify the response...these checks do not check the headers yet.
			if err := verifyGetObjectIfModifiedSince(res, []byte(""), "304 Not Modified"); err != nil {
				errCh <- err
				return
			}
			// Create an acceptable request.
			goodReq, err := newGetObjectIfModifiedSinceReq(config, bucket.Name, objectKey, pastDate)
			if err != nil {
				errCh <- err
				return
			}
			// Execute the response that should give back a body.
			goodRes, err := execRequest(goodReq, config.Client, bucket.Name, objectKey)
			if err != nil {
				errCh <- err
				return
			}
			defer closeResponse(goodRes)
			// Verify that the past date gives back the data.
			if err := verifyGetObjectIfModifiedSince(goodRes, objectBody, "200 OK"); err != nil {
				errCh <- err
				return
			}
			errCh <- nil
		}(object.Key, object.LastModified, object.Body)
	}
	count := len(objects)
	for count > 0 {
		count--
		// Spin scanBar
		scanBar(message)
		err, ok := <-errCh
		if !ok {
			return false
		}
		if err != nil {
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
