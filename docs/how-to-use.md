# 🚀 Como Usar o Data Master CLI

Este guia apresenta os principais comandos e fluxos de uso do **Data Master CLI**, a ferramenta central para provisionamento, simulação, processamento e gestão de dados no projeto Data Master.

---

## 🧑‍💻 Sessão Interativa

Para começar, basta digitar:

```bash
datamaster
````

A sessão interativa guiará você por:

1. Autenticação via AWS Profile ou Access Key
2. Verificação de identidade com `whoami`
3. Fluxo guiado com opções como:

    * [x] Deploy de infraestrutura
    * [x] Simulação de dados (`stream`)
    * [x] Execução do pipeline de processamento (`processing`)
    * [x] Gerenciamento de tabelas no Glue Catalog (`catalog`)

> 💡 A sessão mantém variáveis de ambiente ativas e permite comandos encadeados.

---

## 🔧 Comandos Diretos

Você também pode usar o CLI fora da sessão interativa, com comandos diretos.

### 1. Deploy

Provisiona toda a infraestrutura do projeto (buckets, roles, Glue, Lambda, DMS, etc):

```bash
datamaster deploy
```

> Requer permissões `AdministratorAccess` na conta AWS.

---

### 2. Stream

Simula eventos de inserção em tabelas do Aurora para testar o pipeline de streaming:

```bash
datamaster stream --table customer --count 100
```

Parâmetros:

* `--table`: nome da tabela a ser simulada
* `--count`: número de eventos a gerar

---

### 3. Processing

Executa o processamento de arquivos `.gz` no bucket `dm-stage` e grava Parquet no `dm-datalake`.

```bash
datamaster processing --layer bronze --table customer
```

Também é possível processar um arquivo específico:

```bash
datamaster processing --object raw/customer/file-2025-06-01.json.gz
```

---

### 4. Catalog

Cria ou atualiza tabelas no Glue Catalog com base no schema do Aurora:

```bash
datamaster catalog
```

Ou para atualizar apenas uma camada ou tabela específica:

```bash
datamaster catalog --layer silver --tables order_items,products
```

---

## ✅ Dicas Gerais

* Use `--help` em qualquer comando para ver opções disponíveis:

```bash
datamaster processing --help
```

* Todos os comandos respeitam o ambiente (`--env`) e nível de log (`--loglevel`).

---

## 🧪 Fluxo sugerido para testes

1. Execute `datamaster` e autentique-se
2. Faça o deploy da infraestrutura
3. Simule dados com `stream`
4. Verifique os arquivos `.gz` no bucket `dm-stage`
5. Execute `processing` para gerar os Parquets
6. Confirme a presença dos dados no `dm-datalake`

---

Se desejar, podemos agora expandir com:

* Sessões avançadas (ex: comandos para debug ou integração com Step Functions)
* Comportamento detalhado de logs
* Lista de erros comuns e como resolvê-los

Deseja que eu crie o arquivo `how-to-use.md` com esse conteúdo e envie como resposta completa?
