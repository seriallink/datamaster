# Ingestão de Dados (Bronze)

A ingestão da camada Bronze transforma os arquivos `.gz` da camada Raw em arquivos colunarizados no formato **Parquet**, com schema definido e partições organizadas no padrão Hive. Esse processo é executado por jobs em **Go** (alta performance), e orquestrado de forma automatizada por eventos.

### Visão Geral do Fluxo

* **Origem**: dados chegam na **camada Raw** como arquivos `.json.gz` ou `.csv.gz`, provenientes de:

    * **CDC Streaming**: `DMS → Kinesis → Firehose`
    * **Batch**: uploads manuais ou cargas em lote
* **Destino**: arquivos `.parquet` na **camada Bronze**.

O processamento é controlado por registros no **DynamoDB (ProcessingControl)**, garantindo rastreabilidade, reprocessamento e controle de tentativas.

---

## Etapas de Execução

### 1. Executar o comando `seed`

O primeiro passo para validar o pipeline de ingestão é popular o banco de dados Aurora e a camada raw com dados de exemplo. Isso pode ser feito com o comando:

```
>>> seed
You are about to seed all available datasets from GitHub.
Type 'go' to continue: go
```

Esse comando realiza duas ações principais:

* **Insere registros nas tabelas `brewery`, `beer` e `profile`** diretamente no Aurora, permitindo o teste do fluxo de **streaming CDC** via DMS → Kinesis → Firehose.
* **Envia um arquivo `.gz` para o S3** na estrutura da camada raw da tabela `review`, simulando o fluxo de **ingestão batch**.

Exemplo de saída:

```
→ Seeding 'brewery'...
→ Seeding 'beer'...
→ Seeding 'profile'...
→ Seeding 'review'...
Uploaded batch file to s3://dm-stage-<account-id>/raw/review/dm-batch-process-1-<datetime>-<uuid>.gz
Seeding completed successfully.
```

---

### 2. Verificar arquivos na camada Raw (S3)

Após a execução do `seed`, os dados são enviados para a **camada Raw**, no bucket `dm-stage-<account_id>`, organizados por tabela:

* `raw/beer/`
* `raw/brewery/`
* `raw/profile/`
* `raw/review/`

Cada pasta contém arquivos `.gz` com os dados brutos, sem tratamento. A criação dos arquivos varia conforme o modo de ingestão:

* **Streaming (beer, brewery, profile)**:

    * Os eventos são capturados via DMS → Kinesis → Firehose.
    * O **Firehose armazena os arquivos `.gz` de forma assíncrona**, conforme regras internas de buffer (tempo ou tamanho).
    * **Pode levar alguns minutos para os arquivos aparecerem** após o `seed`, exatamente devido ao comportamento do buffer do Firehose.
* **Batch (review)**:

    * O arquivo `.gz` é carregado diretamente no S3, de forma imediata, durante o `seed`.

#### Verificação via Console

1. Acesse o bucket `dm-stage-<account_id>` no console da AWS.
2. Navegue até a pasta `raw/`.
3. Verifique a criação das subpastas `beer/`, `brewery/`, `profile/` e `review/`.
4. Dentro de cada pasta, localize os arquivos `.gz`.

---

### 3. Verificar os controles criados no DynamoDB

Após a chegada de um novo arquivo `.gz` na camada Raw, um **evento de notificação** do S3 é automaticamente disparado. Esse evento está configurado no bucket `dm-stage-<account_id>` para objetos que atendem ao filtro:

* **Prefixo**: `raw/`
* **Sufixo**: `.gz`

Esse evento ativa a função Lambda `dm-processing-controller`, responsável por **criar dinamicamente uma entrada de controle** na tabela `dm-processing-control` do DynamoDB.

#### Verificando a tabela

1. Acesse o **DynamoDB** no console da AWS.
2. No menu lateral, clique em **Tables**.
3. Selecione a tabela `dm-processing-control`.
4. Clique em **Explore table items** (botão no canto superior direito).
5. Adicione um filtro por `schema_name` com valor `dm_bronze` e clique em **Run**.

Você verá um item para cada arquivo `.gz` recebido na Raw, com os metadados e status do processamento.

#### Estrutura do item de controle

Cada item na tabela possui os seguintes campos:

| Campo            | Tipo     | Descrição                                                           |
| ---------------- | -------- |---------------------------------------------------------------------|
| `control_id`     | `string` | Identificador único do controle (UUID v4).                          |
| `schema_name`    | `string` | Nome lógico do schema de destino (ex: `dm_bronze`).                 |
| `table_name`     | `string` | Nome da tabela de destino (ex: `beer`).                             |
| `object_key`     | `string` | Caminho completo do arquivo no bucket (ex: `raw/beer/file.gz`).     |
| `file_format`    | `string` | Tipo do arquivo (`json` ou `csv`).                                  |
| `file_size`      | `number` | Tamanho do arquivo, em bytes.                                       |
| `record_count`   | `number` | Quantidade de registros processados.                                |
| `checksum`       | `string` | SHA256 do conteúdo do arquivo, usado para verificação.              |
| `status`         | `string` | Status do processamento (`pending`, `running`, `success`, `error`). |
| `attempt_count`  | `number` | Quantidade de tentativas de processamento.                          |
| `compute_target` | `string` | Target de execução (`lambda` ou `ecs`).                             |
| `created_at`     | `string` | Timestamp da criação do item de controle.                           |
| `started_at`     | `string` | Timestamp de início do processamento.                               |
| `finished_at`    | `string` | Timestamp de término do processamento.                              |
| `duration`       | `number` | Tempo total em milissegundos.                                       |
| `updated_at`     | `string` | Última atualização do item.                                         |

> Como o pipeline é 100% orientado a eventos, em condições normais todos os arquivos `.gz` já terão sido processados no momento dessa verificação, ou seja, os itens no DynamoDB estarão com o `status = success`, pois o processamento é executado em poucos segundos (ou menos).

---

### 4. Início automático do processamento (DynamoDB Stream + Step Functions)

A tabela `dm-processing-control` está configurada com **DynamoDB Streams** ativado, no modo **"New image"**. Isso significa que **cada novo item de controle criado dispara um evento**, que é usado para acionar automaticamente o pipeline de ingestão da camada Bronze.

#### Pipeline de ingestão (Step Function)

Quando o Stream detecta um novo item:

* Ele envia o evento para a Step Function **`dm-ingestion-dispatcher`**, do tipo **Express**.
* A Step Function verifica o campo `compute_target` do item:

    * Se `lambda`: aciona o processo leve de ingestão via AWS Lambda.
    * Se `ecs`: aciona um container no ECS otimizado para grandes volumes.

Essa escolha é feita automaticamente com base no campo `record_count` do item, arquivos com mais de **100 mil linhas** são atribuídos ao ECS, garantindo performance sem sobrecarregar funções Lambda.

#### Como verificar as execuções

1. Acesse o serviço **Step Functions** no console da AWS.
2. No menu esquerdo, clique em **State machines**.
3. Selecione `dm-ingestion-dispatcher`.
4. Vá até a aba **Monitoring**.

Você verá gráficos com as seguintes métricas:

* **Executions started**: número de execuções iniciadas.
* **Executions succeeded**: execuções concluídas com sucesso.
* **Executions errors**: falhas no processamento (caso existam).

O número de `executions started` deve corresponder ao número de itens criados na tabela `dm-processing-control` após o comando `seed`.

---

### 5. Verificar os arquivos Parquet gerados (camada Bronze)

Com o pipeline executado com sucesso, os arquivos `.gz` da camada Raw são processados e convertidos para **formato Parquet** no bucket `dm-datalake-<account_id>`, dentro da pasta `bronze/`.

#### Estrutura gerada

A camada Bronze agora utiliza **particionamento Hive-style** baseado na data de criação do item (`created_at`), com a seguinte estrutura:

```
bronze/<tabela>/year=YYYY/month=MM/day=DD/<arquivo>.parquet
```

Exemplo real para a tabela `beer`:

```
bronze/beer/year=2025/month=07/day=21/dm-firehose-stream-1-2025-07-21-18-31-35-e7ee284d-aa35-3442-8a31-a7a3085e2f13.parquet
```

Cada arquivo `.parquet` corresponde diretamente a um `.gz` original da camada Raw, mantendo o nome base para garantir rastreabilidade.

#### Como verificar via console

1. Acesse o console do S3.
2. Vá até o bucket `dm-datalake-<account_id>`.
3. Navegue até a pasta `bronze/`.
4. Entre em uma das tabelas, como `beer/`.
5. Abra as pastas de partição (`year=.../month=.../day=.../`) e verifique os arquivos `.parquet` correspondentes.

> A presença desses arquivos confirma que a transformação para Parquet ocorreu com sucesso e que a camada Bronze está estruturada para suportar leitura eficiente com engines compatíveis com particionamento Hive (Athena, Glue, Spark, etc.).

---

### 6. Verificação final via Athena

Com os arquivos `.parquet` gerados e organizados em partições Hive na camada Bronze, a consulta pode ser feita diretamente pelo **Athena**, utilizando o Glue Catalog.

#### Passos para validar

1. Acesse o serviço **Athena** no console da AWS.
2. No menu lateral, clique em **Query editor**.
3. No canto superior direito, clique no botão do **workgroup atual** (geralmente `primary`).
4. Selecione o workgroup **`dm-athena-workgroup`**. Ao fazer isso, será exibido um modal de confirmação informando as configurações aplicadas ao workgroup, incluindo o local de output dos resultados (`s3://dm-observer-<account_id>/athena/results/`). Basta clicar em **"Acknowledge"** para confirmar e prosseguir com a execução da query.
5. No painel esquerdo, selecione o database **`dm_bronze`**.
6. Localize a tabela **`beer`**.
7. Execute a seguinte query:
```sql
SELECT * FROM dm_bronze.beer LIMIT 10;
```

#### Resultado esperado

A consulta deve retornar 10 registros da tabela `beer`, confirmando que:

* Os arquivos `.parquet` estão corretamente organizados.
* O particionamento Hive foi reconhecido automaticamente.
* A tabela foi registrada corretamente no Glue Catalog.

> Essa é a validação final do pipeline Raw → Bronze.

> Caso veja o erro `No output location provided`, volte e certifique-se de que o **workgroup selecionado é `dm-athena-workgroup`** e que o botão **"Acknowledge"** foi clicado.

---

[Voltar para a página inicial](../README.md#documentação) | [Próximo: Processamento de Dados (Silver)](processing.md)