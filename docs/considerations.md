## Considerações Finais

O projeto Data Master foi concebido com o objetivo de demonstrar domínio técnico e arquitetural em uma jornada completa de dados, utilizando recursos serverless e de baixo custo da AWS. Ao longo de sua construção, foram enfrentados desafios reais relacionados à ingestão, processamento, governança, observabilidade e análise de dados — todos endereçados com foco em simplicidade operacional, performance e reprodutibilidade.

A arquitetura proposta adota o padrão **Medallion (raw, bronze, silver, gold)** com controle centralizado via DynamoDB, tratamento incremental dos dados com suporte a eventos `insert`, `update` e `delete`, e uso estratégico de ferramentas como Glue, Lambda, ECS, EMR e Athena. A etapa de observabilidade foi integrada por meio do Grafana, com dashboards que cobrem desde indicadores analíticos até custos e saúde do ambiente.

Uma das decisões mais ousadas do projeto foi a implementação do processamento da camada **bronze** utilizando **Go puro**, em vez de Spark ou outras ferramentas consolidadas para workloads analíticos. Essa escolha exigiu o desenvolvimento de soluções personalizadas para leitura, transformação e escrita de arquivos **Parquet**. Embora desafiasse o caminho mais comum, essa abordagem se mostrou eficiente, escalável e perfeitamente viável para workloads com alto volume de dados — reforçando o uso do Go como uma alternativa viável e eficiente para pipelines analíticos de alta performance.

Cada etapa do projeto foi registrada e estruturada de forma a permitir reprodutibilidade total por um engenheiro de dados, com uso extensivo de CloudFormation, um CLI interativo desenvolvido em Go e documentação modular.

O projeto demonstra, na prática, como arquiteturas analíticas modernas podem ser implementadas de forma pragmática, com ênfase em modularidade, automação e controle de custos — respeitando os limites operacionais de ambientes corporativos.

---

[Voltar para a página inicial](../README.md#documentação)
