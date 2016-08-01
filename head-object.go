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
	// Set the req URL and Header.
	targetURL, err := makeTargetURL(config.Endpoint, bucketName, objectName, config.Region, nil)
	if err != nil {
		return nil, err
	}
	headObjectReq.Header.Set("User-Agent", appUserAgent)
	// Fill request URL and sign.
	headObjectReq.URL = targetURL
	headObjectReq = signv4.SignV4(*headObjectReq, config.Access, config.Secret, config.Region)
	return headObjectReq, nil
}

// headObjectVerify - Verify that the response received matches what is expected.
func headObjectVerify(res *http.Response, expectedStatusCode int) error {
	if err := verifyStatusHeadObject(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	if err := verifyHeaderHeadObject(res.Header); err != nil {
		return err
	}
	if err := verifyBodyHeadObject(res.Body); err != nil {
		return err
	}
	return nil
}

// verifyStatusHeadObject - Verify that the status received matches what is expected.
func verifyStatusHeadObject(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Response Status Code: wanted %v, got %v", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// verifyBodyHeadObject - Verify that the body recieved is empty.
func verifyBodyHeadObject(resBody io.Reader) error {
	body, err := ioutil.ReadAll(resBody)
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
func verifyHeaderHeadObject(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	// TODO: add verification for ETag formation.
	return nil
}

// mainHeadObject - Entry point for the HeadObject API test with no header set.
func mainHeadObject(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] HeadObject:", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)
	bucket := validBuckets[0]
	objInfoCh := make(chan objectInfoChannel, globalRequestPoolSize)
	for i, object := range objects {
		// Spin scanBar
		scanBar(message)
		go func(objectKey string, cur int) {
			// Create a new HEAD object with no headers.
			req, err := newHeadObjectReq(config, bucket.Name, objectKey)
			if err != nil {
				objInfoCh <- objectInfoChannel{objInfo: ObjectInfo{}, index: cur, err: err}
				return
			}
			// Execute the request.
			res, err := execRequest(req, config.Client, bucket.Name, objectKey)
			if err != nil {
				objInfoCh <- objectInfoChannel{objInfo: ObjectInfo{}, index: cur, err: err}
				return
			}
			defer closeResponse(res)
			// Verify the response.
			if err := headObjectVerify(res, http.StatusOK); err != nil {
				objInfoCh <- objectInfoChannel{objInfo: ObjectInfo{}, index: cur, err: err}
				return
			}
			// If the verification is valid then set the ETag, Size, and LastModified.
			// Remove the odd double quotes from ETag in the beginning and end.
			eTag := strings.TrimPrefix(res.Header.Get("ETag"), "\"")
			eTag = strings.TrimSuffix(eTag, "\"")
			date, err := time.Parse(http.TimeFormat, res.Header.Get("Last-Modified")) // This will never error out because it has already been verified.
			if err != nil {
				objInfoCh <- objectInfoChannel{objInfo: ObjectInfo{}, index: cur, err: err}
				return
			}
			size, err := strconv.ParseInt(res.Header.Get("Content-Length"), 10, 64)
			if err != nil {
				objInfoCh <- objectInfoChannel{objInfo: ObjectInfo{}, index: cur, err: err}
				return
			}
			objInfoCh <- objectInfoChannel{ // Send back the valid data through the channel to be set.
				objInfo: ObjectInfo{
					Key:          objectKey,
					Size:         size,
					ETag:         eTag,
					LastModified: date,
				},
				index: cur,
				err:   nil,
			}
		}(object.Key, i)
		// Spin scanBar
		scanBar(message)
	}
	count := len(objects)
	for count > 0 {
		count--
		// Spin scanBar
		scanBar(message)
		objInfoErr, ok := <-objInfoCh
		if !ok {
			return false
		}
		if objInfoErr.err != nil {
			printMessage(message, objInfoErr.err)
			return false
		}
		// Retrieve the object returned by the specific goroutine.
		object := objects[objInfoErr.index]
		// Set the objects metadata if the test did not fail.
		object.Size = objInfoErr.objInfo.Size
		object.ETag = objInfoErr.objInfo.ETag
		object.LastModified = objInfoErr.objInfo.LastModified
		// Spin scanBar
		scanBar(message)
	}
	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, nil)
	return true
}
