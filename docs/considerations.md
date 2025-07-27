## Considerações Finais

O projeto Data Master foi concebido com o objetivo de demonstrar domínio técnico e arquitetural em uma jornada completa de dados, utilizando uma **arquitetura event-driven** baseada em recursos **serverless** e de **baixo custo** da AWS. Ao longo de sua construção, foram enfrentados desafios reais relacionados à ingestão, processamento, governança, observabilidade e análise de dados — todos endereçados com foco em simplicidade operacional, performance e reprodutibilidade.

A arquitetura proposta adota o padrão **Medallion** (raw, bronze, silver, gold), com controle centralizado via **DynamoDB**, tratamento incremental dos dados com suporte a eventos, e ingestão híbrida — combinando **streaming em tempo quase real** (via Kinesis e Firehose) com **cargas em lote** para dados batch. O projeto também faz uso estratégico de ferramentas como Glue, Lambda, ECS, EMR e Athena. A governança de acesso é gerenciada por meio do **Lake Formation**, permitindo controle granular de permissões por camada e integração com políticas centralizadas. A etapa de observabilidade foi integrada por meio do Grafana, com dashboards que cobrem desde indicadores analíticos até custos e saúde do ambiente.

Uma das iniciativas mais exploratórias do projeto foi a implementação do processamento da camada **bronze** utilizando **Go puro**, em vez de Spark ou outras ferramentas consolidadas para workloads analíticos. Embora fugisse do caminho mais comum, essa abordagem se mostrou eficiente, escalável e extremamente vantajosa tanto em cenários com alto volume de dados quanto em cargas menores e frequentes — onde o uso de **Lambda**, aliado à ausência de cold start significativo e à eliminação da necessidade de clusters, oferece uma otimização expressiva de custo e tempo. Isso reforça o uso do **Go** como uma alternativa concreta e eficiente para pipelines analíticos modernos, especialmente em arquiteturas orientadas a eventos e de alta frequência.

Cada etapa do projeto foi registrada e estruturada de forma a permitir reprodutibilidade total por um engenheiro de dados, com uso extensivo de CloudFormation, um CLI interativo e documentação modular.

O projeto demonstra, na prática, como arquiteturas analíticas modernas podem ser implementadas de forma pragmática, com ênfase em modularidade, automação e controle de custos — respeitando os limites operacionais de ambientes corporativos. Ao mesmo tempo, explora caminhos menos tradicionais, mostrando que soluções simples e performáticas podem surgir fora dos grandes frameworks, especialmente quando aplicadas com propósito e domínio técnico.

---

[Voltar para a página inicial](../README.md#documentação)
