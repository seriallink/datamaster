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
- Governança de dados com `Lake Formation`, controle de acesso e mascaramento automático de PII com `Comprehend`
- Visualização por meio de dashboards analíticos e operacionais com `Grafana`
- Execução `end-to-end` totalmente automatizada, da ingestão até os dashboards

---

## Documentação

> A documentação completa está disponível na pasta [`docs/`](./docs)

- [01. Apresentação do Case](docs/overview.md)
- [02. Modelo de Dados](docs/mer.md)
- [03. Camadas do Data Lake (Medallion Architecture)](./docs/layers.md)
- [04. Visão Geral da Arquitetura](./docs/architecture.md)
- [05. Componentização da Arquitetura por Stacks](./docs/stacks.md)
- [06. Trade-offs e Decisões de Arquitetura](docs/trade-offs.md)
- [07. Pré-Requisitos](docs/pre-requirements.md)
- [08. Instalação do Data Master CLI](./docs/installation.md)
- [09. Utilização do Data Master CLI](./docs/how-to-use.md)
- [10. Provisionamento do Ambiente](./docs/provisioning.md)
- [11. Ingestão de Dados (Bronze)](./docs/ingestion.md)
- [12. Processamento de Dados (Silver)](./docs/processing.md)
- [13. Geração de Dados Analíticos (Gold)](./docs/analytics.md)
- [14. DataViz e Observabilidade](./docs/observability.md)
- [15. Governança e Segurança de Dados](docs/governance.md)
- [16. Roadmap Técnico e Melhorias Futuras](./docs/roadmap.md)
- [17. Considerações Finais](docs/considerations.md)
