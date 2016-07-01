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
	"fmt"
	"strings"

	"github.com/minio/cli"
	"github.com/minio/mc/pkg/console"
)

// removeBucketCmd can be used to run the removebucket compatibility test.
var removeBucketCmd = cli.Command{
	Name:   "removebucket",
	Usage:  "Run the removebucket test.",
	Action: mainRemoveBucket,
	Flags:  append(removeBucketFlags, globalFlags...),
	CustomHelpTemplate: `NAME:
	s3verify {{.Name}} - {{.Usage}}

USAGE:
	s3verify {{.Name}} [FLAGS...]

FLAGS:
	{{range .Flags}}{{.}}
	{{end}}

EXAMPLE:
	1. Test on the Minio server. Note that play.minio.io is a public test server. You are free to use these secret and access keys in all your tests.
		$ S3_URL=https://play.minio.io:9000 S3_ACCESS=Q3AM3UQ867SPQQA43P2F S3_SECRET=zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG s3verify removebucket
	2. Test on the Amazon S3 server using flags. Note that passing access and secret keys as flags should be avoided on a multi-user server for security reasons.
		$ s3verify removebucket --access YOUR_ACCESS_KEY --secret YOUR_SECRET_KEY --url https://s3.amazonaws.com
	`,
}

// Flags supported by the removebucket command.
var removeBucketFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  "help, h",
		Usage: "Help of remove bucket",
	},
}

var (
	removeBucketMessages = []string{"RemoveBucket (Bucket Exists):"}
)

var removeBucketTests = []APItest{mainRemoveBucketExists}

// mainRemoveBucket - Handler for setting up and tearing down a DELETE bucket request.
func mainRemoveBucket(ctx *cli.Context) {
	// TODO: Differentiate errors: s3verify vs Minio vs test failure.
	// Generate a new config.
	config := newServerConfig(ctx)
	s3Client, err := NewS3Client(config.Endpoint, config.Access, config.Secret)
	if err != nil {
		console.Fatalln(err)
	}
	for i, test := range removeBucketTests {
		message := fmt.Sprintf("[%d/%d] "+removeBucketMessages[i], i+1, len(removeBucketTests))
		if err := test(*config, *s3Client, message); err != nil {
			console.Fatalln(err)
		}
		// Print final success message.
		console.Eraseline()
		// Pad the ok to the standard width.
		padding := messageWidth - len([]rune(message))
		console.PrintC(message + strings.Repeat(" ", padding) + "[OK]\n")

	}
}
