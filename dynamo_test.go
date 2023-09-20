package testcontainers_ddb_example_go

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/localstack"
	"log"
	"os"
	"testing"
)

func TestStart(t *testing.T) {
	t.Run("successful", func(t *testing.T) {
		ctx := context.Background()

		// disabling Reaper as issues with Podman local configuration
		//os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

		// to be able to run the tests with Podman (might require `podman machine set --rootful`)
		// cf https://github.com/testcontainers/testcontainers-dotnet/issues/876
		os.Setenv("TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED", "true")              // needed to run RYUK
		os.Setenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE", "/var/run/docker.sock") // needed to apply the bind with statfs

		localstackContainer, err := localstack.RunContainer(ctx,
			testcontainers.WithImage("localstack/localstack:1.4.0"),
		)
		if err != nil {
			log.Fatalf("localstack.RunContainer: %v", err)
		}

		// Clean up the container
		defer func() {
			if err := localstackContainer.Terminate(ctx); err != nil {
				log.Fatalf("localstackContainer.Terminate: %v", err)
			}
		}()

		port, err := localstackContainer.MappedPort(ctx, "4566/tcp")
		if err != nil {
			log.Fatalf("localstackContainer.MappedPort: %v", err)
		}

		// Retrieving from local config dynamo client
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...any) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL: fmt.Sprintf("http://localhost:%s", port),
			}, nil
		})

		awsConf, err := awsConfig.LoadDefaultConfig(context.TODO(), awsConfig.WithEndpointResolverWithOptions(customResolver))
		if err != nil {
			log.Fatalf("awsConfig.LoadDefaultConfig with resolver: %v", err)
		}
		dynamoCli := dynamodb.NewFromConfig(awsConf)
		if dynamoCli == nil {
			log.Fatalf("dynamodb.NewFromConfig: %v", err)
		}
		if err = createTable(dynamoCli); err != nil {
			log.Fatalf("p.CreateTable: %v", err)
		}

	})

}

func createTable(dynamoCli *dynamodb.Client) error {
	input := dynamodb.CreateTableInput{
		TableName: aws.String("testcontainers"),
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("PK"), KeyType: types.KeyTypeHash},
			{AttributeName: aws.String("SK"), KeyType: types.KeyTypeRange},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("PK"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("SK"), AttributeType: types.ScalarAttributeTypeS},
		},
		BillingMode: types.BillingModePayPerRequest,
	}

	_, err := dynamoCli.CreateTable(context.TODO(), &input) // this is async operation
	if err != nil {
		return fmt.Errorf("d.CreateTable[%s]: %w", *input.TableName, err)
	}

	return nil
}
