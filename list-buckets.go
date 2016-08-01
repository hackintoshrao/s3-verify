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
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/minio/s3verify/signv4"
)

// newListBucketsReq - Create a new List Buckets request.
func newListBucketsReq(config ServerConfig) (*http.Request, error) {
	// listBucketsReq - a new HTTP request to list all buckets.
	var listBucketsReq = &http.Request{
		Header: map[string][]string{
			// Set Content SHA with empty body for GET requests because no data is being uploaded.
			"X-Amz-Content-Sha256": {hex.EncodeToString(signv4.Sum256([]byte{}))},
		},
		Body:   nil, // There is no body for GET requests.
		Method: "GET",
	}
	// Set the GET req URL.
	// ListBuckets / GET Service is always run through https://s3.amazonaws.com and subsequently us-east-1.
	targetURL, err := makeTargetURL(config.Endpoint, "", "", config.Region, nil)
	if err != nil {
		return nil, err
	}
	listBucketsReq.URL = targetURL
	listBucketsReq.Header.Set("User-Agent", appUserAgent)
	// Sign the necessary headers.
	listBucketsReq = signv4.SignV4(*listBucketsReq, config.Access, config.Secret, config.Region)
	return listBucketsReq, nil
}

// TODO: these checks only verify correctly corrected buckets for now. There is no test made to fail / check failure yet.

// listBucketsVerify - Check for S3 Compatibility in the response Status, Body, and Header
func listBucketsVerify(res *http.Response, expectedStatusCode int, expectedList *listAllMyBucketsResult) error {
	if err := verifyStatusListBuckets(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	if err := verifyBodyListBuckets(res.Body, expectedList); err != nil {
		return err
	}
	if err := verifyHeaderListBuckets(res.Header); err != nil {
		return err
	}
	return nil
}

// verifyStatusListBuckets - Verify that the test was successful.
func verifyStatusListBuckets(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Response Status Code: wanted %v, got %v", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// verifyHeaderListBuckets - Verify that the headers returned match what is expected.
func verifyHeaderListBuckets(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

func isIn(s string, buckets []BucketInfo) (int, bool) {
	for i, bucket := range buckets {
		if s == bucket.Name {
			return i, true
		}
	}
	return -1, false
}

// verifyBodyListBuckets - Verify that the body of the response matches with what is expected.
func verifyBodyListBuckets(resBody io.Reader, expected *listAllMyBucketsResult) error {
	// Extract body from the HTTP response.
	body, err := ioutil.ReadAll(resBody)
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
	i := 0 // Counter for number of buckets found that should be found.
	for _, bucket := range expected.Buckets.Bucket {
		bucketName := bucket.Name
		if pos, there := isIn(bucketName, result.Buckets.Bucket); there {
			i++
			// Check time of creation vs what is listed in body.
			resultTime := result.Buckets.Bucket[pos].CreationDate
			resultTimeStr := resultTime.Format(http.TimeFormat)
			// Make sure that time is returned in http.TimeFormat.
			if _, err := time.Parse(http.TimeFormat, resultTimeStr); err != nil {
				return err
			}
		}
	}
	if i < 2 {
		err := fmt.Errorf("Not all created buckets were listed!")
		return err
	}
	return nil
}

// Test the ListBuckets API with no added parameters.
func mainListBucketsExist(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] ListBuckets:", curTest, globalTotalNumTest)
	// Spin the scanBar
	scanBar(message)
	expectedList := &listAllMyBucketsResult{
		Owner: owner{
			DisplayName: "s3verify",
			ID:          "",
		},
		Buckets: buckets{
			Bucket: validBuckets,
		},
	}

	// Generate new List Buckets request.
	req, err := newListBucketsReq(config)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin the scanBar
	scanBar(message)

	// Generate the server response.
	res, err := execRequest(req, config.Client, "", "")
	if err != nil {
		printMessage(message, err)
		return false
	}
	defer closeResponse(res)
	// Spin the scanBar
	scanBar(message)
	// Check for S3 Compatibility
	if err := listBucketsVerify(res, http.StatusOK, expectedList); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin the scanBar
	scanBar(message)
	printMessage(message, err)
	return true
}
