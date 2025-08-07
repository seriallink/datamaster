# 10. Referência Técnica do Projeto 

Esta seção consolida a **referência técnica completa** do projeto **Data Master**, com foco nos principais componentes implementados em Go e Python, incluindo o CLI, funções Lambda, scripts de processamento e utilitários.

Toda a documentação foi gerada a partir de comentários no padrão das linguagens (GoDoc e docstrings Python), organizados por módulo.

---

## Estrutura do Projeto

O repositório é organizado em módulos independentes que refletem as camadas da arquitetura e as responsabilidades do projeto. A estrutura favorece manutenibilidade, escalabilidade e reprodutibilidade.

```
.
├── app/                        # Código-fonte da aplicação Data Master (Go)
│   ├── bronze/                 # Modelos Parquet da camada bronze
│   ├── cmd/                    # Comandos CLI (deploy, catalog, process, etc.)
│   ├── core/                   # Núcleo reutilizável: AWS SDK, controle, etc.
│   ├── dialect/                # Dialetos de SQL para Athena, DuckDB, etc.
│   ├── enum/                   # Constantes e tipos enumerados
│   ├── help/                   # Mensagens de ajuda e instruções interativas
│   ├── misc/                   # Funções auxiliares (genéricas)
│   ├── model/                  # Modelos representando os dados da origem 
│   └── tests/                  # Testes automatizados da CLI
├── artifacts/                  # Artefatos compilados (zips/tars para deploy)
├── assets/                     # Imagens utilizadas na documentação
├── catalog/                    # Scripts auxiliares para Glue Catalog
├── dashboards/                 # Dashboards exportadas do Grafana (JSON)
├── database/                   # Modelos e scripts SQL do banco de origem (Aurora)
│   ├── mer/                    # Modelo Entidade-Relacionamento (MER) 
│   └── migrations/             # Scripts de criação de schemas e extensões
├── diagrams/                   # Diagramas arquiteturais (Draw.io e PNGs)
├── docs/                       # Documentação técnica em Markdown
├── scripts/                    # Scripts Python usados no pipeline
│   ├── core/                   # Utilitários reutilizáveis
│   ├── silver/                 # Transformações da camada silver
│   └── gold/                   # Transformações da camada gold
├── stacks/                     # Templates CloudFormation (IaC)
├── workers/                    # Workers Go para ECS ou Lambda 
│   ├── benchmark-go/           # Worker de benchmark de performance
│   ├── bronze-ingestor-lite/   # Ingestão leve da camada bronze (Lambda)
│   ├── bronze-ingestor-mass/   # Ingestão massiva da camada bronze (ECS)
│   ├── firehose-router/        # Processador de eventos Kinesis/Firehose
│   └── processing-controller/  # Worker para orquestração baseada no DynamoDB
└── main.go                     # Ponto de entrada da CLI
```

---

## Como visualizar a documentação (GoDoc)

### 1. Instale o [GoDoc](https://go.dev/doc/comment) (caso ainda não tenha)

```
go install golang.org/x/tools/cmd/godoc@latest
````

### 2. Execute o servidor de documentação

Certifique-se de estar na **raiz do projeto**, onde se encontra o arquivo `go.mod`. Em seguida, execute o comando:

```bash
godoc -http=:6060
```

### 3. Acesse pelo navegador

Abra o endereço abaixo em seu navegador:

```
http://localhost:6060/pkg/github.com/seriallink/datamaster/
```

Esse caminho funcionará em qualquer máquina, desde que o repositório tenha sido clonado com suporte a Go Modules (o projeto já inclui o `go.mod` com o path correto).

### 4. O que você encontrará nesta documentação

Esta página apresenta a **documentação gerada automaticamente** do projeto Data Master, organizada por pacotes Go.

Você encontrará:

* Descrição completa dos **módulos da aplicação** (`app`, `core`, `cmd`, `model`, etc.)
* **Workers Go** para ECS e Lambda documentados separadamente (`benchmark-go`, `bronze-ingestor-mass`, etc.)
* Comentários detalhados com **tipos, funções e comportamentos esperados** de cada componente
* Visão geral dos principais fluxos da plataforma e como cada pacote se encaixa na arquitetura

> A documentação cobre tanto a camada de **infraestrutura e orquestração** (CLI, core, catalog) quanto os **processadores e utilitários** de ingestão, transformação e benchmark.

---

## Visualização da Documentação Python com `pydoc`

A documentação dos scripts Python pode ser visualizada localmente de forma rápida usando o `pydoc`, que já vem embutido no Python.

---

### 1. Verifique se o Python está funcionando corretamente

No terminal, execute:

```bash
python --version
```

Você deve ver algo como:

```
Python 3.xx.x
```

### 2. Navegue até a raiz do projeto

Para que o `pydoc` consiga importar os módulos corretamente, execute os comandos a partir da **raiz do projeto**, onde está a pasta `scripts/`:

```bash
cd $GOPATH/src/github.com/seriallink/datamaster/scripts
```

### 3. Instale as dependências do projeto

Antes de rodar o `pydoc`, é necessário instalar todas as bibliotecas usadas pelos scripts Python:

```bash
pip install -r requirements.txt
```

### 4. Inicie o servidor local do `pydoc`

Execute o servidor HTTP que exibe os módulos documentados:

```bash
python -m pydoc -p 6061
```

> Use `6061` para evitar conflito com o GoDoc, que roda na porta `6060`.

### 5. Acesse a documentação no navegador

Abra no navegador: [http://localhost:6061](http://localhost:6061)

Você verá todos os pacotes Python do projeto, incluindo:

* **`core`**: Funções utilitárias reutilizadas em múltiplas camadas (ex: leitura de controle, acesso ao S3, etc.)
* **`silver`**: Scripts de transformação da camada bronze para silver, incluindo joins, enrichments de datas
* **`gold`**: Scripts para geração de outputs analíticos, como agregações, rankings e materializações finais

---

[Voltar para a página inicial](../README.md#documentação) | [Próxima seção: Provisionamento do Ambiente](provisioning.md)