# Componentização da Arquitetura por Stacks

O projeto **Data Master** é dividido em múltiplas stacks CloudFormation, cada uma responsável por um conjunto específico de recursos. Essa divisão permite modularidade, reutilização, facilidade de manutenção e deploy incremental.

A ordem de criação das stacks está relacionada às dependências entre elas e não necessariamente segue a sequência das jornadas dos dados.

Abaixo estão descritas todas as stacks utilizadas no projeto:

## Sumário

* [dm-network](#dm-network)
* [dm-roles](#dm-roles)
* [dm-security](#dm-security)
* [dm-storage](#dm-storage)
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

---

## dm-network

**Responsabilidade:**
Provisiona a infraestrutura de rede base do projeto, incluindo VPC, subnets públicas e privadas, internet gateway e route table.

**Principais recursos:**

* **VPC:**
  Faixa `10.0.0.0/16`, com DNS habilitado.

* **Internet Gateway:**
  Gateway de saída para recursos em subnets públicas.

* **Subnets Públicas:**

    * `10.0.1.0/24` (AZ1)
    * `10.0.2.0/24` (AZ2)
      Com IP público automático.

* **Subnets Privadas:**

    * `10.0.11.0/24` (us-east-1a)
    * `10.0.12.0/24` (us-east-1b)
      Sem IP público.

* **Route Table:**
  Rota padrão (0.0.0.0/0) apontando para o Internet Gateway, associada às subnets públicas.

---

## dm-roles

**Responsabilidade:**
Define todas as IAM Roles necessárias para operação dos serviços do projeto, com permissões controladas para Kinesis, Firehose, Glue, DMS, Step Functions, Lambda, ECS, EMR Serverless, Grafana, entre outros.

**Principais recursos:**

* **DMSVpcRole:**
  Permite ao DMS gerenciar VPC e assumir outras roles.

* **PipeExecutionRole:**
  Utilizado por EventBridge Pipes para ler de Kinesis/DynamoDB e escrever no Firehose ou iniciar Step Functions.

* **KinesisAccessRole:**
  Permite ao DMS enviar dados para o Kinesis.

* **FirehoseAccessRole:**
  Usada por Firehose para leitura/escrita em S3, Glue e invocação de funções Lambda.

* **LambdaExecutionRole:**
  Role padrão das funções Lambda, com acesso a Firehose e S3.

* **GlueServiceRole:**
  Role para execução de crawlers e jobs Glue, com acesso total ao Glue e ao Secrets Manager.

* **LakeFormationAccessRole:**
  Acesso à leitura/gravação no S3 via Lake Formation.

* **FirehoseRouterFunctionRole:**
  Lambda intermediária de roteamento para Firehose.

* **ProcessingControllerFunctionRole:**
  Usada pela Lambda que registra o controle de processamento no DynamoDB.

* **ProcessingControllerEventRole:**
  Permite ao EventBridge invocar a função `dm-processing-controller`.

* **WorkerRole:**
  Role utilizada por Lambdas e ECS tasks responsáveis pelo processamento dos dados.

* **StepFunctionRole:**
  Executa jobs no Glue, EMR Serverless, Lambda e ECS, com permissão para criar regras EventBridge.

* **TaskExecutionRole:**
  Role padrão do ECS Fargate para executar containers e acessar logs.

* **EmrServerlessExecutionRole:**
  Usada pelos jobs do EMR Serverless com acesso ao Glue, S3, DynamoDB e logs.

* **GrafanaAccessRole:**
  Role assumida pelo Grafana gerenciado para consultar Athena, Glue, CloudWatch e S3.

---

## dm-security

**Responsabilidade:**
Provisiona os recursos relacionados à segurança de rede, controle de acesso a segredos e repositórios privados para containers Docker utilizados no pipeline de processamento.

**Principais recursos:**

* **Security Groups:**

    * `DBAccessSecurityGroup`: Permite acesso à porta 5432 (PostgreSQL) para o Aurora.
    * `GlueSecurityGroup`: Usado pelos Glue Jobs com tráfego interno liberado.
    * `DMSVpcSecurityGroup`: Associado à instância de replicação do DMS.
    * `ProcessingSecurityGroup`: Associado às tasks ECS Fargate do pipeline.

* **Ingress Rules:**

    * Permissão explícita para que o DMS acesse o Aurora.
    * Tráfego interno liberado entre workers Spark no Glue.

* **Secrets Manager:**

    * `DBSecret`: Contém as credenciais do Aurora PostgreSQL.
    * `DBSecretResourcePolicy`: Garante acesso controlado ao deployer via ARN.

* **ECR Repositories:**

    * Repositórios com política de acesso baseada no ARN do deployer, permitindo push de imagens Docker.

---

## dm-storage

**Responsabilidade:**
Provisiona os buckets S3 utilizados no projeto para ingestão, armazenamento estruturado e observabilidade, além do endpoint privado de acesso ao S3 via VPC.

**Principais recursos:**

* **Buckets S3:**

    * `dm-stage`: Armazena arquivos de entrada (`raw`), tanto por streaming (`firehose`) quanto por batch.
    * `dm-datalake`: Armazena os dados organizados em camadas (`bronze`, `silver`, `gold`). Possui versionamento e bloqueio de acesso público.
    * `dm-observer`: Armazena arquivos relacionados à observabilidade (ex: logs, metadados).
    * `dm-artifacts`: Armazena zips e executáveis utilizados por funções Lambda.

* **VPC Endpoint para S3:**

    * `dm-s3-endpoint`: Cria um endpoint do tipo gateway para permitir acesso privado ao S3 a partir da VPC.

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

  * Instância: `dms.t3.micro`, pública para facilitar testes
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

[Voltar para a página inicial](../README.md#documentação) | [Próximo: Trade-offs e Decisões de Arquitetura](trade-offs.md)