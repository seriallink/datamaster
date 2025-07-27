# Roadmap Técnico e Melhorias Futuras

O projeto foi desenvolvido com o objetivo de atender aos critérios do programa de certificação do Data Master. Durante sua construção, porém, diversos pontos foram identificados como oportunidades de evolução arquitetural, operacional e analítica. Algumas dessas melhorias não foram implementadas por restrições técnicas, de escopo ou prioridade, mas representam possíveis caminhos estratégicos para ganho de eficiência, escalabilidade e governança no médio e longo prazo.

---

## Evolução Técnica e de Performance

### Adoção futura do Apache Iceberg com Go (`iceberg-go`)

Atualmente, as camadas **silver** e **gold** do projeto são processadas integralmente em **PySpark**, com persistência dos dados no formato **Iceberg**. Essa abordagem oferece **robustez e flexibilidade**, mas depende da infraestrutura do **EMR Serverless**, o que impacta diretamente:

* O **custo por job**, especialmente para transformações simples ou cargas frequentes
* A **performance geral**, devido ao tempo de **cold start** característico do EMR Serverless, que adiciona latência significativa mesmo em execuções rápidas

O projeto [iceberg-go](https://github.com/apache/iceberg-go), mantido pela própria Apache, está em evolução ativa e representa uma **alternativa estratégica** para o futuro. Com sua maturação, será possível:

* **Eliminar a dependência do PySpark** em jobs mais leves
* **Reduzir custos operacionais**, executando transformações simples diretamente em binários Go via Lambda ou ECS
* **Unificar a base tecnológica**, aproveitando a simplicidade e velocidade de desenvolvimento do Go em outras partes do projeto

> Assim que o `iceberg-go` atingir estabilidade para operações de leitura e escrita com Iceberg em ambientes de produção, será avaliada sua adoção como engine principal em camadas analíticas e no processamento incremental.

---

### Compactação ativa de arquivos Iceberg (RewriteDataFiles)

Para mitigar o problema de `small files`, a opção `optimize_small_files` foi habilitada explicitamente no Glue Catalog, permitindo que o próprio **Iceberg** otimize a escrita com base em heurísticas internas.

Além disso, o pipeline da camada **silver** já acumula todos os arquivos Parquet de um mesmo dia em um único lote de escrita, o que reduz significativamente a fragmentação.

> **Futuro possível:** aplicar técnicas adicionais de **compaction ativa**, como a execução periódica de **jobs de `RewriteDataFiles`**, permitindo a reescrita eficiente dos dados em arquivos maiores e otimizados para leitura em larga escala.

Essa estratégia é especialmente útil em tabelas com **grande volume de updates** ou inserções contínuas, onde a fragmentação pode prejudicar a performance de consultas analíticas.

---

### Políticas de expurgo nas camadas Raw e Bronze

As camadas **raw** e **bronze** acumulam arquivos de dados históricos utilizados no processamento incremental das camadas superiores. No entanto, **não há políticas de expurgo automatizadas**, o que pode gerar crescimento desnecessário no armazenamento ao longo do tempo.

Essa retenção indefinida impacta diretamente o **custo com S3** e dificulta a **gestão do ciclo de vida dos dados**.

> **Futuro possível:** definir e aplicar **regras de expurgo baseadas em data de criação**, utilizando **S3 Lifecycle Rules**. Como exemplo, arquivos `.gz` da camada raw podem ser excluídos após 90 dias, e arquivos Parquet da bronze após 180 dias, considerando que os dados processados já foram consolidados nas camadas analíticas.

---

### Parametrização de capacidade e configurações de infraestrutura

Atualmente, diversas configurações críticas de provisionamento — como **tamanho de instâncias, quantidade de workers, limites de memória e CPU** — estão definidas diretamente nos arquivos de template (CloudFormation), como valores fixos.

Essa abordagem funcionou bem durante a fase de desenvolvimento e estabilização, mas traz limitações importantes:

* Dificuldade para **ajustar recursos conforme a carga** ou ambiente (dev, staging, prod)
* Impossibilidade de **ajustar dinamicamente** os recursos via CLI ou pipeline de deploy

> **Futuro possível:** promover a **parametrização completa das capacidades computacionais** nas stacks CloudFormation, incluindo:

* Parâmetros de provisionamento para Aurora, Kinesis, ECS, EMR, etc.
* Perfis por ambiente, permitindo deploys diferenciados com overrides

Essa melhoria facilita o controle de custos, o ajuste fino de performance e a automação de testes em múltiplos ambientes com requisitos distintos.

---

## Governança e Segurança de Dados

### Governança e detecção de dados sensíveis na camada Raw

A camada **raw** pode conter dados brutos não estruturados ou sem garantias de sanitização, o que impõe riscos de **exposição acidental de informações sensíveis (PII)**. Atualmente, **não há mecanismo automatizado** para identificar ou classificar esses dados.

> **Futuro possível:** integração com o **[Amazon Macie](https://aws.amazon.com/pt/macie/)**, que permitirá:

* Escaneamento automático dos objetos armazenados no S3
* Detecção de **dados pessoais identificáveis (PII)** e outros conteúdos sensíveis
* Geração de **alertas e classificações de sensibilidade** para reforçar a governança e mitigar riscos de conformidade

---

### Validação de qualidade de dados automatizada (Data Quality)

Apesar dos pipelines estarem estruturados, não há verificação sistemática de integridade ou qualidade dos dados ingeridos ou transformados (ex: nulos indevidos, duplicidade, range incorreto).

> **Futuro possível:** adoção de frameworks como [Deequ](https://github.com/awslabs/deequ), ou lógica custom em PySpark para validação automática com alertas via SNS/Grafana.

---

### Catalogação e linhagem de modelos de dados

Atualmente, o Glue Catalog mantém a **estrutura técnica das tabelas**, mas não há uma camada adicional de documentação, versionamento ou rastreabilidade de transformações. Essa limitação dificulta:

* A **visibilidade sobre a origem e o destino dos dados** (data lineage)
* A **colaboração entre times técnicos e de negócio**, especialmente na interpretação de métricas e indicadores
* A **governança de mudanças em schemas**, como adição ou remoção de colunas e impacto em pipelines

> **Futuro possível:** adoção de uma solução integrada de **catalogação e linhagem de dados**, como [DataHub](https://datahubproject.io/) ou [OpenMetadata](https://open-metadata.org/), permitindo:

* **Documentação semântica** das tabelas, com descrições, owners, classificações e relacionamentos
* **Versionamento de modelos de dados**, com histórico de alterações
* **Visualização de linhagem de dados**, desde a origem (raw) até os produtos analíticos (gold), inclusive com nível de coluna
* **Integração com pipelines existentes**, via Spark, dbt ou scripts customizados

---

### Estabelecimento de Data Contracts entre camadas e sistemas

Atualmente, as transformações entre camadas (raw → bronze → silver → gold) seguem uma estrutura padronizada, mas **não há um contrato formal que garanta a estabilidade dos schemas** ou a confiabilidade das entregas de dados entre produtores (ex: pipelines, DMS) e consumidores (ex: dashboards, jobs analíticos).

Essa ausência de contratos pode gerar impactos silenciosos, como:

* Quebra de dashboards ou jobs ao alterar um schema
* Uso de dados inconsistentes por múltiplos consumidores
* Falta de validação automática de alterações em produção

> **Futuro possível:** adoção de **Data Contracts** como mecanismo formal de definição e validação de schemas, incluindo:

* Versão do schema, campos obrigatórios, tipos e constraints
* Acordos de SLA entre camadas ou sistemas consumidores
* Validação automatizada de mudanças via CI/CD, impedindo deploys quebrados

Ferramentas como **OpenMetadata**, uso de **JSON Schema** com validação em pipeline, ou integração com o próprio Glue Catalog podem ser utilizadas como base para esse processo.

---

## Observabilidade e Visualização Operacional

### Visualização de dados operacionais (DynamoDB)

Os dados de controle operacional — armazenados na tabela `dm-processing-control` no DynamoDB — são parte essencial do monitoramento dos pipelines. No entanto, o **plugin oficial para Amazon DynamoDB no Grafana** é exclusivo do **plano Enterprise**.

Essa limitação impede, por ora, a exibição nativa de informações como tentativas com erro, arquivos em processamento e duração de execuções por job.

> **Futuro possível:** replicar periodicamente os dados do DynamoDB para o S3, tornando-os acessíveis via Athena e, consequentemente, integráveis ao Grafana.

---

### Criação de uma API unificada para orquestração e integração

Atualmente, as operações do projeto — como criação de tabelas, deploy de pipelines e execução de processamentos — são realizadas exclusivamente via **CLI**, o que garante automação e rastreabilidade, mas limita a integração com interfaces externas.

> **Futuro possível:** desenvolvimento de uma **API RESTful** que encapsule as funcionalidades hoje disponíveis na CLI, permitindo:

* **Integração com interfaces web** e portais internos para acionamento de pipelines e visualização de status
* **Conexão com sistemas externos**, como plataformas de orquestração, produtos internos ou sistemas de alerta
* **Governança centralizada** e autenticação padronizada, abrindo caminho para um painel operacional visual acessível a usuários não técnicos

---

## Evolução Estratégica e Comunidade

### Suporte a múltiplas tecnologias e engines

As decisões técnicas adotadas ao longo do projeto — como uso de **Kinesis**, **PostgreSQL (Aurora)**, **DynamoDB**, **EMR Serverless**, etc — foram fundamentadas em trade-offs detalhados na documentação. No entanto, a arquitetura foi construída com foco em **modularidade e desacoplamento**, o que permite a substituição ou extensão dessas tecnologias de forma controlada.

> **Futuro possível:** suportar múltiplas variações tecnológicas com impacto mínimo no restante da arquitetura, incluindo:

* **Kinesis ↔ Kafka**, ou mesmo conectores via EventBridge Pipes
* **PostgreSQL ↔ MySQL**, desde que respeitado o dicionário de dados e as camadas de abstração
* **DynamoDB ↔ MongoDB**, com controle de schema e versionamento adaptado
* **EMR Serverless ↔ Glue Jobs**, ou outras engines Spark compatíveis

Essa flexibilidade é essencial para adaptar o projeto a diferentes contextos de adoção, restrições de custo, ferramentas já existentes no trabalho ou preferências da equipe técnica.

---

### Abertura do projeto como iniciativa open source

Embora o projeto tenha sido desenvolvido com foco em uma certificação, ele foi também uma **experiência pessoal extremamente gratificante**. Explorar diferentes arquiteturas, linguagens e serviços AWS tornou esse projeto não apenas desafiador, mas também **divertido de construir**.

Com sua arquitetura modular, uso de tecnologias amplamente adotadas e foco em boas práticas, o projeto se mostra um **bom candidato para se tornar uma iniciativa open source**.

A abertura do código permitirá:

* **Compartilhamento com a comunidade** de engenharia de dados e arquitetura serverless
* **Adoção e adaptação por outros times** com necessidades semelhantes
* **Colaborações externas**, incluindo feedbacks, melhorias, plugins e extensões
* **Reconhecimento técnico e aprendizado contínuo** por parte dos autores e contribuidores

> **Futuro possível:** manter a evolução do repositório com refinamento contínuo da base de código, documentação no padrão open source, testes mais abrangentes, exemplos práticos, cronograma de evolução e abertura para contribuições da comunidade.

Esse movimento amplia o alcance do projeto, estimula contribuições e cria um ciclo positivo de evolução contínua com apoio da comunidade.

---

[Voltar para a página inicial](../README.md#documentação) | [Próximo: Considerações Finais](considerations.md)
