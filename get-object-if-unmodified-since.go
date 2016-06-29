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
	crand "crypto/rand"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/minio/minio-go"
	"github.com/minio/s3verify/signv4"
)

// An HTTP GET request with the If-Unmodified-Since header set.
var GetObjectIfUnModifiedSinceReq = &http.Request{
	Header: map[string][]string{
		// Set Content SHA with empty body for GET requests because no data is being uploaded.
		"X-Amz-Content-Sha256": {hex.EncodeToString(signv4.Sum256([]byte{}))},
		"If-Unmodified-Since":  {""}, // To be filled dynamically.
	},
	Body:   nil, // There is no body for GET requests.
	Method: "GET",
}

// NewGetObjectIfUnModifiedSinceReq - Create a new HTTP GET request with the If-Unmodified-Since header set to perform.
func NewGetObjectIfUnModifiedSinceReq(config ServerConfig, bucketName, objectName, lastModified string) (*http.Request, error) {
	targetURL, err := makeTargetURL(config.Endpoint, bucketName, objectName, config.Region)
	if err != nil {
		return nil, err
	}
	GetObjectIfUnModifiedSinceReq.Header["If-Unmodified-Since"] = []string{lastModified}

	// Fill request URL and sign.
	GetObjectIfUnModifiedSinceReq.URL = targetURL
	GetObjectIfUnModifiedSinceReq = signv4.SignV4(*GetObjectIfUnModifiedSinceReq, config.Access, config.Secret, config.Region)
	return GetObjectIfUnModifiedSinceReq, nil
}

// GetObjectIfUnModifiedSinceInit - Set up a test bucket and object.
func GetObjectIfUnModifiedSinceInit(config ServerConfig) (bucketName, objectName, lastModified string, buf []byte, err error) {
	// Create a new random bucket and object name prefixed by s3verify-get.
	bucketName = randString(60, rand.NewSource(time.Now().UnixNano()), "s3verify-get")
	objectName = randString(60, rand.NewSource(time.Now().UnixNano()), "s3verify-get")
	lastModified = ""
	// Create random data more than 32K.
	buf = make([]byte, rand.Intn(1<<20)+32*1024)
	_, err = io.ReadFull(crand.Reader, buf)
	if err != nil {
		return bucketName, objectName, lastModified, buf, err
	}
	// Only need host part of endpoint for Minio.
	hostURL, err := url.Parse(config.Endpoint)
	if err != nil {
		return bucketName, objectName, lastModified, buf, err
	}
	secure := true // Use HTTPS request.
	s3Client, err := minio.New(hostURL.Host, config.Access, config.Secret, secure)
	if err != nil {
		return bucketName, objectName, lastModified, buf, err
	}
	// Create the test bucket and object.
	err = s3Client.MakeBucket(bucketName, config.Region)
	if err != nil {
		return bucketName, objectName, lastModified, buf, err
	}
	// Upload the random object.
	_, err = s3Client.PutObject(bucketName, objectName, bytes.NewReader(buf), "binary/octet-stream")
	if err != nil {
		return bucketName, objectName, lastModified, buf, err
	}
	// Gather the Last-Modified field of the object.
	objInfo, err := s3Client.StatObject(bucketName, objectName)
	if err != nil {
		return bucketName, objectName, lastModified, buf, err
	}
	lastModifiedTime := objInfo.LastModified
	lastModified = lastModifiedTime.Format(http.TimeFormat)
	return bucketName, objectName, lastModified, buf, err
}

// VerifyGetObjectIfUnModifiedSince - Verify the response matches what is expected.
func VerifyGetObjectIfUnModifiedSince(res *http.Response, expectedBody []byte, expectedStatus string, expectedHeader map[string][]string, shouldFail bool) error {
	if err := VerifyBodyGetObjectIfUnModifiedSince(res, expectedBody, shouldFail); err != nil {
		return err
	}
	if err := VerifyStatusGetObjectIfUnModifiedSince(res, expectedStatus); err != nil {
		return err
	}
	if err := VerifyHeaderGetObjectIfUnModifiedSince(res, expectedHeader); err != nil {
		return err
	}
	return nil
}

// VerifyGetObjectIfUnModifiedSinceBody - Verify that the response body matches what is expected.
func VerifyBodyGetObjectIfUnModifiedSince(res *http.Response, expectedBody []byte, shouldFail bool) error {
	if shouldFail {
		// Decode the supposed error response.
		errBody := minio.ErrorResponse{}
		decoder := xml.NewDecoder(res.Body)
		err := decoder.Decode(&errBody)
		if err != nil {
			return err
		}
		if errBody.Code != "PreconditionFailed" {
			err := fmt.Errorf("Unexpected Error Response: wanted PreconditionFailed, got %v", errBody.Code)
			return err
		}
	} else {
		// The body should be returned in full.
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		if !bytes.Equal(body, expectedBody) {
			err := fmt.Errorf("Unexpected Body Received: wanted %v, got %v", string(expectedBody), string(body))
			return err
		}
	}
	// Otherwise test failed / passed as expected.
	return nil
}

// VerifyStatusGetObjectIfUnModifiedSince - Verify that the response status matches what is expected.
func VerifyStatusGetObjectIfUnModifiedSince(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus {
		err := fmt.Errorf("Unexpected Response Status: wanted %v, got %v", expectedStatus, res.Status)
		return err
	}
	return nil
}

// VerifyHeaderGetObjectIfUnModifiedSince - Verify that the header returned matches what is expected.
func VerifyHeaderGetObjectIfUnModifiedSince(res *http.Response, expectedHeader map[string][]string) error {
	// TODO: Fill this in.
	return nil
}

// Test the GET object API with the If-Unmodified-Since header set.
func mainGetObjectIfUnModifiedSince(config ServerConfig, message string) error {
	// Set up past date.
	pastDate := "Thu, 01 Jan 1970 00:00:00 GMT"
	// Spin scanBar
	scanBar(message)
	// Set up new bucket and object to GET against.
	bucketName, objectName, lastModified, buf, err := GetObjectIfUnModifiedSinceInit(config)
	if err != nil {
		// Attempt a clean up of created object and bucket.
		if errC := GetObjectCleanUp(config, bucketName, objectName); errC != nil {
			return errC
		}
		return err
	}
	// Spin scanBar
	scanBar(message)
	// Form a request with a pastDate to make sure the object is not returned.
	req, err := NewGetObjectIfUnModifiedSinceReq(config, bucketName, objectName, pastDate)
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
	// Verify that the response returns an error.
	if err := VerifyGetObjectIfUnModifiedSince(res, []byte(""), "412 Precondition Failed", nil, true); err != nil {
		// Attempt a clean up of created object and bucket.
		if errC := GetObjectCleanUp(config, bucketName, objectName); errC != nil {
			return errC
		}
		return err
	}
	// Spin scanBar
	scanBar(message)
	// Form a request with a date in the past.
	curReq, err := NewGetObjectIfUnModifiedSinceReq(config, bucketName, objectName, lastModified)
	if err != nil {
		// Attempt a clean up of created object and bucket.
		if errC := GetObjectCleanUp(config, bucketName, objectName); errC != nil {
			return errC
		}
		return err
	}
	// Spin scanBar
	scanBar(message)
	// Execute current request.
	curRes, err := ExecRequest(curReq)
	if err != nil {
		// Attempt a clean up of created object and bucket.
		if errC := GetObjectCleanUp(config, bucketName, objectName); errC != nil {
			return errC
		}
		return err
	}
	// Spin scanBar
	scanBar(message)
	// Verify that the lastModified date in a request returns the object.
	if err := VerifyGetObjectIfUnModifiedSince(curRes, buf, "200 OK", nil, false); err != nil {
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
