package downloader

import "testing"

func TestParseURI(t *testing.T) {
	tests := []struct {
		uri      string
		bucket   string
		region   string
		key      string
		filename string
	}{
		{
			uri:      "s3://my-bucket.s3-us-west-1.amazonaws.com/path/to/file",
			bucket:   "my-bucket",
			region:   "us-west-1",
			key:      "path/to/file",
			filename: "file",
		},
		{
			uri:      "s3://my-bucket.s3.amazonaws.com/path/to/file",
			bucket:   "my-bucket",
			region:   "us-east-1",
			key:      "path/to/file",
			filename: "file",
		},
		{
			uri:      "s3://my.bucket.s3.amazonaws.com/path/to/file",
			bucket:   "my.bucket",
			region:   "us-east-1",
			key:      "path/to/file",
			filename: "file",
		},
	}

	d := &Downloader{}

	for _, test := range tests {
		bucket, region, key, filename := d.parseURI(test.uri)
		if bucket != test.bucket {
			t.Errorf("parseURI() bucket == %s (expected %s)", bucket, test.bucket)
		}
		if region != test.region {
			t.Errorf("parseURI() region == %s (expected %s)", region, test.region)
		}
		if key != test.key {
			t.Errorf("parseURI() key == %s (expected %s)", key, test.key)
		}
		if filename != test.filename {
			t.Errorf("parseURI() filename == %s (expected %s)", filename, test.filename)
		}
	}
}
