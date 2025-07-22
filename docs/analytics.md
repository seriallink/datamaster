# Geração de Dados Analíticos (Gold)

A camada **gold** representa o estágio final do pipeline de dados, onde as informações da silver são transformadas em **métricas analíticas, rankings e agregações otimizadas para dashboards, relatórios e BI**. As tabelas são modeladas com foco em consumo direto por ferramentas como **Athena**, **Grafana**, **Power BI** e **Metabase**, utilizando o formato transacional **Iceberg**.

---

## 1. Visão Geral da Arquitetura

O pipeline é totalmente automatizado e executado diariamente via **Step Function** (`dm-analytics-dispatcher`), acionada por um evento agendado do **EventBridge** (`cron` diário às 7h UTC). O fluxo é composto por múltiplas execuções independentes por tabela analítica:

1. **Processamento Paralelo por Tabela**
   Cada tabela `top_beers_by_rating`, `top_breweries_by_rating`, `state_by_review_volume`, `top_styles_by_popularity` e `top_drinkers` é processada individualmente por jobs Spark executados no **EMR Serverless**. A transformação ocorre via `spark-submit` com argumentos dinâmicos.

2. **Fan-out Controlado**
   A Step Function executa os jobs de forma isolada e paralela (por `MaxConcurrency`, podendo ser ajustado conforme necessidade), com controle transacional e logs centralizados no S3 (`dm-observer-bucket`).

As tabelas resultantes são gravadas em formato **Iceberg**, já particionadas por `review_year` e `review_month`, prontas para consumo imediato no Athena ou Grafana.

---

## 2. Frequência e Estratégia de Processamento

O pipeline da camada gold é executado em **modo batch diário**, logo após o término da camada silver. A estratégia de processamento leva em consideração:

* **Dependência da silver**: as métricas calculadas na gold dependem das tabelas normalizadas, enriquecidas e versionadas da camada anterior.
* **Granularidade mensal**: como os dados da silver são particionados por ano/mês, as tabelas gold também herdam essa granularidade para facilitar filtros temporais em dashboards.
* **Processamento incremental**: apenas os períodos modificados na silver são recalculados na gold.
* **Registro em DynamoDB**: ao final do processamento da silver, são inseridos itens na tabela `dm-processing-control` com `schema_name = dm_gold` e `status = pending`, indicando que determinado período está pronto para ser gerado na gold.

Essa abordagem permite **rastreabilidade, reprocessamento e observabilidade granular**, mantendo o pipeline escalável e auditável.

---

## 3. Como verificar períodos pendentes para geração (silver → gold)

Assim como na silver, os períodos prontos para geração da gold são controlados pela tabela `dm-processing-control` no **DynamoDB**.

### Passos para consulta manual:

1. Acesse o serviço **DynamoDB** no AWS Console.

2. Vá até a tabela **`dm-processing-control`**.

3. Clique em **Explore table items**.

4. Aplique o filtro:

    * **Attribute name**: `schema_name`
    * **Condition**: `Equal to`
    * **Type**: `String`
    * **Value**: `dm_gold`

5. Clique em **Run**.

Isso exibirá os períodos com `status = pending`, indicando que estão aguardando processamento da camada gold.

### Observações

* O campo `created_at` indica quando a silver concluiu o período.
* Os campos `table_name`, `review_year`, `review_month` e `status` ajudam a identificar exatamente o que será processado.
* O campo `compute_target` segue como `"emr"`.

---

## 4. Processamento Manual da Camada Gold

Durante desenvolvimento ou reprocessamento, é possível disparar a execução da gold manualmente com o CLI:

### Comando

```
>>> process --layer gold
```

Esse comando dispara a execução da **Step Function `dm-analytics-dispatcher`**, com base nos registros pendentes da camada `dm_gold` no DynamoDB.

### Detalhes

* O comando **não é síncrono**. Ele apenas inicia a execução.
* A Step Function invoca um job Spark para cada tabela gold.

### Acompanhamento

1. Acesse o **Step Functions** no console da AWS.
2. Clique em **`dm-analytics-dispatcher`**.
3. Vá em **Executions** e localize a execução atual.
4. Clique na execução para acompanhar em tempo real.

---

## 5. Acompanhando o processamento no EMR Serverless

Cada tabela da camada gold é processada por um job Spark individual.

### Como verificar:

1. Acesse o serviço **EMR** no AWS Console.
2. Clique em **EMR Serverless** > **Manage applications**.
3. Selecione a aplicação **`dm-analytics-app`**.
4. Em **Batch job runs**, veja as execuções recentes.
5. Clique no **Job Run ID** desejado para detalhes:

    * Tempo de execução
    * Argumentos (ex: `--layer gold --table top_beers_by_rating`)
    * Logs
    * CPU/memória utilizada

---

## 6. Verificando os dados na camada Gold

Os dados processados são salvos em formato **Iceberg**, no bucket `dm-datalake`, particionados por `review_year` e `review_month`.

### Caminho do bucket

```
s3://dm-datalake-<account_id>/gold/<table_name>/
```

Exemplo para `top_beers_by_rating`:

```
gold/top_beers_by_rating/
├── metadata/
└── data/
    ├── review_year=YYYY/
    │   └── review_month=MM/
    │       └── *.parquet
```

Cada tabela contém:

* Diretório `metadata/` com arquivos internos do Iceberg.
* Diretório `data/` com os arquivos Parquet particionados.

---

## 7. Verificação via Athena

Após o processamento da camada gold, os dados ficam imediatamente disponíveis para consulta no **Athena**, via Glue Catalog.

### Passos:

1. Acesse o **Amazon Athena**.
2. Use o banco de dados **`dm_gold`**.
3. Execute uma query de validação:

```sql
SELECT *
  FROM dm_gold.top_beers_by_rating
 WHERE review_year = 2010
   AND review_month = 12
 LIMIT 10;
```

### Observações

* Todas as tabelas seguem o padrão de particionamento `review_year` e `review_month`.
* O formato Iceberg garante performance e controle de schema.
* As métricas e rankings são recalculadas sempre que a silver for atualizada.

---

## 8. Tabelas Disponíveis

A camada gold é composta por tabelas analíticas prontas para dashboards:

| Tabela                     | Descrição                               |
| -------------------------- | --------------------------------------- |
| `top_beers_by_rating`      | Cervejas com melhores avaliações        |
| `top_breweries_by_rating`  | Cervejarias mais bem avaliadas          |
| `state_by_review_volume`   | Volume de reviews por estado            |
| `top_styles_by_popularity` | Estilos mais populares                  |
| `top_drinkers`             | Usuários com maior número de avaliações |

Todas são otimizadas para leitura e exploráveis via Athena ou conectores como Grafana.

---

[Voltar para a página inicial](../README.md#documentação) | [Próximo: DataViz e Observabilidade](observability.md)