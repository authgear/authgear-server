# Wildcard AWS ACM cert for `*.skygear.dev`

The cert was created with the following command.

```sh
aws --region us-east-1 acm request-certificate --domain-name '*.skygear.dev' --validation-method DNS
```

# Static Asset

Static asset is hosted on Amazon S3 bucket `code.skygear.dev`. The bucket was created with the command

```sh
aws --region us-east-1 s3api create-bucket --acl public-read --bucket "code.skygear.dev" --no-object-lock-enabled-for-bucket
```

```sh
aws --region us-east-1 s3api put-bucket-policy --bucket "code.skygear.dev" --policy file://policy.json
```

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "s3:GetObject",
            "Effect": "Allow",
            "Resource": "arn:aws:s3:::code.skygear.dev/*",
            "Principal": "*"
        }
    ]
}
```

A cloudfront distribution was created with the command

```sh
aws cloudfront create-distribution --distribution-config file://code.skgyear.dev.config.json
```

where `code.skygear.dev.config.json` is

```json
{
    "CallerReference": "1583126441638",
    "Aliases": {
        "Quantity": 1,
        "Items": [
            "code.skygear.dev"
        ]
    },
    "DefaultRootObject": "index.html",
    "Origins": {
        "Quantity": 1,
        "Items": [
            {
                "Id": "S3-code.skygear.dev",
                "DomainName": "code.skygear.dev.s3.amazonaws.com",
                "OriginPath": "",
                "CustomHeaders": {
                    "Quantity": 0
                },
                "S3OriginConfig": {
                    "OriginAccessIdentity": ""
                }
            }
        ]
    },
    "OriginGroups": {
        "Quantity": 0
    },
    "DefaultCacheBehavior": {
        "TargetOriginId": "S3-code.skygear.dev",
        "ForwardedValues": {
            "QueryString": false,
            "Cookies": {
                "Forward": "none"
            },
            "Headers": {
                "Quantity": 0
            },
            "QueryStringCacheKeys": {
                "Quantity": 0
            }
        },
        "TrustedSigners": {
            "Enabled": false,
            "Quantity": 0
        },
        "ViewerProtocolPolicy": "allow-all",
        "MinTTL": 0,
        "AllowedMethods": {
            "Quantity": 2,
            "Items": [
                "HEAD",
                "GET"
            ],
            "CachedMethods": {
                "Quantity": 2,
                "Items": [
                    "HEAD",
                    "GET"
                ]
            }
        },
        "SmoothStreaming": false,
        "DefaultTTL": 86400,
        "MaxTTL": 31536000,
        "Compress": true,
        "LambdaFunctionAssociations": {
            "Quantity": 0
        },
        "FieldLevelEncryptionId": ""
    },
    "CacheBehaviors": {
        "Quantity": 0
    },
    "CustomErrorResponses": {
        "Quantity": 0
    },
    "Comment": "",
    "Logging": {
        "Enabled": false,
        "IncludeCookies": false,
        "Bucket": "",
        "Prefix": ""
    },
    "PriceClass": "PriceClass_All",
    "Enabled": true,
    "ViewerCertificate": {
        "ACMCertificateArn": "!!!redacted!!!",
        "SSLSupportMethod": "sni-only",
        "MinimumProtocolVersion": "TLSv1",
        "Certificate": "!!!redacted!!!",
        "CertificateSource": "acm"
    },
    "Restrictions": {
        "GeoRestriction": {
            "RestrictionType": "none",
            "Quantity": 0
        }
    },
    "WebACLId": "",
    "HttpVersion": "http2",
    "IsIPV6Enabled": true
}
```
