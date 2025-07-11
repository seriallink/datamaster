# Data Master Case – Serverless Data Lake on AWS

![go-data-master.png](assets/go-data-master.png)

> Projeto de arquitetura serverless e orientada a eventos para data lakes na AWS — escalável, de baixo custo e pronta para produção. Desenvolvido para o programa Data Master da F1rst/Santander.

---

## Objetivo

Este projeto tem como objetivo construir uma solução completa de engenharia de dados baseada em nuvem, utilizando arquitetura serverless, modular e orientada a eventos. A proposta simula um domínio real de dados e entrega uma estrutura preparada para:

- Ingestão de dados nos modos `streaming` e `batch`
- Processamento em múltiplas camadas: `raw`, `bronze`, `silver`, `gold`
- Organização e catalogação automatizada com `Glue` e `Iceberg`
- Transformações otimizadas com `Lambda`, `ECS` e `EMR Serverless`
- Orquestração de pipelines com `Step Functions` e `EventBridge`
- Visualização por meio de dashboards analíticos e operacionais com `Grafana`

---

## Documentação

> A documentação completa está disponível na pasta [`docs/`](./docs)

- [01. Apresentação do Case](docs/overview.md)
- [02. Modelo de Dados](docs/mer.md)
- [03. Visão Geral da Arquitetura](./docs/architecture.md)
- [04. Trade-offs e Decisões de Arquitetura](./docs/tradeoffs.md)
- [05. Instalação do Data Master CLI](./docs/installation.md)
- [06. Utilização do Data Master CLI](./docs/how-to-use.md)
- [07. Provisionamento do Ambiente](./docs/provisioning.md)
- [08. Observabilidade](./docs/observability.md)
- [09. Melhorias](./docs/roadmap.md)
