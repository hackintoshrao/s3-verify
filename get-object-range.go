/*
 * Minio S3Verify Library for Amazon S3 Compatible Cloud Storage (C) 2016 Minio, Inc.
 *
	} Licensed under the Apache License, Version 2.0 (the "License");
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
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/minio/minio-go"
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
func NewGetObjectRangeReq(config ServerConfig, bucketName, objectName string, startRange, endRange int) (*http.Request, error) {
	targetURL, err := url.Parse(config.Endpoint)
	if err != nil {
		return nil, err
	}
	targetURL.Path = "/" + bucketName + "/" + objectName // Default to path style.
	if isVirtualStyleHostSupported(targetURL) {
		targetURL.Path = "/" + objectName
		targetURL.Host = bucketName + "." + targetURL.Host
	}
	// Fill the request URL.
	GetObjectRangeReq.URL = targetURL
	// Fill the range request.
	GetObjectRangeReq.Header["Range"] = []string{"bytes=" + strconv.Itoa(startRange) + "-" + strconv.Itoa(endRange)}
	// Sign the request.
	GetObjectRangeReq = signv4.SignV4(*GetObjectRangeReq, config.Access, config.Secret, config.Region)
	return GetObjectRangeReq, nil
}

// GetObjectRangeInit - Set up the GET object range test.
func GetObjectRangeInit(config ServerConfig) (bucketName, objectName string, startRange, endRange int, bufRange []byte, err error) {
	// Create random bucket and object names prefixed by s3verify-get.
	bucketName = randString(60, rand.NewSource(time.Now().UnixNano()), "s3verify-get")
	objectName = randString(60, rand.NewSource(time.Now().UnixNano()), "s3verify-get")
	// Generate random start and end ranges.
	src := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(src)
	startRange = r1.Intn(1 << 10)
	endRange = startRange + r1.Intn(24*1024) // Make sure that end is > start but still within the amount of data generated.
	// Create random data more than 32K.
	buf := make([]byte, rand.Intn(1<<20)+32*1024)
	_, err = io.ReadFull(crand.Reader, buf)
	bufRange = buf[startRange : endRange+1] // Must be inclusive.
	if err != nil {
		return bucketName, objectName, startRange, endRange, bufRange, err
	}
	// Only need host part of endpoint for Minio.
	hostURL, err := url.Parse(config.Endpoint)
	if err != nil {
		return bucketName, objectName, startRange, endRange, bufRange, err
	}
	secure := true // Use HTTPS request
	s3Client, err := minio.New(hostURL.Host, config.Access, config.Secret, secure)
	if err != nil {
		return bucketName, objectName, startRange, endRange, bufRange, err
	}
	// Create a test bucket and object.
	err = s3Client.MakeBucket(bucketName, config.Region)
	if err != nil {
		return bucketName, objectName, startRange, endRange, bufRange, err
	}
	// Upload the entire object.
	_, err = s3Client.PutObject(bucketName, objectName, bytes.NewReader(buf), "binary/octet-stream")
	if err != nil {
		return bucketName, objectName, startRange, endRange, bufRange, err
	}
	return bucketName, objectName, startRange, endRange, bufRange, nil
}

// Test a GET object request with a range header set.
func mainGetObjectRange(config ServerConfig, message string) error {
	// TODO: should errors be returned to the top level or printed here.
	// Spin scanBar
	scanBar(message)
	// Set up a new Bucket and Object to GET
	bucketName, objectName, startRange, endRange, bufRange, err := GetObjectRangeInit(config)
	if err != nil {
		// Attempt a clean up of created object and bucket.
		if errC := GetObjectCleanUp(config, bucketName, objectName); errC != nil {
			return errC
		}
		return err
	}
	// Spin scanBar
	scanBar(message)
	// Create new GET object range request...testing range.
	req, err := NewGetObjectRangeReq(config, bucketName, objectName, startRange, endRange)
	if err != nil {
		// Attempt a clean up of created object and bucket.
		if errC := GetObjectCleanUp(config, bucketName, objectName); errC != nil {
			return errC
		}
		return err
	}
	// Spin scanBar
	scanBar(message)
	// Execute the request.
	res, err := ExecRequest(req)
	if err != nil {
		// Attempt a clean up of created object and bucket.
		if errC := GetObjectCleanUp(config, bucketName, objectName); errC != nil {
			return errC
		}
		return err
	}
	// Spin scanBar
	scanBar(message)
	// Verify the response...these checks do not check the headers yet.
	if err := GetObjectVerify(res, bufRange, "206 Partial Content", nil); err != nil {
		// Attempt a clean up of created object and bucket.
		if errC := GetObjectCleanUp(config, bucketName, objectName); errC != nil {
			return errC
		}
		return err
	}
	// Spin scanBar
	scanBar(message)
	// Clean up after the test.
	if err := GetObjectCleanUp(config, bucketName, objectName); err != nil {
		return err
	}
	// Spin scanBar
	scanBar(message)
	return nil
}
