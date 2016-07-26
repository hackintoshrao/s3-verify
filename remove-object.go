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

// newRemoveObjectReq - Create a new DELETE object HTTP request.
func newRemoveObjectReq(config ServerConfig, bucketName, objectName string) (*http.Request, error) {
	var removeObjectReq = &http.Request{
		Header: map[string][]string{
			// Set Content SHA with empty body for DELETE requests because no data is being uploaded.
			"X-Amz-Content-Sha256": {hex.EncodeToString(signv4.Sum256([]byte{}))},
		},
		Body:   nil, // There is no body for DELETE requests.
		Method: "DELETE",
	}
	targetURL, err := makeTargetURL(config.Endpoint, bucketName, objectName, config.Region, nil)
	if err != nil {
		return nil, err
	}
	removeObjectReq.URL = targetURL
	removeObjectReq = signv4.SignV4(*removeObjectReq, config.Access, config.Secret, config.Region)
	return removeObjectReq, nil
}

// removeObjectVerify - Verify that the response returned matches what is expected.
func removeObjectVerify(res *http.Response, expectedStatus string) error {
	if err := verifyHeaderRemoveObject(res); err != nil {
		return err
	}
	return nil
}

// verifyHeaderRemoveObject - Verify that header returned matches what is expected.
func verifyHeaderRemoveObject(res *http.Response) error {
	if err := verifyStandardHeaders(res); err != nil {
		return err
	}
	return nil
}

// verifyBodyRemoveObject - Verify that the body returned is empty.
func verifyBodyRemoveObject(res *http.Response) error {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if !bytes.Equal(body, []byte{}) {
		err := fmt.Errorf("Unexpected Body Received: %v", string(body))
		return err
	}
	return nil
}

// verifyStatusRemoveObject - Verify that the status returned matches what is expected.
func verifyStatusRemoveObject(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus {
		err := fmt.Errorf("Unexpected Status Received: wanted %v, got %v", expectedStatus, res.Status)
		return err
	}
	return nil
}

// mainRemoveObjectExists - Entry point for the RemoveObject API test when object exists.
func mainRemoveObjectExists(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%d/%d] RemoveObject:", curTest, globalTotalNumTest)
	errCh := make(chan error, 1)
	// Spin scanBar
	scanBar(message)
	for _, bucket := range validBuckets {
		for _, object := range objects {
			// Spin scanBar
			scanBar(message)
			go func(bucketName, objectKey string) {
				// Create a new request.
				req, err := newRemoveObjectReq(config, bucketName, objectKey)
				if err != nil {
					errCh <- err
					return
				}
				// Execute the request.
				res, err := execRequest(req, config.Client)
				if err != nil {
					errCh <- err
					return
				}
				// Verify the response.
				if err := removeObjectVerify(res, "200 OK"); err != nil {
					errCh <- err
					return
				}
				errCh <- nil
			}(bucket.Name, object.Key)
			// Spin scanBar
			scanBar(message)

		}
		count := len(objects)
		for count > 0 {
			count--
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
		for _, object := range copyObjects {
			// Spin scanBar
			scanBar(message)
			go func(bucketName, objectKey string) {
				// Create a new request.
				req, err := newRemoveObjectReq(config, bucketName, objectKey)
				if err != nil {
					errCh <- err
					return
				}
				// Execute the request.
				res, err := execRequest(req, config.Client)
				if err != nil {
					errCh <- err
					return
				}
				// Verify the response.
				if err := removeObjectVerify(res, "200 OK"); err != nil {
					errCh <- err
					return
				}
				errCh <- nil
			}(bucket.Name, object.Key)
			// Spin scanBar
			scanBar(message)
		}
		count = len(copyObjects)
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
		for _, object := range multipartObjects {
			// Spin scanBar
			scanBar(message)
			go func(bucketName, objectKey string) {
				// Create a new request.
				req, err := newRemoveObjectReq(config, bucketName, objectKey)
				if err != nil {
					errCh <- err
					return
				}
				// Execute the request.
				res, err := execRequest(req, config.Client)
				if err != nil {
					errCh <- err
					return
				}
				// Verify the response.
				if err := removeObjectVerify(res, "200 OK"); err != nil {
					errCh <- err
					return
				}
				errCh <- nil
			}(bucket.Name, object.Key)
			// Spin scanBar
			scanBar(message)
		}
		count = len(multipartObjects)
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
	}
	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, nil)
	return true
}
