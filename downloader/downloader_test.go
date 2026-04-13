package downloader

import "testing"

func TestParseURI(t *testing.T) {
	tests := []struct {
		name     string
		uri      string
		cfg      S3Config
		bucket   string
		region   string
		key      string
		filename string
		wantErr  bool
	}{
		{
			name:     "aws regional hostname",
			uri:      "s3://my-bucket.s3.us-west-1.amazonaws.com/path/to/file",
			bucket:   "my-bucket",
			region:   "us-west-1",
			key:      "path/to/file",
			filename: "file",
		},
		{
			name:     "aws global hostname defaults to us-east-1",
			uri:      "s3://my-bucket.s3.amazonaws.com/path/to/file",
			bucket:   "my-bucket",
			region:   "us-east-1",
			key:      "path/to/file",
			filename: "file",
		},
		{
			name: "custom endpoint via s3 scheme",
			uri:  "s3://aptly-dev/ubuntu/dists/stable/InRelease",
			cfg: S3Config{
				EndpointURL:    "https://storage.zigpos.com",
				Region:         "eu-central-1",
				ForcePathStyle: true,
			},
			bucket:   "aptly-dev",
			region:   "eu-central-1",
			key:      "ubuntu/dists/stable/InRelease",
			filename: "InRelease",
		},
		{
			name: "custom endpoint via https scheme",
			uri:  "https://storage.zigpos.com/aptly-dev/ubuntu/dists/stable/InRelease",
			cfg: S3Config{
				EndpointURL:    "https://storage.zigpos.com",
				Region:         "eu-central-1",
				ForcePathStyle: true,
			},
			bucket:   "aptly-dev",
			region:   "eu-central-1",
			key:      "ubuntu/dists/stable/InRelease",
			filename: "InRelease",
		},
		{
			name: "custom endpoint via s3 scheme using endpoint host",
			uri:  "s3://storage.zigpos.com/aptly-dev/ubuntu/dists/stable/InRelease",
			cfg: S3Config{
				EndpointURL:    "https://storage.zigpos.com",
				Region:         "eu-central-1",
				ForcePathStyle: true,
			},
			bucket:   "aptly-dev",
			region:   "eu-central-1",
			key:      "ubuntu/dists/stable/InRelease",
			filename: "InRelease",
		},
		{
			name:    "https uri without endpoint config returns error",
			uri:     "https://storage.zigpos.com/aptly-dev/ubuntu/dists/stable/InRelease",
			wantErr: true,
		},
	}

	for _, test := range tests {
		d := &Downloader{}
		if err := d.Configure(test.cfg); err != nil {
			t.Fatalf("Configure() error: %v", err)
		}

		bucket, region, key, filename, err := d.parseURI(test.uri)
		if test.wantErr {
			if err == nil {
				t.Errorf("%s: parseURI() expected error", test.name)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s: parseURI() unexpected error: %v", test.name, err)
			continue
		}

		if bucket != test.bucket {
			t.Errorf("%s: parseURI() bucket == %s (expected %s)", test.name, bucket, test.bucket)
		}
		if region != test.region {
			t.Errorf("%s: parseURI() region == %s (expected %s)", test.name, region, test.region)
		}
		if key != test.key {
			t.Errorf("%s: parseURI() key == %s (expected %s)", test.name, key, test.key)
		}
		if filename != test.filename {
			t.Errorf("%s: parseURI() filename == %s (expected %s)", test.name, filename, test.filename)
		}
	}
}
