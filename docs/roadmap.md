# Roadmap Técnico e Melhorias Futuras

Embora o projeto atenda aos objetivos principais definidos, que foi a criação de um data lake serverless e de baixo custo, diversos pontos foram identificados como oportunidades de **evolução arquitetural, operacional e analítica**. Algumas dessas melhorias não foram implementadas por restrições técnicas, escopo ou prioridade, mas representam **direções estratégicas importantes** para ganho de eficiência, escalabilidade e governança no médio e longo prazo.

---

### Adoção futura do Apache Iceberg com Go (`iceberg-go`)

Atualmente, as camadas **silver** e **gold** do projeto são processadas integralmente em **PySpark**, com persistência dos dados no formato **Iceberg**. Essa abordagem oferece robustez e flexibilidade, mas depende da infraestrutura do **EMR Serverless**, o que impacta diretamente o **custo por job**, especialmente para transformações simples ou cargas frequentes.

O projeto [iceberg-go](https://github.com/apache/iceberg-go), mantido pela própria Apache, está em evolução ativa e representa uma **alternativa estratégica** para o futuro. Com sua maturação, será possível:

* **Eliminar a dependência do PySpark** em jobs mais leves
* **Reduzir custos operacionais**, executando transformações simples diretamente em binários Go via Lambda ou ECS
* **Unificar a base tecnológica**, aproveitando a simplicidade e velocidade de desenvolvimento do Go em outras partes do projeto

> Assim que o `iceberg-go` atingir estabilidade para operações de leitura e escrita com Iceberg em ambientes de produção, será avaliada sua adoção como engine principal em camadas analíticas e no processamento incremental.

---

### Visualização de dados operacionais (DynamoDB)

Os dados de controle operacional — armazenados na tabela `dm-processing-control` no DynamoDB — são parte essencial do monitoramento dos pipelines. No entanto, o **plugin oficial para Amazon DynamoDB no Grafana** é exclusivo do **plano Enterprise**.

Essa limitação impede, por ora, a exibição nativa de informações como tentativas com erro, arquivos em processamento e duração de execuções por job.

> **Futuro possível:** replicar periodicamente os dados do DynamoDB para o S3, tornando-os acessíveis via Athena e, consequentemente, integráveis ao Grafana.

---

### Validação de qualidade de dados automatizada (Data Quality)

Apesar dos pipelines estarem estruturados, não há verificação sistemática de integridade ou qualidade dos dados ingeridos ou transformados (ex: nulos indevidos, duplicidade, range incorreto).

> **Futuro possível:** adoção de frameworks como [Deequ](https://github.com/awslabs/deequ), ou lógica custom em PySpark para validação automática com alertas via SNS/Grafana.

---

### Políticas de expurgo nas camadas Raw e Bronze

As camadas **raw** e **bronze** acumulam arquivos de dados históricos utilizados no processamento incremental das camadas superiores. No entanto, **não há políticas de expurgo automatizadas**, o que pode gerar crescimento desnecessário no armazenamento ao longo do tempo.

Essa retenção indefinida impacta diretamente o **custo com S3** e dificulta a **gestão do ciclo de vida dos dados**.

> **Futuro possível:** definir e aplicar **regras de expurgo baseadas em data de criação**, utilizando **S3 Lifecycle Rules**. Como exemplo, arquivos `.gz` da camada raw podem ser excluídos após 90 dias, e arquivos Parquet da bronze após 180 dias, considerando que os dados processados já foram consolidados nas camadas analíticas.

---

### Governança e detecção de dados sensíveis na camada Raw

A camada **raw** pode conter dados brutos não estruturados ou sem garantias de sanitização, o que impõe riscos de **exposição acidental de informações sensíveis (PII)**. Atualmente, **não há mecanismo automatizado** para identificar ou classificar esses dados.

> **Futuro possível:** integração com o **Amazon Macie**, que permitirá:

* Escaneamento automático dos objetos armazenados no S3
* Detecção de **dados pessoais identificáveis (PII)** e outros conteúdos sensíveis
* Geração de **alertas e classificações de sensibilidade** para reforçar a governança e mitigar riscos de conformidade

---

[Voltar para a página inicial](../README.md#documentação)