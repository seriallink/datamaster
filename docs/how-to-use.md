# Como Usar o Data Master CLI

O **Data Master CLI** é a ferramenta principal para operar todos os componentes do projeto, incluindo provisionamento de infraestrutura, deploy de funções Lambda, gerenciamento de catálogo no Glue, geração de dashboards e simulação de dados.

> ⚠️ A **sequência recomendada de execução para provisionamento** será apresentada na próxima seção da documentação.

---

## Sessão Interativa

Inicie a interface interativa com:

```bash
datamaster
```

Durante a sessão, o CLI oferece:

* Autenticação com AWS (Profile ou Access/Secret)
* Execução de comandos como `deploy`, `catalog`, `grafana`, etc.
* Retenção de variáveis de ambiente
* Suporte a múltiplos comandos encadeados

---

## Comandos Disponíveis

### `artifacts`

Lista os artefatos Lambda embarcados no CLI:

```bash
datamaster artifacts
```

---

### `auth`

Autentica com a AWS:

```bash
datamaster auth
```

O CLI guiará você para usar:

* Um **AWS Profile nomeado**, ou
* Uma **Access Key e Secret**

---

### `catalog`

Cria ou atualiza tabelas no Glue Catalog com base nos schemas do Aurora:

```bash
datamaster catalog
```

Para uma camada específica:

```bash
datamaster catalog --layer bronze
```

Para tabelas específicas em uma camada:

```bash
datamaster catalog --layer bronze --tables brewery,beer
```

> As tabelas serão criadas se não existirem ou atualizadas se já existirem.

---

### `clear`

Limpa a tela da sessão interativa:

```bash
datamaster clear
```

---

### `deploy`

Provisiona a infraestrutura via CloudFormation:

```bash
datamaster deploy
```

Ou para uma stack específica:

```bash
datamaster deploy --stack storage
```

Use `datamaster stacks` para listar as stacks disponíveis.

---

### `exit`

Sai da sessão interativa:

```bash
datamaster exit
```

---

### `grafana`

Cria ou atualiza dashboards no Grafana:

```bash
datamaster grafana
```

Para uma dashboard específica:

```bash
datamaster grafana --dashboard analytics
```

Dashboards disponíveis:

* `analytics`: mostra top beers, breweries, drinkers, styles e volume de reviews
* `logs`: (futuro) logs de erro e execução
* `costs`: (futuro) métricas de custo e alertas

---

### `lambda`

Deploy de funções Lambda usando artefatos embarcados:

```bash
datamaster lambda
```

Para uma função específica com configurações customizadas:

```bash
datamaster lambda --name processing-controller --memory 256 --timeout 120
```

> Os artefatos são enviados automaticamente para o S3.
> Veja todas as funções disponíveis com: `datamaster artifacts`.

---

### `migration`

Executa scripts de migração no Aurora.

#### Executar todos os scripts embarcados, na ordem:

```bash
datamaster migration
```

Scripts incluídos (em ordem de execução):

* `001_create_dm_core.sql`
* `002_create_dm_view.sql`
* `003_create_dm_mart.sql`

#### Executar um script específico:

```bash
datamaster migration --script database/migrations/001_create_dm_core.sql
```

---

### `seed`

Insere datasets no Aurora (streaming) ou S3 (batch):

```bash
datamaster seed
```

Para um dataset específico:

```bash
datamaster seed --file beer
```

Datasets disponíveis:

* `beer`, `brewery`, `profile`: inserção em Aurora (streaming)
* `review`: upload para S3 (batch)

---

### `stacks`

Exibe todas as stacks disponíveis para o comando `deploy`:

```bash
datamaster stacks
```

---

### `whoami`

Exibe a identidade atual da sessão AWS:

```bash
datamaster whoami
```

Mostra:

* AWS Account ID
* IAM Role/User
* Região

---

## Dicas Gerais

* Use `help` com qualquer comando para ver suas opções:

```bash
datamaster help deploy
```

---

[Voltar para a página inicial](../README.md)