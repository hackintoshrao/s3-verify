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
	"fmt"
	"net/http"
	"net/url"

	"github.com/minio/s3verify/signv4"
)

// Holds all the objects to be uploaded by a multipart request.
var multipartObjects = []*ObjectInfo{
	// An object that will have more than 5MB of data to be uploaded as part of a multipart upload.
	&ObjectInfo{
		Key:         "s3verify-multipart-object",
		ContentType: "application/octet-stream",
		// Body: to be set dynamically,
		// UploadID: to be set dynamically,
	},
	&ObjectInfo{
		Key:         "s3verify-multipart-abort",
		ContentType: "application/octet-stream",
		// Body: to be set dynamically,
		// UploadID: to be set dynamically,
	},
}

// newInitiateMultipartUploadReq - Create a new HTTP request for the initiate-multipart-upload API.
func newInitiateMultipartUploadReq(config ServerConfig, bucketName, objectName string) (*http.Request, error) {
	// Initialize url queries.
	urlValues := make(url.Values)
	urlValues.Set("uploads", "")
	// An HTTP request for a multipart upload.
	var initiateMultipartUploadReq = &http.Request{
		Header: map[string][]string{
		// X-Amz-Content-Sha256 will be set dynamically.
		},
		// Body is empty for the initiate request.
		Method: "POST",
	}
	// Create new targetURL using the ?uploads query.
	targetURL, err := makeTargetURL(config.Endpoint, bucketName, objectName, config.Region, urlValues)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader([]byte(""))
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return nil, err
	}
	// Set the req URL and Header.
	initiateMultipartUploadReq.URL = targetURL
	initiateMultipartUploadReq.Header.Set("User-Agent", appUserAgent)
	initiateMultipartUploadReq.Header.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))

	initiateMultipartUploadReq = signv4.SignV4(*initiateMultipartUploadReq, config.Access, config.Secret, config.Region)
	return initiateMultipartUploadReq, nil
}

// initiateMultipartUploadVerify - verify that the response returned matches what is expected.
func initiateMultipartUploadVerify(res *http.Response, expectedStatus string) (string, error) {
	uploadID, err := verifyBodyInitiateMultipartUpload(res)
	if err != nil {
		return uploadID, err
	}
	if err := verifyHeaderInitiateMultipartUpload(res); err != nil {
		return uploadID, err
	}
	if err := verifyStatusInitiateMultipartUpload(res, expectedStatus); err != nil {
		return uploadID, err
	}
	return uploadID, nil
}

// verifyStatusInitiateMultipartUpload - verify that the status returned matches what is expected.
func verifyStatusInitiateMultipartUpload(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus {
		err := fmt.Errorf("Unexpected Status Received: wanted %v, got %v", expectedStatus, res.Status)
		return err
	}
	return nil
}

// verifyBodyInitiateMultipartUpload - verify that the body returned matches what is expected.
func verifyBodyInitiateMultipartUpload(res *http.Response) (string, error) {
	resInitiateMultipartUpload := initiateMultipartUploadResult{}
	if err := xmlDecoder(res.Body, &resInitiateMultipartUpload); err != nil {
		return "", err
	}
	// Body was sent set the object UploadID.
	uploadID := resInitiateMultipartUpload.UploadID
	return uploadID, nil
}

// verifyHeaderInitiateMultipartUpload - verify that the header returned matches what is expected.
func verifyHeaderInitiateMultipartUpload(res *http.Response) error {
	if err := verifyStandardHeaders(res); err != nil {
		return err
	}
	return nil
}

// mainInitiateMultipartUpload - Entry point for the initiate multipart upload test.
func mainInitiateMultipartUpload(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] Multipart (Initiate-Upload):", curTest, globalTotalNumTest)
	// Spin scanBar.
	scanBar(message)
	// Get the bucket to upload to and the objectName to call the new upload.
	bucket := validBuckets[0]
	multiUploadInitCh := make(chan multiUploadInitChannel, globalRequestPoolSize)
	for i, object := range multipartObjects {
		// Spin scanBar
		scanBar(message)
		go func(objectKey string, cur int) {
			// Create a new InitiateMultiPartUpload request.
			req, err := newInitiateMultipartUploadReq(config, bucket.Name, objectKey)
			if err != nil {
				multiUploadInitCh <- multiUploadInitChannel{
					index:    cur,
					err:      err,
					uploadID: "",
				}
				return
			}
			// Execute the request.
			res, err := execRequest(req, config.Client, bucket.Name, objectKey)
			if err != nil {
				multiUploadInitCh <- multiUploadInitChannel{
					index:    cur,
					err:      err,
					uploadID: "",
				}
				return
			}
			defer closeResponse(res)
			// Verify the response and get the uploadID.
			uploadID, err := initiateMultipartUploadVerify(res, "200 OK")
			if err != nil {
				multiUploadInitCh <- multiUploadInitChannel{
					index:    cur,
					err:      err,
					uploadID: "",
				}
				return
			}
			// Save the current initiate and uploadID.
			multiUploadInitCh <- multiUploadInitChannel{
				index:    cur,
				err:      nil,
				uploadID: uploadID,
			}
		}(object.Key, i)
		// Spin scanBar
		scanBar(message)
	}
	count := len(multipartObjects)
	for count > 0 {
		count--
		// Spin scanBar
		scanBar(message)
		uploadInfo, ok := <-multiUploadInitCh
		if !ok {
			return false
		}
		// If the initiate failed exit.
		if uploadInfo.err != nil {
			printMessage(message, uploadInfo.err)
			return false
		}
		// Retrieve the specific uploadID that was started.
		object := multipartObjects[uploadInfo.index]
		// Set the uploadId of the uploaded object.
		object.UploadID = uploadInfo.uploadID
		// Spin scanBar
		scanBar(message)
	}
	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, nil)
	return true
}
