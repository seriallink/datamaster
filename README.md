# Data Master Case – Serverless Data Lake on AWS

> Projeto de arquitetura serverless e orientada a eventos para data lakes na AWS — escalável, de baixo custo e pronta para produção. Desenvolvido para a certificação Expert do programa Data Master do Santander.

---

## 🧠 Objetivo

Este projeto foi desenvolvido como parte do programa de certificação interna **Data Master** do Santander, com foco em demonstrar domínio técnico em arquitetura de dados moderna, uso de serviços gerenciados na AWS, governança, e boas práticas de engenharia de dados.

## 📐 Visão Geral da Arquitetura

Arquitetura baseada em serviços **serverless**, orientada a eventos e com foco em **custo reduzido**. A stack inclui:

- Aurora PostgreSQL + DMS para captura de mudanças (CDC)
- Kinesis + Firehose para ingestão streaming
- S3 para armazenamento em formatos otimizados (Parquet)
- Glue Data Catalog + Iceberg para estruturação
- Lambda + EMR para processamento em camadas
- Athena + QuickSight para consumo e análise
- Lake Formation para governança
- Step Functions para orquestração

> Diagrama completo disponível em [`diagrams/`](diagrams/).

## 🚀 Guia Rápido de Execução

> Instruções completas no [guia de instalação e execução](docs/setup-guide.md)

1. Clone o repositório
2. Provisione a infraestrutura com CloudFormation
3. Inicie os serviços de ingestão (batch/streaming)
4. Valide os dados e acompanhe o pipeline
5. Acesse o Athena e o QuickSight para análise

```bash
git clone https://github.com/seu-usuario/data-master-case.git
cd data-master-case
