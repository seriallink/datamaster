# Processamento de Dados (Silver)

A camada **silver** transforma os dados brutos da bronze em registros **normalizados, enriquecidos e prontos para análise**, utilizando o formato transacional **Iceberg**. Esse processamento é responsável por aplicar regras de negócio, compor dimensões com joins e manter um histórico consistente via `INSERT`, `UPDATE` e `DELETE`.

## Visão Geral da Arquitetura

O pipeline é totalmente automatizado e executado diariamente via **Step Function** (`dm-processing-dispatcher`), acionada por um evento agendado do **EventBridge** (`cron` diário às 6h UTC). O fluxo é dividido em duas etapas principais:

1. **Processamento em Paralelo**
   As tabelas `brewery`, `beer`, `profile` e `review` são processadas individualmente e de forma isolada por jobs Spark executados no **EMR Serverless**, com recursos otimizados (8 vCPU, 32 GB RAM). O código de transformação está empacotado em um arquivo `.zip` e executado via `spark-submit`, com argumentos dinâmicos por tabela.

2. **Geração da Tabela review_flat**
   Após a execução das tabelas principais, uma nova job Spark constrói a `review_flat`, consolidando dados de `review`, `beer`, `profile` e `brewery` em um modelo unificado e analítico.

O controle de execução é feito diretamente pela Step Function, com registros de logs armazenados no S3 (`dm-observer-bucket`), e a arquitetura é altamente resiliente: falhas em qualquer job são capturadas e interrompem o fluxo com mensagens de erro claras para depuração.

---

## Frequência e Estratégia de Processamento

O pipeline da camada silver foi originalmente idealizado para ser executado em **tempo real**, refletindo continuamente as alterações recebidas da camada bronze. No entanto, após validações práticas e testes de custo, performance e robustez, optou-se por adotar uma estratégia de **processamento em lote diário**. Essa decisão foi motivada por uma série de fatores técnicos e operacionais:

* **Custo-benefício**: Manter jobs Spark disparados com frequência elevada (event-driven) no EMR Serverless se mostrou custoso, especialmente considerando a granularidade dos eventos e a ociosidade entre execuções.
* **Natureza dos dados**: A maioria dos dados da plataforma é utilizada para análises históricas, não havendo demanda real por atualização em tempo real.
* **Robustez e controle transacional**: A execução em lote, via Step Function, permite melhor rastreabilidade, controle de falhas, auditoria dos dados processados e reprocessamentos consistentes.
* **Dependência de joins entre tabelas**: A geração de estruturas enriquecidas como `review_flat` depende de múltiplas dimensões (`review`, `beer`, `profile`, `brewery`) estarem completamente carregadas — algo inviável em um fluxo contínuo com baixa latência.
* **Melhor aproveitamento da partição temporal**: O uso de partições como `review_year` e `review_month` é mais eficiente em modelos batch, com menor fragmentação e overhead de gerenciamento.

Essa mudança estratégica privilegiou a **governança, consistência e otimização de custos**, ao mesmo tempo em que mantém a arquitetura escalável e pronta para evoluir. O pipeline é executado diariamente via cron agendado no EventBridge, garantindo uma janela de atualização previsível e confiável para os dados da camada silver.

---

## Como verificar arquivos pendentes para processamento (bronze → silver)

Ao final do pipeline da camada bronze, cada arquivo Parquet gerado é registrado no **DynamoDB**, na tabela `dm-processing-control`, indicando que está **pronto para ser processado na camada silver**. Esse controle garante rastreabilidade, tentativas, enriquecimento com metadados e consistência entre as camadas.

Esses registros possuem `schema_name = dm_silver` e `status = "pending"`.

### Passos para consulta manual:

1. Acesse o serviço **DynamoDB** no AWS Console.

2. No menu lateral, clique em **Tables**, depois selecione a tabela `dm-processing-control`.

3. No topo direito, clique em **Explore table items**.

4. Aplique o seguinte filtro:

   * **Attribute name**: `schema_name`
   * **Condition**: `Equal to`
   * **Type**: `String`
   * **Value**: `dm_silver`

5. Clique em **Run**.

Isso exibirá os arquivos que estão **aguardando processamento** para a silver, um por linha.

### Observações técnicas

* O campo `compute_target` agora é definido como `"emr"`, indicando que os arquivos serão processados por jobs Spark no **EMR Serverless**.
* O campo `object_key` aponta para a localização do arquivo Parquet na bronze.
* Os campos `attempt_count`, `status`, `created_at` e `table_name` continuam sendo utilizados exatamente como no fluxo raw → bronze.

> Isso garante **uniformidade de controle** entre as camadas e facilita a reusabilidade do código e das estratégias de monitoramento.

---

## Processamento Manual da Camada Silver

Durante a fase de testes, nem sempre é viável esperar pelo agendamento do cron para acionar o processamento da camada **silver**. Por isso, foi criada uma funcionalidade no CLI que permite forçar essa execução manualmente:

### Comando

```
>>> process --layer silver
```

Esse comando dispara a execução **manual** da Step Function responsável pelo processamento, delegando a execução com base nos arquivos pendentes no DynamoDB (status `pending`).

### Importante!

* O comando **não é síncrono**. Ele apenas inicia a execução da Step Function e **retorna imediatamente**.
* Pode levar alguns segundos até que a execução apareça na interface do Step Functions.

## Como verificar se a execução foi iniciada

1. Acesse o serviço **Step Functions** na AWS Console.
2. No menu lateral esquerdo, clique em **State machines**.
3. Clique na state machine chamada **`dm-processing-dispatcher`**.
4. Vá até a aba **Executions**.
5. Procure por uma execução com nome similar a ``manual-<date>-<time>``.
6. Clique sobre a execução para visualizar os detalhes.
7. Você pode acompanhar o progresso da execução em tempo real até o status mudar para `Succeeded` (ou `Failed`, caso haja erro).

---

## Acompanhando o processamento de cada tabela (EMR)

Cada etapa da execução corresponde a uma tabela sendo processada via EMR Serverless. Para verificar os detalhes:

1. Acesse o serviço **EMR** na AWS Console.
2. No menu esquerdo, clique em **EMR Serverless**.
3. Clique em **Manage applications**.
4. No primeiro acesso, será necessário clicar em **Create and launch EMR Studio** — apenas confirme.
5. A EMR Studio abrirá em uma nova aba.
6. Em **Applications**, clique em **`dm-processing-app`**.
7. Em **Batch job runs**, você verá a lista de execuções por tabela.
8. Clique no **Job run ID** de interesse para ver os detalhes completos:

   * Tempo de execução
   * Arguments (ex: `--layer silver --table beer`)
   * Logs
   * Uso de recursos (CPU/memória)

---

### Exemplo de execução manual

```
>>> process --layer silver
You are about to trigger processing for all tables in layer: silver
Type 'go' to continue: go
Processing started successfully.
```

---

## Verificando os dados na camada Silver

Após o processamento manual ou automático da camada **silver**, os dados são salvos no bucket de dados no formato **Iceberg**, já organizados com boas práticas de particionamento e tratamento de small files.

### Caminho do bucket

Os dados ficam armazenados no bucket:

```
s3://dm-datalake-<account_id>/silver/
```

Na raiz do bucket você verá:

* `bronze/`
* `silver/` ← (essa pasta aparece após a primeira execução bem-sucedida)

### Estrutura da camada Silver

A estrutura de diretórios segue a mesma da bronze:

```
silver/
├── beer/
├── brewery/
├── profile/
├── review/
└── review_flat/   ← tabela nova, gerada a partir de joins
```

Dentro de cada tabela, os dados são armazenados em partições por data de processamento:

```
silver/beer/data/partitioned_at=YYYYMMDD/
```

Cada partição contém **um único arquivo `.parquet`**, mesmo que múltiplos arquivos tenham sido gerados na bronze.

> Essa compactação diária é uma estratégia para mitigar o problema de **small files**, comum em arquiteturas com ingestão contínua. O dado é consolidado por dia de processamento.

## Formato Iceberg

A camada **silver** já adota o formato **Apache Iceberg**, que traz os seguintes benefícios:

* **Gerenciamento de partições automático**, sem a necessidade de reprocessar metadados manualmente.
* Suporte a **schemas evolutivos**, permitindo adicionar/remover colunas sem recriar a tabela.
* Operações transacionais (append, update, delete) consistentes.
* Alta performance de leitura com otimizações como **pruning de partição**.

A estrutura de cada tabela Iceberg segue o padrão:

```
silver/<table_name>/
├── metadata/        ← arquivos internos do Iceberg (não mexer)
├── data/            ← diretório onde ficam os dados Parquet
│   └── partitioned_at=YYYYMMDD/
```

## Visualização e verificação

Você pode verificar os arquivos manualmente via AWS Console:

1. Vá até o serviço **S3**.
2. Acesse o bucket **`dm-datalake-<account_id>`**.
3. Navegue até a pasta `silver/beer/data/partitioned_at=YYYYMMDD/`.
4. Você verá um único arquivo `.parquet`.

Caso queira, repita para outras tabelas para confirmar o processamento completo.

---

Excelente, Marcelo. Aqui está a última etapa da verificação da **camada silver** usando o **Athena**, garantindo que os dados foram processados corretamente, enriquecidos com as datas e integrados ao Glue Catalog com formato Iceberg:

---

## Verificação via Athena

Após o processamento da **camada silver**, os dados ficam disponíveis para consulta imediata no **Amazon Athena**, pois as tabelas são registradas no **AWS Glue Catalog** com suporte total a **Apache Iceberg**.

### Executando consulta no Athena

1. Acesse o **Amazon Athena** no console da AWS.
2. No banco de dados `dm_silver`, localize as tabelas como `beer`, `brewery`, `profile`, `review`, `review_flat`.
3. Execute a seguinte query para validar a tabela `beer`:

```sql
SELECT * FROM dm_silver.beer LIMIT 10;
```

Se tudo estiver correto, você verá:

* Registros com os dados da tabela `beer`;
* Campos `created_at` e `updated_at` já preenchidos;
* E o campo `partitioned_at`, correspondente ao dia de processamento (ex: `20250721`).

### Observações sobre o Iceberg no Athena

* A integração com Iceberg permite consultas com **partição automática**, ou seja, o Athena lê apenas os dados da partição correspondente.
* Se desejar consultar uma partição específica:

```sql
SELECT * 
  FROM dm_silver.beer
 WHERE partitioned_at = '20250721';
```

> Essa é a validação final do pipeline Bronze → Silver.

---

[Voltar para a página inicial](../README.md#documentação) | [Próximo: Geração de Dados Analíticos (Gold)](analytics.md)