# Skygear Server Quickstart Example

## Getting started

The simplest way to create a Skygear Server in AWS is using CloudFormation.
You can use the [CloudFormation Console](https://console.aws.amazon.com/cloudformation/home)
or use the [AWS CLI](http://docs.aws.amazon.com/cli/).

Using CloudFormation Console, enter the following URL as the template URL
in the form and follow on-screen instructions.

```
# Copy this to CloudFormation Console
https://skygear-cf-templates.s3.amazonaws.com/quickstart/latest.template
```

Using AWS CLI, you have to [configure AWS credentials](http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html) first, and then enter the following commands:

```
$ aws cloudformation create-stack \
    --region us-east-1 \
    --stack-name "myskygear" \
    --template-url https://skygear-cf-templates.s3.amazonaws.com/quickstart/latest.template \
    --parameters ParameterKey=Instance,ParameterValue=m3.medium \
                 ParameterKey=Data,ParameterValue=30 \
                 ParameterKey=KeyName,ParameterValue=mykeyname \
    --capabilities CAPABILITY_IAM
```

Once your stack is created, your Skygear Server will be up and running at
the Public IP address returned from CloudFormation. Go to `http://<ip address>/`
to confirm that the server is running.

### Control scripts

This example includes a Fabric scripts to remotely control your Skygear Server.
You need to install Fabric.

To restart Skygear Server:

```
$ fab -H ubuntu@<ip address> restart:server
```

(replace `<ip address>` with the public IP address of the server instance)

To upgrade Skygear Server to some version:

```
$ fab -H ubuntu@<ip address> upgrade:v0.6.0
```

Replace `v0.6.0` with the version number of the Skygear Server. To upgrade
to the currently latest version:

```
$ fab -H ubuntu@<ip address> upgrade:latest
```
