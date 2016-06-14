/*
 * Minio Go Library for Amazon S3 Compatible Cloud Storage (C) 2015, 2016 Minio, Inc.
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
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/minio/s3-verify/signv4"
)

// MakeBucket creates a new bucket with bucketName.

type MakeBucket struct {
	bucketName string
}

// Construct URLPath for MakeBucket.
func (mb MakeBucket) MakeURLPath(endPoint string) (*url.URL, error) {
	makeBucketPath := func(path string) string {
		return "/" + bucketName + "/"
	}
	targetURL, err := url.Parse(endPoint)
	if err != nil {
		return nil, err
	}
	targetURL.Path = makeBucketPath(mb.BucketName)
	return targetURL, nil
}

// Construct Http request with URL Path for MakeBucket.
func (mb MakeBucket) MakePlainRequest() (*http.Request, error) {
	// returns path for creating the bucket.
	// Parse parses rawurl into a URL structure.
	targetURL := mb.MakeURLPath(endPoint)
	req, err := http.NewRequest("PUT", targetURL.String(), nil)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (mb MakeBucket) SignRequest(req *http.Request) (*http.Request, error) {
	return signv4.SignV4(*req, c.accessKeyID, c.secretAccessKey, "us-east-1")
}

// No request body required in this case.
// Method exists to satisy S3Verify interface.
func (mb MakeBucket) SetBody(req *http.Request) *http.Request {
	return req
}

// Set headers.
func (mb MakeBucket) SetHeaders(req *http.Request) (*http.Request, error) {
	req.Header.Set("X-Amz-Content-Sha256", hex.EncodeToString(sum256([]byte{})))
	return req

}

// Executes the HTTP request and returns the response.
// Part of satisfying s3Verify interface.
func (mb MakeBucket) ExecRequest(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, err
}

// Verify the HTTP response.
func (mb MakeBucket) VerifyResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Make Bucket Failed")
		return nil
	}
	return nil
}
