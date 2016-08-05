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

import "github.com/minio/cli"

// Collection of flags currently supported by every command.
var globalFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "access, a",
		Usage: "Set AWS S3 access key.",
		// Allow env. variables to be used as well as flags.
		EnvVar: "S3_ACCESS",
	},
	cli.StringFlag{
		Name:  "secret, s",
		Usage: "Set AWS S3 secret key.",
		// Allow env. variables to be used as well as flags.
		EnvVar: "S3_SECRET",
	},
	cli.StringFlag{
		Name:  "region, r",
		Value: globalDefaultRegion,
		Usage: "Set AWS S3 region. Defaults to 'us-east-1'. Do not use 'us-east-1' with AWS endpoint or automatic clean up will fail.",
		// Allow env. variables to used as well as flags.
		EnvVar: "S3_REGION",
	},
	cli.StringFlag{
		Name:   "url, u",
		Usage:  "URL to S3 compatible server.",
		EnvVar: "S3_URL",
	},
	cli.BoolFlag{
		Name:  "verbose, v",
		Usage: "Enable verbose output.",
	},
	cli.BoolFlag{
		Name:  "extended",
		Usage: "Enable testing of extra S3 APIs.",
	},
	cli.BoolFlag{
		Name:  "prepare",
		Usage: "Prepare the testing environment with 2 buckets and 1001 objects and 1001 multipart object parts.",
	},
	cli.StringFlag{
		Name:  "clean",
		Usage: "Remove all s3verify objects and multipart object parts from the specified bucket. If only s3verify material is in the bucket the bucket itself will be removed as well.",
	},
}
