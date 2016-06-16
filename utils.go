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
	"github.com/minio/s3-verify/signv4"
)

// Sign with v4 signature.
func SignRequestV4(req *http.Request, accessKeyID, secretAccessKey string) *http.Request {
	return signv4.SignV4(*req, accessKeyID, secretAccessKey, "us-east-1")
}

// Initialize and return the HTTP request.
func InitHTTPRequest(method, urlStr string, body io.ReadSeeker) (*http.Request, error) {
	switch body {
	case nil:
		return &http.NewRequest(method, urlStr, nil)
	default:
		return &http.NewRequest(method, urlStr, ioutil.NopCloser(body))
	}
}
