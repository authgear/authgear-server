## Skygear Server Quickstart Image

This shows you how to update the AWS AMI and CloudFormation template
to create the Quickstart Skygear Server.

### Requirements

1. Install [Packer](https://www.packer.io/)
1. Install [AWS CLI](http://docs.aws.amazon.com/cli/)
1. Configure [AWS Credentials](http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html)

### Build images

Before you start building the AMI images, you should make sure that the Skygear
Server for the needed version is already pushed to the Docker Hub. Building
Docker image is not part of this guide.

```
$ packer build -var skygear_version=$SKYGEAR_VERSION template.json
```

where `$SKYGEAR_VERSION` is the version name of the Skygear Server to be
installed. This corresponds to the version name in the Docker Hub image tab.

If you have an alternative AWS credentials profile, run your command like this:

```
$ AWS_PROFILE=profilename packer build -var skygear_version=$SKYGEAR_VERSION teplate.json
```

where `profilename` is your AWS credentials profile defined in credentials file.

### Update CloudFormation template

Modify the CloudFormationt template `cloudformation.template` with the
AMI IDs created by Packer.

You can test the CloudFormation template by running this command:

```
$ aws cloudformation create-stack \
    --region ap-southeast-1 --stack-name stackname \
    --template-body file://`pwd`/cloudformation.template \
    --parameters ParameterKey=Instance,ParameterValue=t2.small \
                 ParameterKey=Data,ParameterValue=30 \
                 ParameterKey=KeyName,ParameterValue=cheungpat \
    --capabilities CAPABILITY_IAM
```

Upload the template with this command

```
$ aws s3 cp cloudformation.template s3://skygear-cf-templates/quickstart/$SKYGEAR_VERSION.template
```
