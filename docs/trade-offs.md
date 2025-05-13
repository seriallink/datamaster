# Decisões de Design e Trade-offs

Nesta seção, explicamos as principais decisões de design e os trade-offs avaliados ao longo do projeto. As escolhas foram baseadas nos critérios de **baixo custo**, **simplicidade operacional**, **escalabilidade** e **alinhamento com os requisitos do problema**.

## 1. CloudFormation vs Terraform
### Decisão: **CloudFormation**
#### Por que escolhemos CloudFormation?
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

## 2. Kinesis vs Apache Kafka
### Decisão: **Kinesis**
#### Por que escolhemos Kinesis?
- **Serverless**: O Kinesis é completamente gerenciado, eliminando a necessidade de gerenciar clusters ou infraestrutura.
- **Custo baseado em consumo**: Ideal para startups e pequenos negócios que precisam minimizar custos iniciais.
- **Simplicidade de integração**: Integração direta com outros serviços AWS, como Firehose e Lambda.

#### Alternativa Considerada: **Apache Kafka**
- **Prós do Kafka**:
    - Open-source: Maior controle sobre a configuração e possibilidade de evitar lock-in.
    - Escalabilidade: Pode ser mais adequado para altíssimas cargas de dados em sistemas muito grandes.
- **Contras do Kafka**:
    - Complexidade operacional: Requer a gestão de clusters e configurações avançadas.
    - Custos iniciais: Mesmo no modo gerenciado (ex.: Confluent Cloud), o custo pode ser significativamente maior do que o Kinesis.

---

## 3. Managed Grafana vs QuickSight
### Decisão: **Managed Grafana**
#### Por que escolhemos Managed Grafana?
- **Flexibilidade**: Suporte a diversas fontes de dados, incluindo Athena e CloudWatch.
- **Observabilidade**: Além de dashboards, permite visualização de métricas e logs.
- **CloudFormation**: Totalmente integrado ao CloudFormation, facilitando o provisionamento.

#### Alternativa Considerada: **QuickSight**
- **Prós do QuickSight**:
    - Simplicidade: Focado em visualização de dados, com boas opções para business intelligence.
    - Baixo custo: Paga-se por usuário, o que pode ser vantajoso em alguns cenários.
- **Contras do QuickSight**:
    - Integração limitada: Não suporta todas as fontes de dados exigidas neste projeto.
    - Não é 100% compatível com CloudFormation, dificultando a automação.

---

## 4. S3 + Athena vs Redshift
### Decisão: **S3 + Athena**
#### Por que escolhemos S3 + Athena?
- **Custo zero na infraestrutura de consulta**: Você paga apenas pelas queries executadas, sem custos fixos.
- **Escalabilidade natural**: O S3 é altamente escalável e pode gerenciar volumes massivos de dados.
- **Simplicidade**: Não requer a configuração de clusters, tornando-o ideal para startups com equipes pequenas.

#### Alternativa Considerada: **Redshift**
- **Prós do Redshift**:
    - Alta performance em consultas analíticas complexas.
    - Mais adequado para workloads de BI com consultas frequentes e pesadas.
- **Contras do Redshift**:
    - Custos fixos mesmo com uso pequeno.
    - Necessidade de configuração e tuning de clusters, aumentando a complexidade.

---

## 5. Step Functions vs Apache Airflow
### Decisão: **Step Functions**
#### Por que escolhemos Step Functions?
- **Serverless**: Reduz a carga operacional e elimina a necessidade de gerenciar servidores.
- **Integração com AWS**: Simples de orquestrar serviços como Lambda, Glue e S3.
- **Custo por execução**: Ideal para pipelines event-driven.

#### Alternativa Considerada: **Apache Airflow**
- **Prós do Airflow**:
    - Flexibilidade no design de DAGs complexos.
    - Open-source: Pode ser executado em qualquer ambiente.
- **Contras do Airflow**:
    - Requer gerenciamento de infraestrutura (mesmo na versão gerenciada, como MWAA).
    - Mais complexo para integrar com serviços nativos da AWS.

---

## Conclusão
Todas as decisões foram tomadas com base no escopo do projeto, priorizando **custo**, **simplicidade** e **escalabilidade**. Ferramentas como Terraform, Kafka ou Redshift podem ser consideradas em cenários futuros, caso as necessidades do projeto mudem.