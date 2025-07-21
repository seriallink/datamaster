# Como Usar o Data Master CLI

O **Data Master CLI** é a ferramenta principal para operar todos os componentes do projeto, incluindo provisionamento de infraestrutura, deploy de funções Lambda, gerenciamento de catálogo no Glue, geração de dashboards e simulação de dados.

> A **sequência recomendada de execução para provisionamento** será apresentada na próxima seção da documentação.

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
>>> artifacts
```

---

### `auth`

Autentica com a AWS:

```bash
>>> auth
```

O CLI guiará você para usar:

* Um **AWS Profile nomeado**, ou
* Uma **Access Key e Secret**

---

### `catalog`

Cria ou atualiza tabelas no Glue Catalog com base nos schemas do Aurora:

```bash
>>> catalog
```

Para uma camada específica:

```bash
>>> catalog --layer bronze
```

Para tabelas específicas em uma camada:

```bash
>>> catalog --layer bronze --tables brewery,beer
```

> As tabelas serão criadas se não existirem ou atualizadas se já existirem.

---

### `clear`

Limpa a tela da sessão interativa:

```bash
>>> clear
```

---

### `deploy`

Provisiona a infraestrutura via CloudFormation:

```bash
>>> deploy
```

Ou para uma stack específica:

```bash
>>> deploy --stack storage
```

Use `stacks` para listar as stacks disponíveis.

---

### `exit`

Sai da sessão interativa:

```bash
>>> exit
```

---

### `grafana`

Cria ou atualiza dashboards no Grafana:

```bash
>>> grafana
```

Para uma dashboard específica:

```bash
>>> grafana --dashboard analytics
```

Dashboards disponíveis:

* `analytics`: mostra top beers, breweries, drinkers, styles e volume de reviews
* `logs`: (futuro) logs de erro e execução
* `costs`: (futuro) métricas de custo e alertas

---

### `lambda`

Deploy de funções Lambda usando artefatos embarcados:

```bash
>>> lambda
```

Para uma função específica com configurações customizadas:

```bash
>>> lambda --name processing-controller --memory 256 --timeout 120
```

> Os artefatos são enviados automaticamente para o S3.
> Veja todas as funções disponíveis com: `artifacts`.

---

### `migration`

Executa scripts de migração no Aurora.

#### Executar todos os scripts embarcados, na ordem:

```bash
>>> migration
```

Scripts incluídos (em ordem de execução):

* `001_create_dm_core.sql`
* `002_create_dm_view.sql`
* `003_create_dm_mart.sql`

#### Executar um script específico:

```bash
>>> migration --script database/migrations/001_create_dm_core.sql
```

---

### `process`

Executa o processamento de uma camada do pipeline (como silver ou gold), orquestrado por Step Functions:

```bash
>>> process --layer silver
```

Para processar apenas tabelas específicas:

```bash
>>> process --layer silver --tables beer,brewery
```

Parâmetros:

* `--layer` (obrigatório): camada a ser processada (`silver`, `gold`, etc.)
* `--tables` (opcional): lista de tabelas separadas por vírgula

> Esse comando substitui execuções manuais de pipelines via console. A orquestração é feita por Step Functions automatizadas.

---

### `seed`

Insere datasets no Aurora (streaming) ou S3 (batch):

```bash
>>> seed
```

Para um dataset específico:

```bash
>>> seed --file beer
```

Datasets disponíveis:

* `beer`, `brewery`, `profile`: inserção em Aurora (streaming)
* `review`: upload para S3 (batch)

---

### `stacks`

Exibe todas as stacks disponíveis para o comando `deploy`:

```bash
>>> stacks
```

---

### `whoami`

Exibe a identidade atual da sessão AWS:

```bash
>>> whoami
```

Mostra:

* AWS Account ID
* IAM Role/User
* Região

---

## Dicas Gerais

* Use `help` com qualquer comando para ver suas opções:

```bash
>>> help deploy
```

---

[Voltar para a página inicial](../README.md#documentação) | [Próximo: Provisionamento do Ambiente](provisioning.md)