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

// Add support for testing GET object requests including range headers.

// GetObjectRangeReq - a new HTTP request for a GET object with a specific range request.
var GetObjectRangeReq = &http.Request{
	Header: map[string][]string{
		// Set Content SHA with empty body for GET requests because no data is being uploaded.
		"X-Amz-Content-Sha256": {hex.EncodeToString(signv4.Sum256([]byte{}))},
		"Range":                {""}, // To be filled later.
	},
	Body:   nil, // There is no body sent for GET requests.
	Method: "GET",
}

// NewGetObjectRangeReq - create a new GET object range request.
func NewGetObjectRangeReq(config ServerConfig, bucketName, objectName string, startRange, endRange int64) (*http.Request, error) {
	targetURL, err := makeTargetURL(config.Endpoint, bucketName, objectName, config.Region)
	if err != nil {
		return nil, err
	}
	// Fill the request URL.
	GetObjectRangeReq.URL = targetURL
	// Fill the range request.
	GetObjectRangeReq.Header.Set("Range", "bytes="+strconv.FormatInt(startRange, 10)+"-"+strconv.FormatInt(endRange, 10))
	// Sign the request.
	GetObjectRangeReq = signv4.SignV4(*GetObjectRangeReq, config.Access, config.Secret, config.Region)
	return GetObjectRangeReq, nil
}

// Test a GET object request with a range header set.
func mainGetObjectRange(config ServerConfig, message string) error {
	// TODO: should errors be returned to the top level or printed here.
	bucket := testBuckets[0]
	rand.Seed(time.Now().UnixNano())
	for _, object := range objects {
		startRange := rand.Int63n(object.Size)
		endRange := rand.Int63n(int64(object.Size-startRange)) + startRange
		// Spin scanBar
		scanBar(message)
		// Create new GET object range request...testing range.
		req, err := NewGetObjectRangeReq(config, bucket.Name, object.Key, startRange, endRange)
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
		bufRange := object.Body[startRange : endRange+1]
		// Verify the response...these checks do not check the headers yet.
		if err := GetObjectVerify(res, bufRange, "206 Partial Content"); err != nil {
			return err
		}
		// Spin scanBar
		scanBar(message)
	}
	return nil
}
