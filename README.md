# Data Master Case – Serverless Data Lake na AWS

![img.png](assets/go-data-master.png)

> Projeto de arquitetura serverless e orientada a eventos para data lakes na AWS — escalável, de baixo custo e pronta para produção. Desenvolvido para o programa Data Master da F1rst/Santander.

---

## Objetivo

Este projeto tem como objetivo construir uma solução completa de engenharia de dados baseada em nuvem, utilizando arquitetura serverless, modular e orientada a eventos. A proposta simula um domínio real de dados e entrega uma estrutura preparada para:

- Ingestão de dados nos modos streaming e batch
- Processamento em múltiplas camadas: raw, bronze, silver, gold
- Organização e catalogação automatizada com Glue e Iceberg
- Transformações otimizadas com Lambda, ECS e EMR Serverless
- Orquestração de pipelines com Step Functions e EventBridge
- Visualização por meio de dashboards analíticos e operacionais com Grafana

---

## Documentação

> A documentação completa está disponível na pasta [`docs/`](./docs)

- [1. Visão Geral da Arquitetura](./docs/architecture.md)
- [2. Instalação do Projeto](./docs/installation.md)
- [3. Uso do CLI](./docs/cli.md)
- [4. Provisionamento do Ambiente](./docs/provisioning.md)
- [5. Camadas de Dados](./docs/layers.md)
- [6. Observabilidade](./docs/observability.md)
- [7. Governança e Custos](./docs/governance.md)
- [8. FAQ](./docs/faq.md)
