### ✅ Como habilitar as métricas de billing no CloudWatch (via conta root)

1. Acesse o **Console da AWS** logado com a **conta root**.

2. Vá até o menu **Billing** (ou "Faturamento").

3. No menu lateral, clique em **Billing preferences** (Preferências de faturamento).

4. Na seção **Alert preferences**, clique no botão **Edit** (Editar).

5. Marque a opção:

   > ✅ **Receive CloudWatch billing alerts**
   > *(Receber alertas de faturamento via CloudWatch)*

   ⚠️ **Atenção:** essa opção **não pode ser desativada depois** de habilitada.

6. Clique no botão **Update** (Atualizar) para salvar.

---

### 🕐 E agora?

* Após isso, as métricas estarão disponíveis no CloudWatch no namespace `AWS/Billing`.
* As métricas são atualizadas **diariamente**.
* Acesse sempre a **região us-east-1 (N. Virginia)** para visualizá-las.

Perfeito, Marcelo. Aqui está o tutorial revisado, **sem a criação do bucket**, assumindo que o bucket já foi criado pela stack `storage`.

---

# 🧾 Tutorial: Ativar o Cost and Usage Report (CUR) com Athena na AWS

## 📌 Objetivo

Habilitar relatórios de custo detalhados (CUR) na AWS e integrá-los ao Amazon Athena para análise com SQL. Os dados são entregues em formato **Parquet**, com granularidade **horária**, e armazenados em um bucket S3 já existente.

---

## ✅ Pré-requisitos

* Acesso à **conta root** (utilizado durante a configuração).
* Bucket S3 criado previamente via stack `storage` com nome semelhante a `dm-costs-<account-id>`.
* Permissão para configurar CUR e acessar Athena.

---

## 📊 Etapa 1 — Ativar o Cost and Usage Report (CUR)

### Caminho:

1. Vá para o console da AWS.
2. Acesse **Billing → Cost & Usage Reports**.
3. Clique em **Create report**.

### Preencha os campos:

* **Report name:** `dm-costs-report`
* **Include resource IDs:** ☑️ (habilitado para permitir detalhamento por recurso)
* **Data refresh settings:** ☑️ Opt-in
* **Time granularity:** `Hourly`
* **Report versioning:** `Overwrite existing report`
* **Compression type:** `Parquet`
* **Integration:** ☑️ Amazon Athena

---

## 📦 Etapa 2 — Configurar destino no S3

* **Bucket:** `dm-costs-<account-id>` (provisionado via stack `storage`)
* **Prefixo S3:** `reports/`

O caminho final dos arquivos será:

```
s3://dm-costs-<account-id>/reports/dm-costs-report/date-range/...
```

---

## 🧠 Etapa 3 — Criar a tabela no Athena

Apesar de marcar a integração com Athena, a AWS **não cria a tabela automaticamente**. Você precisa executar um script SQL para criar a tabela no Athena com base no layout dos arquivos gerados.

Posso gerar esse script para você com base no bucket/prefixo se desejar.

---

## 🧪 Etapa 4 — Consultar no Athena

1. Vá para o **Amazon Athena**.
2. Selecione o database criado via script (ex: `cur_db`).
3. Execute queries como:

```sql
SELECT *
FROM cur_dm_costs_report
WHERE product_product_name = 'Amazon EC2';
```

---

## 📌 Observações importantes

* Os arquivos `.parquet` serão entregues em pastas por intervalo de datas, ex: `20240701-20240731/`.
* Um **arquivo manifest** será incluído para facilitar a leitura com Athena.
* Após isso, você pode:

   * Criar dashboards no **QuickSight**
   * Exportar para outros sistemas
   * Alimentar pipelines de custo (como o `dm-costs.yml`)

---


https://docs.aws.amazon.com/cur/latest/userguide/cur-query-athena.html