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
	"encoding/hex"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/minio/s3verify/signv4"
)

// newGetObjectRangeReq - Create a new GET object range request.
func newGetObjectRangeReq(config ServerConfig, bucketName, objectName string, startRange, endRange int64) (*http.Request, error) {
	// getObjectRangeReq - a new HTTP request for a GET object with a specific range request.
	var getObjectRangeReq = &http.Request{
		Header: map[string][]string{
			// Set Content SHA with empty body for GET requests because no data is being uploaded.
			"X-Amz-Content-Sha256": {hex.EncodeToString(signv4.Sum256([]byte{}))},
		},
		Body:   nil, // There is no body sent for GET requests.
		Method: "GET",
	}
	targetURL, err := makeTargetURL(config.Endpoint, bucketName, objectName, config.Region)
	if err != nil {
		return nil, err
	}
	// Fill the request URL.
	getObjectRangeReq.URL = targetURL
	// Fill the range request.
	getObjectRangeReq.Header.Set("Range", "bytes="+strconv.FormatInt(startRange, 10)+"-"+strconv.FormatInt(endRange, 10))
	// Sign the request.
	getObjectRangeReq = signv4.SignV4(*getObjectRangeReq, config.Access, config.Secret, config.Region)
	return getObjectRangeReq, nil
}

// Test a GET object request with a range header set.
func mainGetObjectRange(config ServerConfig, message string) error {
	bucket := testBuckets[0]
	rand.Seed(time.Now().UnixNano())
	for _, object := range objects {
		startRange := rand.Int63n(object.Size)
		endRange := rand.Int63n(int64(object.Size-startRange)) + startRange
		// Spin scanBar
		scanBar(message)
		// Create new GET object range request...testing range.
		req, err := newGetObjectRangeReq(config, bucket.Name, object.Key, startRange, endRange)
		if err != nil {
			return err
		}
		// Spin scanBar
		scanBar(message)
		// Execute the request.
		res, err := execRequest(req, config.Client)
		if err != nil {
			return err
		}
		// Spin scanBar
		scanBar(message)
		bufRange := object.Body[startRange : endRange+1]
		// Verify the response...these checks do not check the headers yet.
		if err := getObjectVerify(res, bufRange, "206 Partial Content"); err != nil {
			return err
		}
		// Spin scanBar
		scanBar(message)
	}
	return nil
}
