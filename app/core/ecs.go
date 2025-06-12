package core

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

func EnsureECSServiceLinkedRole(cfg aws.Config) error {

	client := iam.NewFromConfig(cfg)

	fmt.Println("Checking for ECS service-linked role...")
	_, err := client.GetRole(context.TODO(), &iam.GetRoleInput{
		RoleName: aws.String("AWSServiceRoleForECS"),
	})
	if err == nil {
		fmt.Println("ECS service-linked role already exists.")
		return nil
	}

	fmt.Println("Creating ECS service-linked role...")
	_, err = client.CreateServiceLinkedRole(context.TODO(), &iam.CreateServiceLinkedRoleInput{
		AWSServiceName: aws.String("ecs.amazonaws.com"),
		Description:    aws.String("Service-linked role for ECS Fargate task execution"),
	})
	if err != nil {
		return fmt.Errorf("failed to create ECS service-linked role: %w", err)
	}

	fmt.Println("ECS service-linked role created.")
	return nil

}
