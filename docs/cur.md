# Como habilitar o Cost and Usage Report (CUR) e integrá-lo ao Athena

Este passo a passo mostra como configurar a exportação de relatórios de custo detalhados da AWS em formato otimizado para consulta com Athena, Glue e processamento analítico. Utiliza o novo formato **CUR 2.0** com arquivos **Parquet**, estrutura limpa e integração facilitada.

---

## Pré-requisitos

1. **Conta root** ou IAM user com:

    * Permissão para acessar a área de **Billing**
    * Policy gerenciada **`Billing`** anexada

2. O acesso de IAM a Billing deve estar ativado em:

    * **Account Settings > IAM user and role access to Billing information: Activated**

---

## Etapas para habilitar o CUR

### 1. Acesse o menu de Billing

* Vá até: [https://console.aws.amazon.com/billing/home](https://console.aws.amazon.com/billing/home)
* Navegue até: **Cost and Usage Reports**
* Se estiver usando o modo antigo, clique em **Try new experience**

---

### 2. Export details

* Selecione: **Standard data export**
* **Export name:** `cur`

  > Esse será o nome da tabela Glue e da subpasta de dados no S3

---

### 3. Data table content settings

* **Tipo de relatório:** `CUR 2.0`
  * Essa é a **única versão suportada no novo formato**. Se essa opção **não estiver visível**, você ainda está usando o **modelo legado (deprecado)**.
  * Nesse caso, volte à tela anterior e clique em **"Try new experience"** para ativar o fluxo atualizado.

* Marque: **Include resource IDs**
  * Permite que o relatório contenha colunas com `resource_id`, essenciais para rastrear gastos por recurso.

* **Time granularity:** `Daily`
  * Recomendada para análises detalhadas e dashboards de custo por dia.

* Deixe **desmarcado**: `Split cost allocation data`
  * Ative **apenas se você realmente precisa de alocação detalhada para ECS/EKS**, pois essa opção aumenta o tamanho e a complexidade dos dados.

---

### 4. Data export delivery options

* Frequência: **Daily**
* Formato: **Parquet**
* Versão de arquivo: **Overwrite existing data export file**

---

### 5. Data export storage settings

* **Bucket:** `dm-costs-<account_id>`
  * Este bucket já foi provisionado automaticamente pela stack `storage` e está preparado para receber os dados do CUR.
  * Ao selecionar o bucket durante a configuração do CUR, será exibido um aviso informando que a **bucket policy será sobrescrita** para permitir que os serviços da AWS (como `billingreports.amazonaws.com` e `bcm-data-exports.amazonaws.com`) escrevam os relatórios no bucket.
  * Essa alteração é obrigatória e deve ser aceita para que a exportação funcione corretamente.

* **S3 path prefix:** `dm_costs`
  * Este prefixo será usado como base da estrutura de diretórios no S3 e será também o **nome do Glue Database**.

> O database `dm_costs` é criado via IaC na stack `catalog.yml`, seguindo o mesmo padrão adotado para `dm_bronze`, `dm_silver` e `dm_gold`.

---

### Estrutura final no S3

```
s3://dm-costs-<account_id>/dm_costs/cur/
```

* `dm_costs/` → representa o **banco de dados** (`dm_costs`)
* `cur/` → representa a **tabela** (`cur`) com os dados particionados por `BILLING_PERIOD`

---

### 6. Confirmação após criação

Após finalizar a criação, você verá a listagem no painel **Exports and dashboards**, com as seguintes informações:

* **Export name:** `cur`
* **Status:** Healthy
* **Export type:** `Standard data export`
* **Data table:** `CUR 2.0`
* **S3 bucket:** `dm-costs-<account_id>`

A AWS começará a gerar os arquivos dentro de até **24 horas**.

---

### 7. O que será gerado

A estrutura gerada no S3 será:

```
s3://dm-costs-<account_id>/dm_costs/cur/
├── data/
│   └── BILLING_PERIOD=YYYY-MM/
│       └── <execution_id>/dm_costs-00001.snappy.parquet
└── metadata/
    └── BILLING_PERIOD=YYYY-MM/
        └── <execution_id>/
            └── dm_costs-Manifest.json
```

#### **`data/`**

Contém os arquivos `.parquet` particionados por `BILLING_PERIOD`, no formato otimizado para Athena e Glue.

#### **`metadata/`**

Contém os metadados do relatório: colunas, período, e localização dos dados.

---

### 8. Criação manual da tabela no Athena

Após o primeiro ciclo de geração (que pode levar até 24 horas), você verá no bucket `dm-costs-<account_id>` a estrutura com a pasta `data/BILLING_PERIOD=...`.

**Somente após isso**, vá até o **Athena** e execute o script `CREATE EXTERNAL TABLE` para registrar os dados no Glue Catalog.

> Utilize o script: [`create_table_cur.sql`](../catalog/scripts/create_table_cur.sql)

No script, o campo `LOCATION` estará com o valor:

```sql
LOCATION 's3://dm-costs-<account_id>/dm_costs/cur/data/';
```

**Substitua `<account_id>` pelo número real da sua conta AWS antes de executar.**

---

Após criar a tabela, adicione a partição correspondente ao report do mês que foi gerado. Por exemplo, para o mês de julho de 2025:

```sql
ALTER TABLE dm_costs.cur ADD PARTITION (BILLING_PERIOD='2025-07')
LOCATION 's3://dm-costs-<account_id>/dm_costs/cur/data/BILLING_PERIOD=2025-07/';
```

---
