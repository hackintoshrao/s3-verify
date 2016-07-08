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
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/minio/s3verify/signv4"
)

var (
	// First object to test PUT on.
	putObj1 = &ObjectInfo{
		Key: "s3verify-put-object-test1",
		// LastModified: to be set dynamically,
		// Size: to be set dynamically,
		// ETag: to be set dynamically,
		ContentType: "application/octet-stream",
		Body:        []byte("Nemo enim ipsam voluptatem, quia voluptas sit, aspernature aut odit aut fugit,"),
	}
	// Second object to test PUT on.
	putObj2 = &ObjectInfo{
		Key: "s3verify-put-object-test2",
		// LastModified: to be set dynamically,
		// Size: to be set dynamically,
		// ETag: to be set dynamically,
		ContentType: "application/octet-stream",
		Body:        []byte("sed quia consequuntur magni dolores eos, qui ratione voluptatem sequi nescuint,"),
	}
	// Third object to test PUT on.
	putObj3 = &ObjectInfo{
		Key: "s3verify-put-object-test3",
		// LastModified: to be set dynamically,
		// Size: to be set dynamically,
		// ETag: to be set dynamically,
		ContentType: "application/octet-stream",
		Body:        []byte("neque porro quisquam est, qui dolorem ipsum, quia dolor sit amet, consectetur,"),
	}
)

// Store all regularly PUT objects.
var objects = []*ObjectInfo{
	putObj1,
	putObj2,
	putObj3,
}

// Store all objects that were copied.
var copyObjects = []*ObjectInfo{}

// An HTTP request for a PUT object.
var PutObjectReq = &http.Request{
	Header: map[string][]string{
	// Set Content SHA dynamically because it is based on data being uploaded.
	// Set Content MD5 dynamically because it is based on data being uploaded.
	// Set Content-Length dynamically because it is based on data being uploaded.
	},
	// Body will be set dynamically.
	// Body:
	Method: "PUT",
}

// NewPutObjectReq - Create a new HTTP request for PUT object.
func NewPutObjectReq(config ServerConfig, bucketName, objectName string, objectData []byte) (*http.Request, error) {
	targetURL, err := makeTargetURL(config.Endpoint, bucketName, objectName, config.Region)
	if err != nil {
		return nil, err
	}
	// Fill request headers and URL.
	PutObjectReq.URL = targetURL

	// Compute md5Sum and sha256Sum from the input data.
	reader := bytes.NewReader(objectData)
	md5Sum, sha256Sum, contentLength, err := computeHash(reader)
	if err != nil {
		return nil, err
	}
	PutObjectReq.Header.Set("Content-MD5", base64.StdEncoding.EncodeToString(md5Sum))
	PutObjectReq.Header.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))

	PutObjectReq.ContentLength = contentLength
	// Set the body to the data held in objectData.
	PutObjectReq.Body = ioutil.NopCloser(reader)
	PutObjectReq = signv4.SignV4(*PutObjectReq, config.Access, config.Secret, config.Region)
	return PutObjectReq, nil
}

// PutObjectVerify - Verify the response matches what is expected.
func PutObjectVerify(res *http.Response, expectedStatus string) error {
	if err := VerifyHeaderPutObject(res); err != nil {
		return err
	}
	if err := VerifyStatusPutObject(res, expectedStatus); err != nil {
		return err
	}
	if err := VerifyBodyPutObject(res); err != nil {
		return err
	}
	return nil
}

// VerifyStatusPutObject - Verify that the res status code matches what is expected.
func VerifyStatusPutObject(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus {
		err := fmt.Errorf("Unexpected Response Status Code: wanted %v, got %v", expectedStatus, res.Status)
		return err
	}
	return nil
}

// VerifyBodyPutObject - Verify that the body returned matches what is uploaded.
func VerifyBodyPutObject(res *http.Response) error {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	// A PUT request should give back an empty body.
	if !bytes.Equal(body, []byte{}) {
		err := fmt.Errorf("Unexpected Body Recieved: expected empty body but recieved: %v", string(body))
		return err
	}
	return nil
}

// VerifyHeaderPutObject - Verify that the header returned matches waht is expected.
func VerifyHeaderPutObject(res *http.Response) error {
	if err := verifyStandardHeaders(res); err != nil {
		return err
	}
	return nil
}

// Test a PUT object request with no special headers set. This adds one object to each of the test buckets.
func mainPutObjectNoHeader(config ServerConfig, message string) error {
	// TODO: create tests designed to fail.
	for _, object := range objects {
		bucket := testBuckets[0]
		// Spin scanBar
		scanBar(message)
		// PUT each object in each available bucket.
		// Generate a new PUT object HTTP req.
		req, err := NewPutObjectReq(config, bucket.Name, object.Key, object.Body)
		if err != nil {
			return err
		}
		// Spin scanBar
		scanBar(message)
		// Execute the request.
		res, err := ExecRequest(req, config.Client)
		if err != nil {
			return err
		}
		// Spin scanBar
		scanBar(message)
		// Verify the response.
		if err := PutObjectVerify(res, "200 OK"); err != nil {
			return err
		}
		// Spin scanBar
		scanBar(message)
	}
	return nil

}
