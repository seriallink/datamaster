package mock

import (
	"github.com/aws/aws-sdk-go-v2/aws"
)

// StackOutputFetcher define um contrato para leitura de outputs do CloudFormation.
type StackOutputFetcher interface {
	GetStackOutput(cfg aws.Config, key string) (string, error)
}

// stackOutputFunc é uma função que implementa StackOutputFetcher.
type stackOutputFunc func(cfg aws.Config, key string) (string, error)

func (f stackOutputFunc) GetStackOutput(cfg aws.Config, key string) (string, error) {
	return f(cfg, key)
}

var stackFetcher StackOutputFetcher

// InjectStackOutputFetcher permite injetar uma implementação personalizada para testes.
func InjectStackOutputFetcher(fn func(cfg aws.Config, key string) (string, error)) {
	stackFetcher = stackOutputFunc(fn)
}

// GetStackFetcher retorna a instância atual para uso no core.
func GetStackFetcher() StackOutputFetcher {
	return stackFetcher
}
