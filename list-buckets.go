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
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/minio/minio-go"
	"github.com/minio/s3verify/signv4"
)

// Container for the result of a ListBuckets request.
type listAllMyBucketsResult struct {
	Buckets buckets
	Owner   owner
}
type buckets struct {
	Bucket []bucketInfo
}

// Owner of the buckets listed by a ListBuckets request.
type owner struct {
	ID          string
	DisplayName string
}

// Container for data associated with each bucket listed by a ListBuckets request.
type bucketInfo struct {
	Name         string    `json:"name"`
	CreationDate time.Time `json:"creationDate"`
}

// ListBucketsReq - hardcode the static portions of a new List buckets request.
var ListBucketsReq = &http.Request{
	Header: map[string][]string{
		// Set Content SHA with empty body for GET / DELETE requests because no data is being uploaded.
		"X-Amz-Content-Sha256": {hex.EncodeToString(signv4.Sum256([]byte{}))},
	},
	Body:   nil, // There is no body for GET / DELETE requests.
	Method: "GET",
}

// NewListBucketsReq - Create a new List Buckets request.
func NewListBucketsReq(config ServerConfig) (*http.Request, error) {
	// Set the GET req URL.
	targetURL, err := url.Parse(config.Endpoint)
	if err != nil {
		return nil, err
	}
	targetURL.Path = "/"
	ListBucketsReq.URL = targetURL
	// Sign the necessary headers.
	ListBucketsReq = signv4.SignV4(*ListBucketsReq, config.Access, config.Secret, config.Region)
	return ListBucketsReq, nil
}

// TODO: ListBuckets can also run with no buckets...how to test?

// ListBucketsInit - Create a list of buckets to List using
func ListBucketsInit(config ServerConfig) (*listAllMyBucketsResult, error) {
	created := []bucketInfo{}
	// Minio only needs Host part of URL.
	secure := true
	hostURL, err := url.Parse(config.Endpoint)
	if err != nil {
		return nil, err
	}
	s3Client, err := minio.New(hostURL.Host, config.Access, config.Secret, secure)
	if err != nil {
		return nil, err
	}
	// Generate x amount of buckets to test listing on.
	// Should increase this beyond 3.
	for i := 0; i < 3; i++ {
		bucketName := randString(60, rand.NewSource(time.Now().UnixNano()), "s3verify-lb")
		start := time.Now().UTC()
		err := s3Client.MakeBucket(bucketName, config.Region)
		duration := time.Since(start)
		if err != nil {
			return nil, err
		}
		createdAt := start.Add(duration) // Best guess for creation time. Will accept a window around this.
		if err != nil {
			return nil, err
		}
		// Create a bucketInfo struct for each new bucket.
		bucket := bucketInfo{
			Name:         bucketName,
			CreationDate: createdAt,
		}
		created = append(created, bucket)
	}
	// Create owner struct for the expected results.
	owner := owner{
		ID:          "",
		DisplayName: "S3Verify",
	}
	// Fill the buckets part of the expected results with the Minio made buckets.
	buckets := buckets{
		Bucket: created,
	}
	// Create an overall expected struct to compare with the response.
	expected := &listAllMyBucketsResult{
		Buckets: buckets,
		Owner:   owner,
	}
	return expected, nil
}

// ListBucketsCleanUp - Purge all buckets that were added in order to test listing.
func ListBucketsCleanUp(config ServerConfig, created *listAllMyBucketsResult) error {
	secure := true
	hostURL, err := url.Parse(config.Endpoint)
	if err != nil {
		return err
	}
	s3Client, err := minio.New(hostURL.Host, config.Access, config.Secret, secure)
	if err != nil {
		return err
	}
	buckets := created.Buckets.Bucket
	// Delete all test created buckets.
	for _, bucket := range buckets {
		err := s3Client.RemoveBucket(bucket.Name)
		if err != nil {
			// Bucket may not have been created successfully.
			if minio.ToErrorResponse(err).Code == "NoSuchBucket" { // Only use codes for now, strings unreliable.
				return nil
			}
			return err
		}
	}
	return nil
}

// TODO: these checks only verify correctly corrected buckets for now. There is no test made to fail / check failure yet.

// ListBucketsVerify - Check for S3 Compatibility in the response Status, Body, and Header
func ListBucketsVerify(res *http.Response, expected *listAllMyBucketsResult) error {
	if err := VerifyStatusListBuckets(res); err != nil {
		return err
	}
	if err := VerifyBodyListBuckets(res, expected); err != nil {
		return err
	}
	if err := VerifyHeaderListBuckets(res); err != nil {
		return err
	}
	return nil
}

// VerifyStatusListBuckets - Verify that the test was successful.
func VerifyStatusListBuckets(res *http.Response) error {
	if res.StatusCode != http.StatusOK {
		err := fmt.Errorf("Unexpected Response Status Code: %v", res.StatusCode)
		return err
	}
	return nil
}

// VerifyHeaderListBuckets - Verify that the headers returned match what is expected.
func VerifyHeaderListBuckets(res *http.Response) error {
	// TODO: Needs discussion on what to actually check here...
	return nil
}

func isIn(s string, buckets []bucketInfo) (int, bool) {
	for i, bucket := range buckets {
		if s == bucket.Name {
			return i, true
		}
	}
	return -1, false
}

// VerifyBodyListBuckets - Verify that the body of the response matches with what is expected.
func VerifyBodyListBuckets(res *http.Response, expected *listAllMyBucketsResult) error {
	// Extract body from the HTTP response.
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	// Create list of buckets out of the response.
	result := listAllMyBucketsResult{}
	err = xml.Unmarshal([]byte(body), &result)
	if err != nil {
		return err
	}
	// Check that lists contain all created buckets.
	if len(result.Buckets.Bucket) < len(expected.Buckets.Bucket) {
		err := fmt.Errorf("Fewer buckets reported than were created!")
		return err
	}
	i := 0                                         // Counter for number of buckets found that should be found.
	acceptable := time.Duration(300) * time.Second // Acceptable range between request and creation of test buckets
	for _, bucket := range expected.Buckets.Bucket {
		bucketTime := bucket.CreationDate
		bucketName := bucket.Name
		if pos, there := isIn(bucketName, result.Buckets.Bucket); there {
			i++
			// Check time of creation vs what is listed in body.
			resultTime := result.Buckets.Bucket[pos].CreationDate
			// TODO: Check this logic...
			if resultTime.Sub(bucketTime) >= acceptable {
				err := fmt.Errorf(`The returned CreationDate for bucket %v is more than %v seconds wrong.
				Wanted %v +/- %v, got %v.`, bucketName, acceptable, bucketTime, acceptable, resultTime)
				return err
			}
		}
	}
	if i < 3 {
		err := fmt.Errorf("Not all created buckets were listed!")
		return err
	}
	return nil
}

// Test the ListBuckets API with no added parameters.
func mainListBucketsExist(config ServerConfig, message string) error {
	// Spin the scanBar
	scanBar(message)
	// Create a pseudo body for a http.Response
	expectedBody, err := ListBucketsInit(config)
	if err != nil {
		// Attempt a clean up of the created buckets.
		if errC := ListBucketsCleanUp(config, expectedBody); errC != nil {
			return errC
		}
		return err
	}
	// Spin the scanBar
	scanBar(message)
	// Generate new List Buckets request.
	req, err := NewListBucketsReq(config)
	if err != nil {
		// Attempt a clean up of the created buckets.
		if errC := ListBucketsCleanUp(config, expectedBody); errC != nil {
			return errC
		}
		return err
	}
	// Spin the scanBar
	scanBar(message)

	// Generate the server response.
	res, err := ExecRequest(req)
	if err != nil {
		// Attempt a clean up of the created buckets.
		if errC := ListBucketsCleanUp(config, expectedBody); errC != nil {
			return errC
		}
		return err
	}
	// Spin the scanBar
	scanBar(message)

	// Check for S3 Compatibility
	if err := ListBucketsVerify(res, expectedBody); err != nil {
		// Attempt a clean up of the created buckets.
		if errC := ListBucketsCleanUp(config, expectedBody); errC != nil {
			return errC
		}
		return err
	}
	// Spin the scanBar
	scanBar(message)

	// Delete all Minio created test buckets.
	if err := ListBucketsCleanUp(config, expectedBody); err != nil {
		return err
	}
	return nil
}
