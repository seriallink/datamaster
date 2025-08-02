# Data Lake - Arquitetura Medallion

Este projeto implementa um Data Lake baseado na **Medallion Architecture**, com quatro camadas principais: **raw**, **bronze**, **silver** e **gold**. Cada camada foi projetada com objetivos distintos e decisões técnicas específicas, equilibrando performance, custo, governança e flexibilidade.

---

## Visão Geral

| Camada | Finalidade Principal                                | Formato               | Ferramentas Principais  |
| ------ | --------------------------------------------------- |-----------------------|-------------------------|
| Raw    | Armazenar o dado original, sem tratamento           | `.json.gz`, `.csv.gz` | Kinesis + Firehose + Go |
| Bronze | Padronizar e transformar para formato colunar       | Hive (Parquet)        | Go (parquet-go)         |
| Silver | Normalizar, aplicar lógica de negócio e enrichments | Iceberg (Parquet)     | Spark (PySpark)         |
| Gold   | Métricas analíticas e agregações finais             | Iceberg (Parquet)     | Spark (PySpark)         |

---

## Camada Raw

### Conceito

A camada **raw** é a base do Data Lake. Os dados são armazenados **exatamente como foram recebidos**, com o mínimo de transformação. Esta camada é essencial para **rastreabilidade**, **auditoria** e **reprocessamento seguro**.

### Formatos Suportados

* `.json.gz`
* `.csv.gz`

### Ferramentas 

* **Kinesis Data Streams**: recebe eventos de CDC do DMS.
* **Kinesis Firehose**: responsável pela entrega no S3.
* **AWS Lambda**: usada como função transformadora no Firehose, para extrair metadados e definir dinamicamente partições.
* **Firehose Writer**: grava diretamente os dados .gz no bucket dm-stage, sob a estrutura raw por tabela.

### Justificativas Técnicas

* **Compressão `.gz`**: simples, portátil e eficiente.
* **Manutenção do formato original**: evita acoplamento precoce a schemas.
* **Sem parsing ou validação**: ingestão resiliente, desacoplada do processamento.

---

## Camada Bronze

### Conceito

A camada **bronze** realiza a primeira transformação dos dados: converte arquivos brutos em um formato colunar eficiente com **schema definido**, porém ainda próximo do dado original.

### Formato: Parquet com Partições Hive

* Esquema estático por tabela.
* Dados armazenados em Parquet, organizados por `year/month/day`.

### Ferramenta: Go (alta performance)

* Uso de bibliotecas `parquet-go` e processamento concorrente.
* Ideal para execução leve (Lambda, ECS), sem necessidade de Spark.
* Controle de ingestão e status via DynamoDB (`ProcessingControl`).

### Justificativas Técnicas

* **Parquet**: formato colunar, compressão embutida, leitura seletiva.
* **Go em vez de Spark**:
    * Evita sobrecarga de infraestrutura.
    * Oferece controle refinado e ótimo desempenho.
    * Apresenta cold start muito mais rápido (especialmente em Lambda).
    * Ideal para workloads event-driven e streaming-friendly.
* **Sem Iceberg nesta camada**:
    * Iceberg é desnecessário para dados brutos e descartáveis.
    * Não há necessidade de `time travel` ou `merge`.
    * Mantém a camada leve e rápida.

---

## Camada Silver

### Conceito

A camada **silver** representa os dados **validados, normalizados e enriquecidos**. Aqui os registros recebem tratamento de *lógica de negócio* e são preparados para consumo analítico.

### Formato: Iceberg

* Permite operações `INSERT`, `UPDATE`, `DELETE`.
* Suporte a `schema evolution`, `time travel` e propriedades `ACID`.
* Particionado por `partitioned_at` (formato `yyyy-MM-dd`), facilitando o controle de cargas incrementais e otimização de queries.

### Ferramenta: Spark (PySpark)

* Joins e enrichments entre dimensões (ex: `review` + `beer` + `profile` + `brewery`).
* Suporte nativo a Iceberg.

### Estratégias de Small Files

* A opção `optimize_small_files` foi habilitada explicitamente no Glue Catalog, permitindo que o processamento de dados do Iceberg otimiza a escrita consolidando arquivos menores.
* Outras técnicas adicionais podem ser empregadas futuramente, como jobs periódicos de compactação (RewriteDataFiles).

### Enriquecimentos aplicados

* `created_at`, `updated_at` derivados do controle de processamento.
* Exclusões lógicas via `deleted_at` para suporte a soft delete.
* Tratamento incremental por tabela.

### Justificativas Técnicas

* Iceberg oferece **governança completa** nesta camada.
* Spark facilita processamento paralelo e lógicas de transformação complexas.
* Time travel é valioso para **debug, auditoria e recuperação**.

---

## Camada Gold

### Conceito

A camada **gold** é voltada para **consumo direto por dashboards, relatórios e BI**. Aqui os dados são agregados, refinados e modelados como **fatos e dimensões analíticas**.

### Formato: Iceberg

* Mantém todos os benefícios da camada silver (`merge`, `evolution`, `ACID`).
* Otimizado para leitura via Athena, Grafana, etc.
* Todas as tabelas são particionadas por `review_year` e `review_month`, permitindo **filtros dinâmicos em painéis analíticos**.

### Ferramentas

* Spark para geração das métricas.
* Athena e Grafana para consumo e visualização.

### Estratégia de Atualização

* Processamento incremental por período.
* Geração baseada em dados atualizados na camada silver.

### Justificativas Técnicas

* Iceberg permite **recalcular métricas mantendo histórico**.
* Layout de tabelas otimizado para **queries analíticas de baixo custo**.
* Facilita **versionamento de snapshots** e **rastreabilidade de alterações**.

---

## Considerações Finais

* Cada camada foi projetada com **foco na função que exerce**, evitando overengineering.
* A separação clara entre raw → bronze → silver → gold permite modularidade, reprocessamento isolado e governança granular.
* A decisão de **usar Go na bronze** e **Spark nas camadas analíticas** equilibra **performance, custo e robustez**.
* Iceberg foi adotado apenas nas camadas onde seus benefícios realmente se justificam.

---

[Voltar para a página inicial](../README.md#documentação) | [Próximo: Visão Geral da Arquitetura](architecture.md)