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
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/minio/minio-go"
	"github.com/minio/s3verify/signv4"
)

var PutObjectCopyReq = &http.Request{
	Header: map[string][]string{
	// X-Amz-Content-Sha256 will be set dynamically.
	// Content-MD5 will be set dynamically.
	// Content-Length will be set dynamically.
	// x-amz-copy-source will be set dynamically.
	},
	// Body will be set dynamically.
	// Body:
	Method: "PUT",
}

type copyObjectResult struct {
	ETag         string
	LastModified string // time string format "2006-01-02T15:04:05.000Z"
}

// NewPutObjectCopyReq - Create a new HTTP request for PUT object with copy-
func NewPutObjectCopyReq(config ServerConfig, sourceBucketName, sourceObjectName, destBucketName, destObjectName string, objectData []byte) (*http.Request, error) {
	targetURL, err := makeTargetURL(config.Endpoint, destBucketName, destObjectName, config.Region)
	if err != nil {
		return nil, err
	}
	// Fill request URL.
	PutObjectCopyReq.URL = targetURL

	// Compute md5Sum and sha256Sum from the input data.
	reader := bytes.NewReader(objectData)
	md5Sum, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return nil, err
	}
	// Fill request headers.
	PutObjectCopyReq.Header.Set("Content-MD5", base64.StdEncoding.EncodeToString(md5Sum))
	PutObjectCopyReq.Header.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	PutObjectCopyReq.Header.Set("x-amz-copy-source", url.QueryEscape(sourceBucketName+"/"+sourceObjectName))

	PutObjectCopyReq = signv4.SignV4(*PutObjectCopyReq, config.Access, config.Secret, config.Region)

	return PutObjectCopyReq, nil
}

// PutObjectCopyReqInit - Create two test buckets and one object to copy.
func PutObjectCopyInit(s3Client minio.Client, config ServerConfig) (sourceBucketName, sourceObjectName, destBucketName, destObjectName string, buf []byte, err error) {
	// Create a random source/dest bucketName and objectName prefixed by 's3verify-copy'.
	sourceBucketName = randString(60, rand.NewSource(time.Now().UnixNano()), "s3verify-copy")
	sourceObjectName = randString(60, rand.NewSource(time.Now().UnixNano()), "s3verify-copy")
	destBucketName = randString(60, rand.NewSource(time.Now().UnixNano()), "s3verify-copy")
	destObjectName = randString(60, rand.NewSource(time.Now().UnixNano()), "s3verify-copy")

	// Create random data more than 32K.
	buf = make([]byte, rand.Intn(1<<20)+32*1024)
	_, err = io.ReadFull(crand.Reader, buf)
	if err != nil {
		return sourceBucketName, sourceObjectName, destBucketName, destObjectName, buf, err
	}
	// Create the test bucket and object.
	err = s3Client.MakeBucket(sourceBucketName, config.Region)
	if err != nil {
		return sourceBucketName, sourceObjectName, destBucketName, destObjectName, buf, err
	}
	_, err = s3Client.PutObject(sourceBucketName, sourceObjectName, bytes.NewReader(buf), "binary/octet-stream")
	if err != nil {
		return sourceBucketName, sourceObjectName, destBucketName, destObjectName, buf, err
	}
	// Create test copy bucket.
	err = s3Client.MakeBucket(destBucketName, config.Region)
	if err != nil {
		return sourceBucketName, sourceObjectName, destBucketName, destObjectName, buf, err
	}
	return sourceBucketName, sourceObjectName, destBucketName, destObjectName, buf, err
}

//
func PutObjectCopyVerify(res *http.Response, expectedStatus string) error {
	if err := VerifyHeaderPutObjectCopy(res); err != nil {
		return err
	}
	if err := VerifyBodyPutObjectCopy(res); err != nil {
		return err
	}
	if err := VerifyStatusPutObjectCopy(res, expectedStatus); err != nil {
		return err
	}
	return nil
}

// VerifyHeadersPutObjectCopy - Verify that the header returned matches what is expected.
func VerifyHeaderPutObjectCopy(res *http.Response) error {
	if err := verifyStandardHeaders(res); err != nil {
		return err
	}
	return nil
}

// VerifyBodyPutObjectCopy - Verify that the body returned is empty.
func VerifyBodyPutObjectCopy(res *http.Response) error {
	copyObjRes := copyObjectResult{}
	decoder := xml.NewDecoder(res.Body)
	err := decoder.Decode(&copyObjRes)
	if err != nil {
		return err
	}
	return nil
}

// VerifyStatusPutObjectCopy - Verify that the status returned matches what is expected.
func VerifyStatusPutObjectCopy(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus {
		err := fmt.Errorf("Unexpected Response Status Code: wanted %v, got %v", expectedStatus, res.Status)
		return err
	}
	return nil
}

func CleanUpCopyObject(s3Client minio.Client, bucketNames, objectNames []string) error {
	for _, bucketName := range bucketNames {
		if err := cleanUpBucket(s3Client, bucketName, objectNames); err != nil {
			return err
		}
	}
	return nil
}

// Test a PUT object request with the copy header set.
func mainPutObjectCopy(config ServerConfig, s3Client minio.Client, message string) error {
	// TODO: differentiate errors s3verify, minio-go, failed tests.
	// Spin scanBar
	scanBar(message)
	// Set up new test buckets and object.
	sourceBucketName, sourceObjectName, destBucketName, destObjectName, buf, err := PutObjectCopyInit(s3Client, config)
	bucketNames := []string{sourceBucketName, destBucketName}
	objectNames := []string{sourceObjectName, destObjectName}
	if err != nil {
		// Attempt a clean up of the created object and buckets.
		if errC := CleanUpCopyObject(s3Client, bucketNames, objectNames); errC != nil {
			return errC
		}
		return err
	}
	// Spin scanBar
	scanBar(message)
	// Create a new request.
	req, err := NewPutObjectCopyReq(config, sourceBucketName, sourceObjectName, destBucketName, destObjectName, buf)
	if err != nil {
		// Attempt a clean up of the created object and buckets.
		if errC := CleanUpCopyObject(s3Client, bucketNames, objectNames); errC != nil {
			return errC
		}
		return err
	}
	// Spin scanBar
	scanBar(message)
	// Execute the request.
	res, err := ExecRequest(req, config.Client)
	if err != nil {
		// Attempt a clean up of the created object and buckets.
		if errC := CleanUpCopyObject(s3Client, bucketNames, objectNames); errC != nil {
			return errC
		}
		return err
	}
	// Spin scanBar
	scanBar(message)
	// Verify the response.
	if err = PutObjectCopyVerify(res, "200 OK"); err != nil {
		// Attempt a clean up of the created object and buckets.
		if errC := CleanUpCopyObject(s3Client, bucketNames, objectNames); errC != nil {
			return errC
		}
		return err
	}
	// Spin scanBar
	scanBar(message)
	// Clean up the test.
	if err = CleanUpCopyObject(s3Client, bucketNames, objectNames); err != nil {
		return err
	}
	// Spin scanBar
	scanBar(message)
	return nil
}
