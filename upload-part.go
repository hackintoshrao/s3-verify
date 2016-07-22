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
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/minio/s3verify/signv4"
)

// One part of a multipart upload.
var object1Part = &objectPart{}

// Complete multipart upload.
var complMultipartUploads = []*completeMultipartUpload{
	&completeMultipartUpload{
	// To be filled out by the test.
	},
	&completeMultipartUpload{
	// To be filled out by the test.
	},
}

// newUploadPartReq - Create a new HTTP request for an upload part request.
func newUploadPartReq(config ServerConfig, bucketName, objectName, uploadID string, partNumber int, partData []byte) (*http.Request, error) {
	// Create a new request for uploading a part.
	var uploadPartReq = &http.Request{
		Header: map[string][]string{
		// X-Amz-Content-Sha256 will be set dynamically.
		// Content-Length will be set dynamically.
		// Content-MD5 will be set dynamically.
		},
		// Body: will be set dynamically,
		Method: "PUT",
	}
	urlValues := make(url.Values)
	// Set part number.
	urlValues.Set("partNumber", strconv.Itoa(partNumber))
	// Set upload id.
	urlValues.Set("uploadId", uploadID)

	targetURL, err := makeTargetURL(config.Endpoint, bucketName, objectName, config.Region, urlValues)
	if err != nil {
		return nil, err
	}
	// Compute md5sum, sha256Sum and contentlength.
	reader := bytes.NewReader(partData)
	md5Sum, sha256Sum, contentLength, err := computeHash(reader)
	// Set the Header values, URL, and Body of request.
	uploadPartReq.URL = targetURL
	uploadPartReq.Body = ioutil.NopCloser(reader)
	uploadPartReq.ContentLength = contentLength
	uploadPartReq.Header.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	uploadPartReq.Header.Set("Content-MD5", base64.StdEncoding.EncodeToString(md5Sum))

	uploadPartReq = signv4.SignV4(*uploadPartReq, config.Access, config.Secret, config.Region)
	return uploadPartReq, nil
}

// uploadPartVerify - verify that the response returned matches what is expected.
func uploadPartVerify(res *http.Response, expectedStatus string) error {
	if err := verifyBodyUploadPart(res); err != nil {
		return err
	}
	if err := verifyStatusUploadPart(res, expectedStatus); err != nil {
		return err
	}
	if err := verifyHeaderUploadPart(res); err != nil {
		return err
	}
	return nil
}

// verifyBodyUploadPart - verify that that body returned is empty.
func verifyBodyUploadPart(res *http.Response) error {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if !bytes.Equal(body, []byte("")) { // Body for PUT responses should be empty.
		err := fmt.Errorf("Unexpected Body Received: %v", string(body))
		return err
	}
	return nil
}

// verifyStatusUploadPart - verify that the status returned matches what is expected.
func verifyStatusUploadPart(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus {
		err := fmt.Errorf("Unexpected Status Received: wanted %v, got %v", expectedStatus, res.Status)
		return err
	}
	return nil
}

// verifyHeaderUploadPart - verify that the header returned matches what is expected.
func verifyHeaderUploadPart(res *http.Response) error {
	if err := verifyStandardHeaders(res); err != nil {
		return err
	}
	return nil
}

// mainUploadPart - Entry point for the upload part test.
func mainUploadPart(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] Multipart (Upload-Part):", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)
	bucket := validBuckets[0]
	// Create some random data at most 5MB to upload via multipart operations.
	objectData := make([]byte, rand.Intn(1<<20)+4*1024*1024)
	_, err := io.ReadFull(crand.Reader, objectData)
	if err != nil {
		printMessage(message, err)
		return false
	}
	object1Part.PartNumber = 1                // First part uploaded
	object1Part.Size = int64(len(objectData)) // Upload all of the data in one part.
	// Spin scanBar
	scanBar(message)
	for i, object := range multipartObjects {
		// Create a new multipart upload part request.
		req, err := newUploadPartReq(config, bucket.Name, object.Key, object.UploadID, 1, objectData)
		if err != nil {
			printMessage(message, err)
			return false
		}
		// Spin scanBar
		scanBar(message)
		// Execute the request.
		res, err := execRequest(req, config.Client)
		if err != nil {
			printMessage(message, err)
			return false
		}
		// Spin scanBar
		scanBar(message)
		// Verify the response.
		if err := uploadPartVerify(res, "200 OK"); err != nil {
			printMessage(message, err)
			return false
		}
		// Spin scanBar
		scanBar(message)
		// Update the ETag of the part.
		object1Part.ETag = strings.TrimPrefix(res.Header.Get("ETag"), "\"")
		object1Part.ETag = strings.TrimSuffix(object1Part.ETag, "\"")
		var complPart completePart
		complPart.ETag = object1Part.ETag
		complPart.PartNumber = object1Part.PartNumber
		// Save the completed part into the complMultiPartUpload struct.
		complMultipartUploads[i].Parts = append(complMultipartUploads[i].Parts, complPart)
	}
	printMessage(message, nil)
	return true
}
