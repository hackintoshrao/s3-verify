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

import "github.com/minio/minio-go/pkg/set"

// BucketPolicy - Bucket level policy.
type BucketPolicy string

const (
	BucketPolicyNone      BucketPolicy = "none"
	BucketPolicyReadOnly               = "readonly"
	BucketPolicyReadWrite              = "readwrite"
	BucketPolicyWriteOnly              = "writeonly"
)

// User - canonical users list.
type User struct {
	AWS set.StringSet
}

// Statement - minio policy statement
type Statement struct {
	Sid        string
	Effect     string
	Principal  User
	Actions    []string                     `json:"Principal"`
	Resources  set.StringSet                `json:"Action"`
	Conditions map[string]map[string]string `json:"Condition,omnitempty"`
}

// BucketAccessPolicy - created bucket policy.
type BucketAccessPolicy struct {
	Version    string      // date in 0000-00-00 format
	Statements []Statement `json:"Statement"`
}
