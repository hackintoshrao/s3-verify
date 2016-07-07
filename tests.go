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

var (
	// Messages printed during the running of the GetObject tests.
	// When a new test for GetObject is added make sure it is added here.
	getObjectMessages = []string{
		"GetObject (No Header):",
		"GetObject (Range):",
		"GetObject (If-Match):",
		"GetObject (If-None-Match):",
		"GetObject (If-Modified-Since):",
		"GetObject (If-Unmodified-Since):",
	}

	// Declare all tests run for the GetObject API.
	// When a new test for GetObject is added make sure its added here.
	getObjectTests = []APItest{
		mainGetObjectNoHeader,
		mainGetObjectRange,
		mainGetObjectIfMatch,
		mainGetObjectIfNoneMatch,
		mainGetObjectIfModifiedSince,
		mainGetObjectIfUnModifiedSince,
	}

	// Messages printed during the running of the HeadObject tests.
	// When a new test for HeadObject is added make sure its message is added here.
	headObjectMessages = []string{
		"HeadObject:",
	}

	// Declare all tests run for the HeadObject API.
	// When a new test for HeadObject is added make sure its added here.
	headObjectTests = []APItest{
		mainHeadObjectNoHeader,
	}
	// Messages printed during the running of the ListBuckets tests.
	// When a new test for ListBuckets is added make sure its message is added here.
	listBucketsMessages = []string{
		"ListBuckets (No Params):",
	}

	// Declare all tests run for the ListBuckets API.
	// When a new test for ListBuckets is added make sure its added here.
	listBucketsTests = []APItest{
		mainListBucketsExist,
	}

	// Messages printed during the running of the MakeBucket tests.
	// When a new test for MakeBucket is added make sure to add its message here.
	makeBucketMessages = []string{
		"MakeBucket (No Header):",
	}

	// Declare all tests run for the MakeBucket API.
	// When a new test for MakeBucket is added make sure its added here.
	makeBucketTests = []APItest{
		mainMakeBucketNoHeader,
	}

	// Messages to be printed during the PutObject tests.
	// When a new test for PutObject is added make sure its message is added here.
	putObjectMessages = []string{
		"PutObject (No Header):",
	}
	// Declare all tests run for the PutObject API.
	// When a new test for PutObject is added make sure its added here.
	putObjectTests = []APItest{
		mainPutObjectNoHeader,
	}

	// Messages to be printed during the RemoveBucket tests.
	// When a new test for RemoveBucket is added make sure its message is added here.
	removeBucketMessages = []string{
		"RemoveBucket (Bucket Exists):",
	}

	// Declare all tests run for the RemoveBucket API.
	// When a new test for RemoveBucket is added make sure its added here.
	removeBucketTests = []APItest{
		mainRemoveBucketExists,
	}

	// Messages to be printed during the RemoveObject tests.
	// When a new test for RemoveObject is added make sure its message is added here.
	removeObjectMessages = []string{
		"RemoveObject (Object Exists):",
	}

	// Declare all tests run for the RemoveObject API.
	// When a new test for RemoveObject is added make sure its added here.
	removeObjectTests = []APItest{
		mainRemoveObjectExists,
	}
)
