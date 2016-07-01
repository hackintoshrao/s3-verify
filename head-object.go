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

// HeadObjectReq - an HTTP request for HEAD with no headers set.
var HeadObjectReq = &http.Request{
	Header: map[string][]string{
		// Set Content SHA with an empty for HEAD requests because no data is being uploaded.
		"X-Amz-Content-Sha256": {hex.EncodeToString(signv4.Sum256([]byte{}))},
	},
	Body:   nil, // No body is sent with HEAD requests.
	Method: "HEAD",
}

// NewHeadObjectReq - Create a new HTTP request for a HEAD object.
func NewHeadObjectReq(config ServerConfig, bucketName, objectName string) (*http.Request, error) {
	targetURL, err := makeTargetURL(config.Endpoint, bucketName, objectName, config.Region)
	if err != nil {
		return nil, err
	}
	// Fill request URL and sign.
	HeadObjectReq.URL = targetURL
	HeadObjectReq = signv4.SignV4(*HeadObjectReq, config.Access, config.Secret, config.Region)
	return HeadObjectReq, nil
}

// HeadObjectInit - Create a test bucket and object to perform the HEAD request on.
func HeadObjectInit(s3Client minio.Client, config ServerConfig) (bucketName, objectName string, err error) {
	// Generate random bucket and object names prefixed by s3verify-head.
	bucketName = randString(60, rand.NewSource(time.Now().UnixNano()), "s3verify-head")
	objectName = randString(60, rand.NewSource(time.Now().UnixNano()), "s3verify-head")

	// Create random data more than 32K.
	buf := make([]byte, rand.Intn(1<<20)+32*1024)
	_, err = io.ReadFull(crand.Reader, buf)
	if err != nil {
		return bucketName, objectName, err
	}
	// Create the test bucket and object.
	err = s3Client.MakeBucket(bucketName, config.Region)
	if err != nil {
		return bucketName, objectName, err
	}
	_, err = s3Client.PutObject(bucketName, objectName, bytes.NewReader(buf), "binary/octet-stream")
	if err != nil {
		return bucketName, objectName, err
	}
	return bucketName, objectName, nil
}

// HeadObjectVerify - Verify that the response received matches what is expected.
func HeadObjectVerify(res *http.Response, expectedStatus string, expectedHeader map[string][]string) error {
	if err := VerifyStatusHeadObject(res, expectedStatus); err != nil {
		return err
	}
	if err := VerifyHeaderHeadObject(res, expectedHeader); err != nil {
		return err
	}
	if err := VerifyBodyHeadObject(res); err != nil {
		return err
	}
	return nil
}

// VerifyStatusHeadObject - Verify that the status received matches what is expected.
func VerifyStatusHeadObject(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus {
		err := fmt.Errorf("Unexpected Response Status Code: wanted %v, got %v", expectedStatus, res.Status)
		return err
	}
	return nil
}

// VerifyBodyHeadObject - Verify that the body recieved is empty.
func VerifyBodyHeadObject(res *http.Response) error {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if !bytes.Equal(body, []byte{}) {
		err := fmt.Errorf("Unexpected Body Recieved: HEAD requests should not return a body, but got back: %v", string(body))
		return err
	}
	return nil
}

// VerifyHeaderHeadObject - Verify that the header received matches what is exepected.
func VerifyHeaderHeadObject(res *http.Response, expectedHeader map[string][]string) error {
	// TODO: fill this in.
	return nil
}

// Test the HeadObject API with no header set.
func mainHeadObjectNoHeader(config ServerConfig, s3Client minio.Client, message string) error {
	// Spin scanBar
	scanBar(message)
	// Set up new bucket and object to test on.
	bucketName, objectName, err := HeadObjectInit(s3Client, config)
	if err != nil {
		// Attempt a clean up of created object and bucket.
		if errC := cleanUpTest(s3Client, []string{bucketName}, []string{objectName}); errC != nil {
			return errC
		}
		return err
	}
	// Spin scanBar
	scanBar(message)
	// Create a new HEAD object with no headers.
	req, err := NewHeadObjectReq(config, bucketName, objectName)
	if err != nil {
		// Attempt a clean up of created object and bucket.
		if errC := cleanUpTest(s3Client, []string{bucketName}, []string{objectName}); errC != nil {
			return errC
		}
		return err
	}
	// Spin scanBar
	scanBar(message)
	res, err := ExecRequest(req, config.Client)
	if err != nil {
		// Attempt a clean up of created object and bucket.
		if errC := cleanUpTest(s3Client, []string{bucketName}, []string{objectName}); errC != nil {
			return errC
		}
		return err
	}
	// Spin scanBar
	scanBar(message)

	// Verify the response.
	if err := HeadObjectVerify(res, "200 OK", nil); err != nil {
		// Attempt a clean up of the object and bucket.
		if errC := cleanUpTest(s3Client, []string{bucketName}, []string{objectName}); errC != nil {
			return errC
		}
		return err
	}
	// Spin scanBar
	scanBar(message)
	// Clean up after the test.
	if err := cleanUpTest(s3Client, []string{bucketName}, []string{objectName}); err != nil {
		return err
	}
	// Spin scanBar
	scanBar(message)
	return nil
}
