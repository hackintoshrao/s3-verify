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
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/minio/s3verify/signv4"
)

//
func newCompleteMultipartUploadReq(config ServerConfig, bucketName, objectName, uploadID string, complete *completeMultipartUpload) (*http.Request, error) {
	var completeMultipartUploadReq = &http.Request{
		Header: map[string][]string{
		// X-Amz-Content-Sha256 will be set dynamically,
		// Content-Length will be set dynamically,
		},
		// Body: will be set dynamically,
		Method: "POST",
	}
	// Initialize url queries.
	urlValues := make(url.Values)
	urlValues.Set("uploadId", uploadID)

	targetURL, err := makeTargetURL(config.Endpoint, bucketName, objectName, config.Region, urlValues)
	if err != nil {
		return nil, err
	}
	completeMultipartUploadBytes, err := xml.Marshal(complete)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(completeMultipartUploadBytes)
	// Compute sha256Sum and contentLength.
	_, sha256Sum, contentLength, err := computeHash(reader)
	if err != nil {
		return nil, err
	}

	// Set the Body, Header and URL of the request.
	completeMultipartUploadReq.URL = targetURL
	completeMultipartUploadReq.ContentLength = contentLength
	completeMultipartUploadReq.Header.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	completeMultipartUploadReq.Body = ioutil.NopCloser(reader)

	completeMultipartUploadReq = signv4.SignV4(*completeMultipartUploadReq, config.Access, config.Secret, config.Region)
	return completeMultipartUploadReq, nil
}

// TODO: So far only valid multipart requests are used. Implement tests that SHOULD fail.
//
func completeMultipartUploadVerify(res *http.Response, expectedStatus string) error {
	if err := verifyStatusCompleteMultipartUpload(res, expectedStatus); err != nil {
		return err
	}
	if err := verifyBodyCompleteMultipartUpload(res); err != nil {
		return err
	}
	if err := verifyHeaderCompleteMultipartUpload(res); err != nil {
		return err
	}
	return nil
}

//
func verifyStatusCompleteMultipartUpload(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus {
		err := fmt.Errorf("Unexpected Status Received: wanted %v, got %v", expectedStatus, res.Status)
		return err
	}
	return nil
}

func verifyBodyCompleteMultipartUpload(res *http.Response) error {
	resCompleteMultipartUploadResult := completeMultipartUploadResult{}
	if err := xmlDecoder(res.Body, &resCompleteMultipartUploadResult); err != nil {
		return err
	}
	return nil
}

func verifyHeaderCompleteMultipartUpload(res *http.Response) error {
	if err := verifyStandardHeaders(res); err != nil {
		return err
	}
	return nil
}

// mainCompleteMultipartUpload - Entry point for the Complete Multipart Upload API test.
func mainCompleteMultipartUpload(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] Multipart (Complete-Upload):", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)
	bucket := validBuckets[0]
	object := multipartObjects[0]
	// Create a new completeMultipartUpload request.
	req, err := newCompleteMultipartUploadReq(config, bucket.Name, object.Key, object.UploadID, complMultipartUploads[0])
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Execute the request.
	res, err := execRequest(req, config.Client, bucket.Name, object.Key)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Verify the response.
	if err := completeMultipartUploadVerify(res, "200 OK"); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	printMessage(message, nil)
	return true
}
