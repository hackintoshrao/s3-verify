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
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/minio/minio-go"
	"github.com/minio/s3verify/signv4"
)

var RemoveObjectReq = &http.Request{
	Header: map[string][]string{
		// Set Content SHA with empty body for DELETE requests because no data is being uploaded.
		"X-Amz-Content-Sha256": {hex.EncodeToString(signv4.Sum256([]byte{}))},
	},
	Body:   nil, // There is no body for DELETE requests.
	Method: "DELETE",
}

// NewRemoveObjectReq - Create a new DELETE object HTTP request.
func NewRemoveObjectReq(config ServerConfig, bucketName, objectName string) (*http.Request, error) {
	targetURL, err := makeTargetURL(config.Endpoint, bucketName, objectName, config.Region)
	if err != nil {
		return nil, err
	}
	RemoveObjectReq.URL = targetURL

	RemoveObjectReq = signv4.SignV4(*RemoveObjectReq, config.Access, config.Secret, config.Region)
	return RemoveObjectReq, nil
}

// RemoveObjectReqInit - Create a test bucket and object with Minio-Go.
func RemoveObjectInit(s3Client minio.Client, config ServerConfig) (bucketName, objectName string, err error) {
	bucketName = randString(60, rand.NewSource(time.Now().UnixNano()), "s3verify-rm")
	objectName = randString(60, rand.NewSource(time.Now().UnixNano()), "s3verify-rm")

	buf := make([]byte, rand.Intn(1<<20)+32*1024)
	_, err = io.ReadFull(crand.Reader, buf)
	if err != nil {
		return bucketName, objectName, err
	}
	reader := bytes.NewReader(buf)
	// Create a new bucket and object.
	if err = s3Client.MakeBucket(bucketName, config.Region); err != nil {
		return bucketName, objectName, err
	}
	if _, err = s3Client.PutObject(bucketName, objectName, reader, "application/octet-stream"); err != nil {
		return bucketName, objectName, err
	}
	return bucketName, objectName, nil
}

// CleanUpRemoveObject - Clean up after a failed or successful removeobject test.
func CleanUpRemoveObject(s3Client minio.Client, bucketName, objectName string) error {
	if err := cleanUpBucket(s3Client, bucketName, []string{objectName}); err != nil {
		return err
	}
	return nil
}

// RemoveObjectVerify - Verify that the response returned matches what is expected.
func RemoveObjectVerify(res *http.Response, expectedStatus string) error {
	if err := VerifyHeaderRemoveObject(res); err != nil {
		return err
	}
	return nil
}

// VerifyHeaderRemoveObject - Verify that header returned matches what is expected.
func VerifyHeaderRemoveObject(res *http.Response) error {
	if err := verifyStandardHeaders(res); err != nil {
		return err
	}
	return nil
}

// VerifyBodyRemoveObject - Verify that the body returned is empty.
func VerifyBodyRemoveObject(res *http.Response) error {
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

// VerifyStatusRemoveObject - Verify that the status returned matches what is expected.
func VerifyStatusRemoveObject(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus {
		err := fmt.Errorf("Unexpected Status Received: wanted %v, got %v", expectedStatus, res.Status)
		return err
	}
	return nil
}

//
func mainRemoveObjectExists(config ServerConfig, s3Client minio.Client, message string) error {
	// Spin scanBar
	scanBar(message)
	bucketName, objectName, err := RemoveObjectInit(s3Client, config)
	if err != nil {
		// Attempt a clean up.
		if errC := CleanUpRemoveObject(s3Client, bucketName, objectName); errC != nil {
			return errC
		}
		return err
	}
	// Spin scanBar
	scanBar(message)
	// Create a new request.
	req, err := NewRemoveObjectReq(config, bucketName, objectName)
	if err != nil {
		// Attempt a clean up.
		if errC := CleanUpRemoveObject(s3Client, bucketName, objectName); errC != nil {
			return errC
		}

		return err
	}
	// Spin scanBar
	scanBar(message)
	// Execute the request.
	res, err := ExecRequest(req, config.Client)
	if err != nil {
		// Attempt a clean up.
		if errC := CleanUpRemoveObject(s3Client, bucketName, objectName); errC != nil {
			return errC
		}

		return err
	}
	// Spin scanBar
	scanBar(message)
	// Verify the response.
	if err := RemoveObjectVerify(res, "200 OK"); err != nil {
		// Attempt a clean up.
		if errC := CleanUpRemoveObject(s3Client, bucketName, objectName); errC != nil {
			return errC
		}
		return err
	}
	// Spin scanBar
	scanBar(message)
	// Cleanup after the test.
	if errC := CleanUpRemoveObject(s3Client, bucketName, objectName); errC != nil {
		return errC
	}
	// Spin scanBar
	scanBar(message)
	return nil

}
