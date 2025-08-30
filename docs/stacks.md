# Componentização da Arquitetura por Stacks

O projeto **Data Master** é dividido em múltiplas stacks CloudFormation, cada uma responsável por um conjunto específico de recursos. Essa divisão permite modularidade, reutilização, facilidade de manutenção e deploy incremental.

A ordem de criação das stacks está relacionada às dependências entre elas e não necessariamente segue a sequência das jornadas dos dados.

Abaixo estão descritas todas as stacks utilizadas no projeto:

## Sumário

* [dm-network](#dm-network)
* [dm-security](#dm-security)
* [dm-roles](#dm-roles)
* [dm-storage](#dm-storage)
* [dm-repositories](#dm-repositories)
* [dm-database](#dm-database)
* [dm-catalog](#dm-catalog)
* [dm-governance](#dm-governance)
* [dm-consumption](#dm-consumption)
* [dm-control](#dm-control)
* [dm-functions](#dm-functions)
* [dm-streaming](#dm-streaming)
* [dm-ingestion](#dm-ingestion)
* [dm-processing](#dm-processing)
* [dm-analytics](#dm-analytics)
* [dm-observability](#dm-observability)
* [dm-benchmark](#dm-benchmark)
* [dm-demo](#dm-demo)

---

## dm-network

**Responsabilidade:**
Provisiona a infraestrutura de rede base do projeto, incluindo VPC, subnets públicas e privadas, Internet Gateway e tabela de rotas.

**Principais recursos:**

* **VPC:**
  Faixa `10.0.0.0/16`, com suporte a DNS hostnames e resolução habilitado.

* **Internet Gateway:**
  Responsável pela saída de recursos em subnets públicas para a internet.

* **Subnets Públicas:**

  * `10.0.1.0/24` (us-east-1a)
  * `10.0.2.0/24` (us-east-1b)
    Ambas com atribuição automática de IP público.

* **Subnets Privadas:**

  * `10.0.11.0/24` (us-east-1a)
  * `10.0.12.0/24` (us-east-1b)
    Sem atribuição de IP público.

* **Tabela de Rotas:**
  Rota padrão (`0.0.0.0/0`) apontando para o Internet Gateway, associada às subnets públicas.

---

## dm-security

**Responsabilidade:**
Provisiona os recursos relacionados à **segurança de rede**, **criptografia em repouso** e **controle de acesso a segredos** utilizados no pipeline de processamento de dados.

**Principais recursos:**

### **Security Groups:**

* **DBAccessSecurityGroup**: Permite acesso à porta 5432 (PostgreSQL) para o Aurora.
* **GlueSecurityGroup**: Usado pelos Glue Jobs com tráfego interno liberado.
* **DMSVpcSecurityGroup**: Associado à instância de replicação do DMS.
* **ProcessingSecurityGroup**: Associado às ECS tasks (Fargate) do pipeline de dados.

### **Ingress Rules:**

* Permissão explícita para que o DMS acesse o Aurora PostgreSQL.
* Liberação de tráfego interno entre workers Spark do Glue.

### **Secrets Manager:**

* **DBSecret**: Contém as credenciais do banco Aurora PostgreSQL.
* **DBSecretResourcePolicy**: Restringe o acesso ao secret ao ARN do deployer.

### **KMS Key para criptografia de dados**

* **RawKmsKey**: Chave KMS usada para criptografar os buckets `dm-raw` e `dm-stage`, garantindo proteção dos dados em repouso via **SSE-KMS**.
* **RawKmsAlias**: Alias `alias/dm-raw-kms-key` criado para facilitar a identificação e o uso da chave.
* **Key Policy**: Define acesso granular apenas às roles que realmente precisam usar a chave, como `dm-worker-role` e `dm-firehose-role`.

---

## dm-roles

**Responsabilidade:**
Define todas as IAM Roles necessárias para operação dos serviços do projeto, com permissões controladas para Kinesis, Firehose, Glue, DMS, Step Functions, Lambda, ECS, EMR Serverless, Grafana, entre outros.

**Principais recursos:**

* **DMSVpcRole:**
  Permite ao DMS gerenciar recursos de rede (VPC, ENIs) e assumir outras roles.

* **PipeExecutionRole:**
  Utilizada por EventBridge Pipes para ler de Kinesis ou DynamoDB e escrever no Firehose ou acionar Step Functions.

* **KinesisAccessRole:**
  Permite ao DMS enviar dados para streams Kinesis.

* **FirehoseAccessRole:**
  Usada pelo Firehose para leitura e gravação em S3, acesso ao Glue Catalog e invocação de funções Lambda.

* **LambdaExecutionRole:**
  Role padrão das funções Lambda, com permissões para acessar o Firehose e buckets do S3.

* **GlueServiceRole:**
  Permite a execução de crawlers e jobs do Glue, com acesso ao Glue Catalog e ao Secrets Manager.

* **LakeFormationAccessRole:**
  Concede acesso controlado via Lake Formation para leitura e escrita em S3.

* **FirehoseRouterFunctionRole:**
  Usada pela função Lambda intermediária de roteamento de eventos para diferentes fluxos no Firehose.

* **ProcessingControllerFunctionRole:**
  Lambda responsável por registrar controles de processamento no DynamoDB.

* **ProcessingControllerEventRole:**
  Permite ao EventBridge invocar a função `dm-processing-controller`.

* **WorkerRole:**
  Compartilhada por Lambdas e ECS tasks responsáveis pelo processamento de dados nas camadas bronze, silver e gold.

* **StepFunctionRole:**
  Executa jobs no Glue, EMR Serverless, Lambda e ECS, com permissão para gerenciar regras do EventBridge.

* **TaskExecutionRole:**
  Role padrão para execução de containers no ECS Fargate, com acesso a logs e imagens no ECR.

* **EmrServerlessExecutionRole:**
  Role usada pelos jobs do EMR Serverless com acesso ao Glue, S3, DynamoDB e CloudWatch Logs.

* **GrafanaAccessRole:**
  Assumida pelo Grafana gerenciado para consultar Athena, Glue Catalog, CloudWatch e S3.

---

## dm-storage

**Responsabilidade:**
Provisiona os buckets S3 utilizados para ingestão de dados, armazenamento estruturado por camadas, artefatos de execução e dados de observabilidade. Também cria o endpoint privado para acesso ao S3 via VPC.

**Principais recursos:**

### Buckets S3:

* **`dm-stage`**
  Armazena arquivos da camada `raw/`, recebidos via ingestão **streaming** (ex: Firehose) ou **batch** (ex: cargas de CSV).
  É o **ponto de entrada dos dados no data lake**, com **criptografia em repouso habilitada via SSE-KMS** utilizando a chave `alias/dm-raw-kms-key`.

* **`dm-datalake`**
  Armazena os dados organizados em camadas (`bronze/`, `silver/`, `gold/`).
  Possui versionamento ativado e bloqueio completo de acesso público.

* **`dm-artifacts`**
  Armazena zips, binários e outros artefatos utilizados por funções Lambda e containers ECS.
  Serve como repositório intermediário para deploy automatizado.

* **`dm-observer`**
  Armazena arquivos relacionados à observabilidade, como logs, relatórios e dumps temporários.

* **`dm-costs`**
  Armazena os dados exportados pelo AWS Cost and Usage Report (CUR), com partições por `BILLING_PERIOD`.
  Usado para análise de custos no Athena.

### VPC Endpoint:

* **`dm-s3-endpoint`**
  Endpoint do tipo *Gateway* que permite acesso aos buckets S3 a partir da VPC sem passar pela internet pública.

---

## dm-repositories

**Responsabilidade:**
Provisiona os repositórios privados do Amazon ECR utilizados para armazenar as imagens Docker responsáveis pelas cargas e transformações das camadas bronze, silver, gold e benchmark do projeto.

**Principais recursos:**

### Repositórios ECR:

* **`dm-bronze-ingestor-mass`**
  Armazena a imagem utilizada nas tasks ECS para ingestão em massa da camada *bronze* a partir de arquivos `.csv.gz`.

* **`dm-silver-processor-mass`**
  Contém a imagem com o processamento da camada *silver*, utilizando DuckDB para transformações eficientes.

* **`dm-gold-analytics-mass`**
  Repositório da imagem responsável pela geração das tabelas analíticas da camada *gold*, com joins e agregações.

* **`dm-benchmark-go`**
  Armazena a imagem usada nos testes de benchmark de performance com Go puro, focando em ingestão e escrita em Parquet.

### Controle de acesso:

Todos os repositórios aplicam uma `RepositoryPolicyText` que concede permissões de push para o ARN informado via parâmetro `DeployerArn`, com ações explícitas:

* `ecr:BatchCheckLayerAvailability`
* `ecr:CompleteLayerUpload`
* `ecr:InitiateLayerUpload`
* `ecr:PutImage`
* `ecr:UploadLayerPart`

---

## dm-database

**Responsabilidade:**
Provisiona o cluster Aurora PostgreSQL Serverless v2 com CDC habilitado, utilizado como fonte de dados transacional para ingestão e processamento no projeto.

**Principais recursos:**

* **Aurora PostgreSQL Serverless v2:**

    * Cluster `dm-database-cluster` com CDC ativado (`rds.logical_replication = 1`)
    * Versão: `16.6`
    * Escala entre `0.5` e `1 ACU`
    * Exporte de logs PostgreSQL para o CloudWatch
    * `Performance Insights` e `Data API` habilitados
    * Credenciais gerenciadas via Secrets Manager

* **DB Instance:**

    * Instância writer (`dm-database-instance-writer`)
    * Classe: `db.serverless`
    * Marcada como **pública por ser um ambiente de experimentação e desenvolvimento**

* **Infraestrutura associada:**

    * Subnet Group com subnets públicas
    * Security Group de acesso ao banco
    * Parameter Group customizado com CDC

---

## dm-catalog

**Responsabilidade:**
Cria os databases no Glue Catalog correspondentes às camadas da arquitetura Medallion, utilizados para organização lógica dos dados no data lake.

**Principais recursos:**

* **Glue Databases:**

  * `dm_stage`: Armazena a catalogação dos dados crus recebidos (pré-processamento).
  * `dm_bronze`: Representa os dados brutos organizados por tabela e particionados.
  * `dm_silver`: Dados tratados, validados e enriquecidos.
  * `dm_gold`: Tabelas analíticas, agregadas para consumo final.
  * `dm_costs`: Database para armazenar os dados de custos do projeto.

---

## dm-governance

**Responsabilidade:**
Aplica controles de governança de dados utilizando o AWS Lake Formation, incluindo o registro da localização física do data lake e concessão de permissões para acesso aos bancos de dados do Glue Catalog.

**Principais recursos:**

* **Lake Formation Resource:**

  * Registro do bucket `dm-datalake` como uma localização controlada pelo Lake Formation.

* **Permissões Lake Formation:**

  * Concessão de permissões aos databases no Glue Catalog:

    * `dm_stage`
    * `dm_bronze`
    * `dm_silver`
    * `dm_gold`
    * `dm_costs`

---

## dm-consumption

**Responsabilidade:**
Provisiona os recursos de consulta e exploração de dados no projeto, com foco em análise via AWS Athena.

**Principais recursos:**

* **Athena WorkGroup:**

  * Nome: `dm-athena-workgroup`
  * Usa o Athena Engine v3, com suporte a tabelas Iceberg.
  * Resultado das consultas é armazenado no bucket `dm-observer`.

---

## dm-control

**Responsabilidade:**
Gerencia o controle de processamento de arquivos em **todas as camadas da arquitetura Medallion** (raw, bronze, silver, gold), centralizando o rastreamento de arquivos processados, status, schemas, tabelas e checksums. A tabela serve como fonte única de verdade para orquestração e deduplicação.

**Principais recursos:**

* **Tabela DynamoDB:**

  * Nome: `dm-processing-control`
  * Chave primária: `control_id`
  * Armazena metadados como `object_key`, `schema_name`, `table_name`, `status`, `checksum`.

* **Índices Globais:**

  * `dm-gsi-object-key`: Permite localizar um controle a partir da chave do objeto no S3.
  * `dm-gsi-schema-status`: Usado para identificar arquivos pendentes por schema e status (ex: `bronze:pending`).
  * `dm-gsi-table-checksum`: Garante idempotência ao evitar reprocessamento de arquivos com o mesmo checksum.

* **Stream habilitado:**
  A tabela possui stream do tipo `NEW_IMAGE`, permitindo acionar consumidores como funções Lambda ou Pipes.

---

## dm-functions

**Responsabilidade:**
Provisiona as funções Lambda utilizadas em diferentes pontos da arquitetura, incluindo roteamento de dados, controle de ingestão e transformação de dados da camada raw para bronze.

**Principais recursos:**

* **dm-firehose-router:**
  Utilizada pelo Firehose para determinar dinamicamente o prefixo S3 com base na tabela de origem.

  * Entrada: registros Firehose
  * Saída: objetos `.json.gz` particionados em `raw/` por schema/tabela

* **dm-processing-controller:**
  Registrador de objetos no DynamoDB `dm-processing-control`, utilizado como controle central de ingestão.

  * Entrada: eventos S3
  * Saída: item de controle com status `pending` por tabela/arquivo

* **dm-bronze-ingestor-lite:**
  Responsável por processar os arquivos da camada `raw`, aplicar transformação e escrever em Parquet na camada bronze.

  * Entrada: `object_key` de arquivos `.gz`
  * Saída: arquivos `.parquet` organizados por raw/tabela/

---

## dm-streaming

**Responsabilidade:**
Provisiona todo o pipeline de ingestão por CDC (Change Data Capture), conectando o banco Aurora PostgreSQL ao S3 via DMS, Kinesis e Firehose. É a espinha dorsal da ingestão em tempo real do projeto.

**Principais recursos:**

* **Kinesis Stream:**

  * `dm-kinesis-stream` (modo on-demand)
  * Recebe eventos capturados via DMS.

* **Firehose Delivery Stream:**

  * `dm-firehose-stream` com origem no Kinesis
  * Salva arquivos `.json.gz` no bucket `dm-stage`, usando roteamento dinâmico por Lambda (`dm-firehose-router`)

* **Log Group do Firehose:**

  * Armazena logs de execução da entrega no CloudWatch (`/aws/kinesisfirehose/dm-cdc-stream`)

* **DMS Source e Target Endpoints:**

  * Origem: Aurora PostgreSQL
  * Destino: Kinesis, com `MessageFormat: json`

* **DMS Replication Instance e Task:**

  * Instância: `dms.t3.small`, pública para facilitar testes
  * Task: realiza `full-load-and-cdc` do schema `dm_core` com performance otimizada (`ParallelApplyThreads`, `BatchApplyEnabled`, etc.)

* **Subnets e Grupos de Segurança:**

  * Utiliza subnets públicas para replicação
  * Role associada via `dm-kinesis-role` e `dm-firehose-role`

* **Lambda Permission:**

  * Permite ao Firehose invocar a função `dm-firehose-router`

---

## dm-ingestion

**Responsabilidade:**
Orquestra o processamento dos arquivos da camada `raw` para a camada `bronze`, utilizando uma Step Function que decide dinamicamente se o processamento será feito via Lambda ou ECS Fargate com base no campo `compute_target`.

**Principais recursos:**

* **EventBridge Pipe:**

  * Nome: `dm-ingestion-pipe`
  * Fonte: stream do DynamoDB (`dm-processing-control`)
  * Alvo: Step Function `dm-ingestion-dispatcher`
  * Filtro: apenas registros com `schema_name = dm_bronze` e `eventName = INSERT`

* **Step Function (Express):**

  * Nome: `dm-ingestion-dispatcher`
  * Tipo: `Map` com Choice State para selecionar entre Lambda (`dm-bronze-ingestor-lite`) ou ECS
  * Invoca a função ou task com base no campo `compute_target` (`lambda` ou `ecs`)

* **ECS Fargate (execução sob demanda):**

  * Task Definition `dm-ingestion-task`, com 2 vCPU e 4 GB
  * Container: imagem custom informada por parâmetro (`BronzeIngestorMassImageUri`)
  * Cluster: `dm-ingestion-cluster`
  * Service: `dm-ingestion-service` (sem tasks em execução por padrão)

* **CloudWatch Logs:**

  * Logs da Step Function: `/aws/stepfunctions/dm-ingestion-dispatcher`
  * Logs da task ECS: `/ecs/dm-ingestion-task`

---

## dm-processing

**Responsabilidade:**
Executa o pipeline de transformação dos dados da camada bronze para a camada silver, utilizando EMR Serverless e orquestração via Step Function com agendamento diário.

**Principais recursos:**

* **EMR Serverless Application:**

  * Nome: `dm-processing-app`
  * Tipo: Spark, arquitetura ARM64
  * Auto start/stop com 8 vCPU e 32 GB
  * Acesso à rede via subnets públicas e security group dedicado

* **Step Function (Standard):**

  * Nome: `dm-processing-dispatcher`
  * Recebe lista de tabelas (`brewery`, `beer`, `profile`, `review`)
  * Executa cada uma via EMR Serverless (`main.py`) com `--layer silver --table <tabela>`
  * Após isso, executa o processamento da `review_flat`

* **Agendamento:**

  * EventBridge Rule `dm-processing-cron`
  * Executa diariamente às 6h UTC

* **CloudWatch Logs:**

  * Grupo de logs: `/aws/stepfunctions/dm-processing-dispatcher`

---

## dm-analytics

**Responsabilidade:**
Executa o pipeline de transformação da camada silver para gold, utilizando EMR Serverless e Step Function com agendamento diário. As tabelas da camada gold são utilizadas para dashboards analíticos.

**Principais recursos:**

* **EMR Serverless Application:**

  * Nome: `dm-analytics-app`
  * Tipo: Spark, arquitetura ARM64
  * Auto start/stop com 8 vCPU e 32 GB
  * Acesso à rede via subnets públicas e security group dedicado

* **Step Function (Standard):**

  * Nome: `dm-analytics-dispatcher`
  * Processa as tabelas:

    * `top_beers_by_rating`
    * `top_breweries_by_rating`
    * `state_by_review_volume`
    * `top_styles_by_popularity`
    * `top_drinkers`
  * Cada tabela é processada individualmente via EMR Serverless (`--layer gold --table <tabela>`)

* **Agendamento:**

  * EventBridge Rule `dm-analytics-cron`
  * Executa diariamente às 7h UTC

* **CloudWatch Logs:**

  * Grupo de logs: `/aws/stepfunctions/dm-analytics-dispatcher`

---

## dm-observability

**Responsabilidade:**
Provisiona o ambiente de observabilidade do projeto Data Master com Grafana gerenciado pela AWS, habilitado para integração com Athena e CloudWatch.

**Principais recursos:**

* **Grafana Workspace:**

  * Nome: `dm-grafana-workspace`
  * Tipo: AWS Managed Grafana
  * Permissões gerenciadas (`SERVICE_MANAGED`)
  * Plugins administrativos habilitados
  * Acesso concedido via `dm-grafana-access-role-arn`
  * Fontes de dados integradas:

    * **Athena** (consulta aos dados do Lakehouse)
    * **CloudWatch** (monitoramento de logs e métricas)

---

## dm-benchmark

**Responsabilidade:**
Provisiona toda a infraestrutura necessária para executar benchmarks comparando performance entre containers Go (via ECS Fargate) e jobs PySpark (via Glue). A stack inclui armazenamento, catálogo de dados, execução em ECS e Glue, além de logs e outputs.

**Principais recursos:**

* **Bucket S3:**

  * `dm-benchmark-<account-id>`: Armazena arquivos de entrada (`input/`), saída (`output/`), resultados (`results/`) e scripts (`scripts/`) utilizados nos benchmarks.

* **Glue Catalog:**

  * `dm_benchmark`: Glue Database que representa o catálogo lógico para benchmarks, apontando para os dados armazenados no bucket `dm-benchmark`.

* **ECS:**

  * `dm-benchmark-cluster`: Cluster Fargate usado para execução sob demanda dos testes em Go.
  * `dm-benchmark-task`: Task Definition com 2 vCPU e 4 GB de memória, baseada na imagem Docker fornecida via parâmetro.
  * `dm-benchmark-service`: Serviço ECS com `DesiredCount: 0`, configurado para execução sob demanda dos benchmarks.

* **IAM Roles:**

  * `dm-benchmark-ecs-task-execution`: Role para execução da task ECS com permissões gerenciadas padrão (`AmazonECSTaskExecutionRolePolicy`).

* **CloudWatch Logs:**

  * `/ecs/dm-benchmark-task`: Grupo de logs com retenção de 7 dias, utilizado para registrar a saída dos benchmarks em Go.

* **Glue Job:**

  * `dm-benchmark-glue`: Job PySpark (Glue 4.0) com 2 workers (G.1X), configurado para comparar performance com base nos mesmos arquivos `.csv.gz`.
    Utiliza variáveis de ambiente para receber os parâmetros `BENCHMARK_BUCKET`, `BENCHMARK_INPUT`, `BENCHMARK_OUTPUT` e `BENCHMARK_RESULT`.

---

## dm-demo

**Responsabilidade:**
Provisiona um ambiente de demonstração com um usuário IAM restrito, que pode consultar os dados da camada *gold* via Athena. Também cria um bucket dedicado para armazenar os resultados das queries.

**Principais recursos:**

* **IAM User:**

  * `dm-demo-user`: Usuário IAM com permissão de leitura sobre o catálogo `dm_gold`, acesso ao Athena e ao bucket `dm-datalake`, e permissão de escrita em um bucket exclusivo para resultados de queries.

* **Bucket S3:**

  * `dm-demo-<account-id>`: Bucket exclusivo para armazenar os resultados das queries executadas via Athena pelo usuário `dm-demo-user`.

* **Permissões Athena:**

  * Execução de queries nos workgroups `primary` e `dm-athena-workgroup`.
  * Leitura de resultados e metadados do catálogo de dados (`AwsDataCatalog`).
  * Escrita dos resultados no bucket `dm-demo`.

* **Permissões Glue:**

  * Acesso de leitura ao Glue Catalog (`dm_gold`), incluindo databases, tabelas e partições.

* **Permissões S3:**

  * Leitura dos dados em `dm-datalake-<account-id>/gold/*`.
  * Escrita e leitura no bucket `dm-demo-<account-id>`.
  * Permissão genérica `ListAllMyBuckets` para compatibilidade com o console do Athena.

* **Permissão Lake Formation:**

  * Permissão `SELECT` concedida via `AWS::LakeFormation::Permissions` sobre todas as tabelas do database `dm_gold`.

---

[Voltar para a página inicial](../README.md#documentação) | [Próximo: Trade-offs e Decisões de Arquitetura](trade-offs.md)