package cloudstorage

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3Storage struct {
	Bucket    string
	Region    string
	AccessKey string
	SecretKey string

	session *session.Session
	s3      *s3.S3
}

var _ Storage = &S3Storage{}

const (
	S3HeaderAccess = "x-amz-meta-access"
)

var S3ProprietaryToStandardMap = map[string]string{
	"x-amz-meta-accesscontrolalloworigin":      "access-control-allow-origin",
	"x-amz-meta-accesscontrolexposeheaders":    "access-control-expose-headers",
	"x-amz-meta-accesscontrolmaxage":           "access-control-max-age",
	"x-amz-meta-accesscontrolallowcredentials": "access-control-allow-credentials",
	"x-amz-meta-accesscontrolallowmethods":     "access-control-allow-methods",
	"x-amz-meta-accesscontrolallowheaders":     "access-control-allow-headers",
}

var S3StandardToProprietaryMap = map[string]string{
	"access-control-allow-origin":      "x-amz-meta-accesscontrolalloworigin",
	"access-control-expose-headers":    "x-amz-meta-accesscontrolexposeheaders",
	"access-control-max-age":           "x-amz-meta-accesscontrolmaxage",
	"access-control-allow-credentials": "x-amz-meta-accesscontrolallowcredentials",
	"access-control-allow-methods":     "x-amz-meta-accesscontrolallowmethods",
	"access-control-allow-headers":     "x-amz-meta-accesscontrolallowheaders",
}

func NewS3Storage(accessKey string, secretKey string, region string, bucket string) *S3Storage {
	cred := credentials.NewStaticCredentials(accessKey, secretKey, "")
	sess := session.New(&aws.Config{
		Credentials: cred,
		Region:      aws.String(region),
	})
	s3 := s3.New(sess)

	return &S3Storage{
		Region:    region,
		Bucket:    bucket,
		AccessKey: accessKey,
		SecretKey: secretKey,

		session: sess,
		s3:      s3,
	}
}

func (s *S3Storage) PresignPutObject(name string, accessType AccessType, header http.Header) (*http.Request, error) {
	header = s.StandardToProprietary(header)
	header.Set(S3HeaderAccess, string(accessType))

	input := &s3.PutObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(name),
	}

	metadata := map[string]*string{}
	for name := range header {
		lower := strings.ToLower(name)
		if strings.HasPrefix(lower, "x-amz-meta-") {
			metadataName := strings.TrimPrefix(lower, "x-amz-meta-")
			value := header.Get(name)
			metadata[metadataName] = aws.String(value)
		} else {
			switch lower {
			case "content-type":
				input.SetContentType(header.Get(name))
			case "content-disposition":
				input.SetContentDisposition(header.Get(name))
			case "content-encoding":
				input.SetContentEncoding(header.Get(name))
			case "content-length":
				contentLengthStr := header.Get(name)
				contentLength, err := strconv.ParseInt(contentLengthStr, 10, 64)
				if err != nil {
					return nil, err
				}
				input.SetContentLength(contentLength)
			case "content-md5":
				input.SetContentMD5(header.Get(name))
			case "cache-control":
				input.SetCacheControl(header.Get(name))
			}
		}
	}
	input.SetMetadata(metadata)

	req, _ := s.s3.PutObjectRequest(input)
	req.NotHoist = true
	urlStr, _, err := req.PresignRequest(1 * time.Hour)
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	return &http.Request{
		Method: "PUT",
		URL:    u,
		Header: header,
	}, nil
}

func (s *S3Storage) PresignGetObject(name string) (*url.URL, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(name),
	}
	req, _ := s.s3.GetObjectRequest(input)
	req.NotHoist = false
	urlStr, _, err := req.PresignRequest(1 * time.Hour)
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *S3Storage) StandardToProprietary(header http.Header) http.Header {
	return RewriteHeaderName(header, S3StandardToProprietaryMap)
}

func (s *S3Storage) ProprietaryToStandard(header http.Header) http.Header {
	return RewriteHeaderName(header, S3ProprietaryToStandardMap)
}

func (s *S3Storage) AccessType(header http.Header) AccessType {
	a := header.Get(S3HeaderAccess)
	switch a {
	case string(AccessTypePublic):
		return AccessTypePublic
	case string(AccessTypePrivate):
		return AccessTypePrivate
	default:
		return AccessTypePrivate
	}
}
