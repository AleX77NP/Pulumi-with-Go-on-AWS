package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/s3"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Create an AWS resource (S3 Bucket)
		bucket, err := s3.NewBucket(ctx, "pulumi-bucket", nil)
		if err != nil {
			return err
		}

		// Create security group
		group, err := ec2.NewSecurityGroup(ctx, "web-secg", &ec2.SecurityGroupArgs{
			Ingress: ec2.SecurityGroupIngressArray{
				ec2.SecurityGroupIngressArgs{
					Protocol: pulumi.String("tcp"),
					FromPort: pulumi.Int(80),
					ToPort: pulumi.Int(80),
					CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
				},
			},
		})
		if err != nil {
			return err
		}

		mostRecent := true
		// Latest ami 
		ami, err := ec2.LookupAmi(ctx, &ec2.LookupAmiArgs{
			Filters: []ec2.GetAmiFilter{
				{
					Name: "name",
					Values: []string{"amzn-ami-hvm-*-x86_64-ebs"},
				},
			},
			Owners: []string{"137112412989"},
			MostRecent: &mostRecent,
		})
		if err != nil {
			return err
		}

		// Create new ec2 instance - free tier 
		srv, err := ec2.NewInstance(ctx, "web-server-pulumi", &ec2.InstanceArgs{
			Tags: pulumi.StringMap{"Name": pulumi.String("web-server-www")},
			InstanceType: pulumi.String("t2-micro"),
			VpcSecurityGroupIds: pulumi.StringArray{group.ID()},
			Ami: pulumi.String(ami.Id),
			UserData: pulumi.String(`#!/bin/bash
			echo "Hello, World!" > index.html
			nohup python -m SimpleHTTPServer 80 &`),
		})

		// Export the name of the bucket
		ctx.Export("bucketName", bucket.ID())
		ctx.Export("publicIp", srv.PublicIp)
		ctx.Export("publicHostName", srv.PublicDns)
		return nil
	})
}
