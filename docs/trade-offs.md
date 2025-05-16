# Decisões de Design e Trade-offs

Nesta seção, explico as principais decisões de design e os trade-offs avaliados ao longo do projeto. As escolhas foram baseadas nos critérios de **baixo custo**, **simplicidade operacional**, **escalabilidade** e **alinhamento com os requisitos do problema**.

## 1. CloudFormation vs Terraform
### Decisão: **CloudFormation**
#### Por que optei por CloudFormation?
- **Integração nativa com AWS**: CloudFormation é profundamente integrado com os serviços AWS, o que simplifica o gerenciamento de recursos.
- **Custo zero**: Não há custos adicionais para usar CloudFormation, enquanto ferramentas externas como Terraform podem exigir custos indiretos (ex.: execução em pipelines).
- **Simplicidade**: Para o escopo deste projeto, CloudFormation é suficiente para gerenciar os recursos necessários.

#### Alternativa Considerada: **Terraform**
- **Prós do Terraform**:
  - Multi-cloud: Suporte a várias nuvens, o que pode ser útil em projetos multi-cloud futuros.
  - Modularidade flexível: Mais opções para criar módulos reutilizáveis.
- **Contras do Terraform**:
  - Requer instalação e configuração adicionais fora da AWS.
  - Menos integração nativa com a AWS, o que pode aumentar a complexidade operacional.

---

## 2. Aurora Serverless v2 vs RDS provisionado
### Decisão: **Aurora Serverless v2**
#### Por que optei por Aurora Serverless?
- **Escalabilidade automática e sob demanda**: O banco cresce e reduz a capacidade de forma transparente.
- **Modelo de cobrança por segundo**: Ideal para cenários com carga variável e jobs event-driven.
- **Dispensa configuração de tamanho de instância**: Elimina a preocupação com right-sizing manual.

#### Alternativa Considerada: **RDS provisionado**
- **Prós do RDS provisionado**:
  - Estável para cargas previsíveis.
  - Disponível em mais regiões e opções de engine.
- **Contras do RDS provisionado**:
  - Custo fixo, mesmo em períodos ociosos.
  - Requer planejamento de escalabilidade e manutenção de instâncias.

---

## 3. CLI Catalog Sync vs Glue Crawlers
### Decisão: **CLI Catalog Sync**
#### Por que optei por não usar Glue Crawlers?
- **Automação total via CLI**: o projeto foi construído com foco em reprodutibilidade e infraestrutura como código. Crawlers exigiriam agendamento e execução fora do controle direto do CLI, quebrando essa filosofia.
- **Schemas versionados e determinísticos**: como o schema do banco é criado por migrations e mantido sob versionamento, não há vantagem em inferência automática. Toda criação de tabela no Glue Catalog é derivada diretamente do schema relacional.
- **Mapeamento semântico estruturado**: o CLI já mapeia os schemas do Aurora (`core`, `views`, `marts`) para os Glue Databases (`dm_bronze`, `dm_silver`, `dm_gold`), o que garante consistência sem necessidade de descoberta dinâmica.

#### Alternativa Considerada: **Glue Crawlers**
- **Prós dos Crawlers**:
  - Descoberta automática de tabelas e colunas em fontes como S3 ou JDBC.
  - Útil quando o schema não é conhecido ou muda frequentemente.
- **Contras dos Crawlers**:
  - Pouco controle sobre execução e versionamento.
  - Depende de permissões específicas e pode causar conflitos em ambientes com múltiplos times.

---

## 4. Glue Jobs vs EMR Serverless
### Decisão: **Glue Jobs**
#### Por que optei por Glue Jobs?
- **Custo menor e modelo pay-per-use simplificado**
- **Integração direta com S3, Glue Catalog e Step Functions**
- **Provisionamento mais rápido e menos complexo**

#### Alternativa Considerada: **EMR Serverless**
- **Prós do EMR Serverless**:
  - Mais flexível para jobs Spark complexos
  - Suporte a múltiplas engines (Spark, Hive, Presto)
- **Contras do EMR Serverless**:
  - Requer configuração mais detalhada (application, release label)
  - Modelo de cobrança menos previsível para workloads leves

---

## 5. Iceberg vs Delta Lake
### Decisão: **Apache Iceberg**
#### Por que optei por Iceberg?
- **Integração nativa com AWS Glue e Athena**: o suporte da AWS a Iceberg é oficial e estável, com suporte a catálogo via Glue e leitura direta pelo Athena.
- **Separação de metadados e dados**: Iceberg trata metadados de forma transacional e com versionamento leve, o que melhora a confiabilidade sem exigir engine dedicada.
- **Independência de engine**: Iceberg não depende do Spark como Delta, e pode ser lido por diversas engines sem lock-in.

#### Alternativa Considerada: **Delta Lake**
- **Prós do Delta**:
  - Maturidade no ecossistema Spark/Databricks.
  - Bom para pipelines que já dependem fortemente de Delta.
- **Contras do Delta**:
  - Dependência maior de Spark e Databricks.
  - Integração com AWS é mais limitada e exige workarounds.

---

## 6. Kinesis vs Apache Kafka
### Decisão: **Kinesis**
#### Por que optei por Kinesis?
- **Serverless**: o Kinesis é completamente gerenciado, eliminando a necessidade de gerenciar clusters ou infraestrutura.
- **Custo baseado em consumo**: ideal para startups e pequenos negócios que precisam minimizar custos iniciais.
- **Simplicidade de integração**: integração direta com outros serviços AWS, como Firehose e Lambda.

#### Alternativa Considerada: **Apache Kafka**
- **Prós do Kafka**:
  - Open-source: maior controle sobre a configuração e possibilidade de evitar lock-in.
  - Escalabilidade: pode ser mais adequado para altíssimas cargas de dados em sistemas muito grandes.
- **Contras do Kafka**:
  - Complexidade operacional: requer a gestão de clusters e configurações avançadas.
  - Custos iniciais: mesmo no modo gerenciado (ex.: Confluent Cloud), o custo pode ser significativamente maior do que o Kinesis.

---

## 7. Step Functions vs Apache Airflow
### Decisão: **Step Functions**
#### Por que optei por Step Functions?
- **Serverless**: reduz a carga operacional e elimina a necessidade de gerenciar servidores.
- **Integração com AWS**: simples de orquestrar serviços como Lambda, Glue e S3.
- **Custo por execução**: ideal para pipelines event-driven.

#### Alternativa Considerada: **Apache Airflow**
- **Prós do Airflow**:
  - Flexibilidade no design de DAGs complexos.
  - Open-source: pode ser executado em qualquer ambiente.
- **Contras do Airflow**:
  - Requer gerenciamento de infraestrutura (mesmo na versão gerenciada, como MWAA).
  - Mais complexo para integrar com serviços nativos da AWS.

---

## 8. S3 + Athena vs Redshift
### Decisão: **S3 + Athena**
#### Por que optei por S3 + Athena?
- **Custo zero na infraestrutura de consulta**: você paga apenas pelas queries executadas, sem custos fixos.
- **Escalabilidade natural**: o S3 é altamente escalável e pode gerenciar volumes massivos de dados.
- **Simplicidade**: não requer a configuração de clusters, tornando-o ideal para startups com equipes pequenas.

#### Alternativa Considerada: **Redshift**
- **Prós do Redshift**:
  - Alta performance em consultas analíticas complexas.
  - Mais adequado para workloads de BI com consultas frequentes e pesadas.
- **Contras do Redshift**:
  - Custos fixos mesmo com uso pequeno.
  - Necessidade de configuração e tuning de clusters, aumentando a complexidade.

---

## 9. Managed Grafana vs QuickSight
### Decisão: **Managed Grafana**
#### Por que optei por Managed Grafana?
- **Flexibilidade**: suporte a diversas fontes de dados, incluindo Athena e CloudWatch.
- **Observabilidade**: além de dashboards, permite visualização de métricas e logs.
- **CloudFormation**: totalmente integrado ao CloudFormation, facilitando o provisionamento.

#### Alternativa Considerada: **QuickSight**
- **Prós do QuickSight**:
  - Simplicidade: focado em visualização de dados, com boas opções para business intelligence.
  - Baixo custo: paga-se por usuário, o que pode ser vantajoso em alguns cenários.
- **Contras do QuickSight**:
  - Integração limitada: não suporta todas as fontes de dados exigidas neste projeto.
  - Não é 100% compatível com CloudFormation, dificultando a automação.

---

## Conclusão
Todas as decisões foram tomadas com base no escopo do projeto, priorizando **custo**, **simplicidade**, **governança automatizada** e **resiliência operacional**.
