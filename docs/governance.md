# Governança e Segurança de Dados

A governança de dados no projeto foi estruturada com foco em **segurança, rastreabilidade e escalabilidade**. As principais iniciativas implementadas visam garantir o controle sobre o acesso, o tratamento de dados sensíveis e a transparência do uso.

---

## Lake Formation habilitado

O **Lake Formation** foi habilitado para centralizar o controle de acesso aos dados do Data Lake. Essa escolha permite:

- **Permissões granulares** por banco, tabela, coluna e até linha
- Integração com **IAM Identity Center (SSO)** e suporte a **roles específicas** por consumidor de dados
- **Auditoria nativa** de acessos via CloudTrail

A stack `dm-governance` registra o bucket do Lake como um recurso gerenciado e aplica permissões completas (ALL) sobre os bancos `dm_stage`, `dm_bronze`, `dm_silver` e `dm_gold` para a role `dm-lakeformation-role`.

### Como validar via CloudTrail

Para auditar ações de acesso e controle aplicadas pelo Lake Formation:

1. Acesse o console do **AWS CloudTrail**
2. Vá para o menu **Event history**
3. Filtre por:
    - **Event source:** `lakeformation.amazonaws.com`
    - (Opcional) **User name:** `dm-demo-user` ou a role usada para execução
4. Verifique eventos como:
    - `GrantPermissions` — concessão de acesso
    - `GetDataAccess` ou `GetTable` — tentativas de leitura

> Esses eventos confirmam que o Lake Formation está sendo usado como mecanismo central de controle e que as ações estão sendo auditadas em tempo real.

---

## Detecção automática de PII com AWS Comprehend

Durante o processo de ingestão na camada **bronze**, o projeto utiliza o **AWS Comprehend** para detectar dinamicamente **informações pessoais identificáveis (PII)**. Essa análise é aplicada a amostras dos arquivos e serve como mecanismo de proteção automatizado.

Quando campos com alto score de PII são detectados:

- Um processo automático aplica **mascaramento de dados** antes do armazenamento em Parquet
- Não há necessidade de configuração manual por coluna ou tabela
- Garante que dados sensíveis não fiquem expostos mesmo nas camadas internas do Data Lake

### Como validar no Athena

Para verificar o funcionamento do mascaramento de PII:

1. Acesse o console do **Athena**
2. Selecione o banco de dados `dm_bronze`
3. Execute a query:

```sql
SELECT profile_name, email FROM dm_bronze.profile LIMIT 25;
```

* O campo `email` deve aparecer **mascarado** (por exemplo, `***@***.com`), conforme a lógica aplicada no job de ingestão
* Já o campo `profile_name`, embora frequentemente se pareça com um nome de pessoa, **não foi considerado PII com alto grau de confiança** nas amostras avaliadas — o que evitou **falsos positivos e mascaramentos indevidos** que poderiam prejudicar análises legítimas

---

## Usuário demo com acesso restrito

Para validar a aplicação das permissões, foi criado um **usuário IAM de demonstração**, com as seguintes características:

- Acesso restrito exclusivamente à camada **gold** 
- Políticas limitam o uso do Athena apenas às tabelas do banco `dm_gold`
- Tentativas de acessar dados em outras camadas (como `dm_silver` ou `dm_bronze`) resultam em erro de permissão

### Como ativar o usuário demo

1. Execute o comando abaixo via CLI para provisionar os recursos:

    ```text
    >>> deploy --stack demo
    ```

2. Esse comando cria automaticamente:

* O **usuário IAM**: `dm-demo-user`
* A **senha inicial**: `DemoUser123!` (será exigido redefinir no primeiro login)
* Um **bucket de resultados do Athena**: `dm-demo-<account_id>`

---

### Como validar o acesso via console

1. Acesse o **console da AWS** e entre como o usuário `dm-demo-user`
2. Após o login, altere a senha conforme solicitado

---

### Como configurar o Athena

1. No console do Athena, você verá a seguinte mensagem:

> *"Before you run your first query, you must specify a query result location in Amazon S3."*

2. Clique no botão **"Edit settings"**
3. No campo **"Query result location"**, informe o bucket criado para esse usuário:

```
s3://dm-demo-<account_id>/
```

4. Salve as configurações

---

### Consulta autorizada

1. No painel lateral, você verá apenas o banco `dm_gold` e suas tabelas
2. Execute a query abaixo:

```sql
SELECT * FROM dm_gold.top_beers_by_rating LIMIT 10;
```

> Essa consulta deve retornar resultados normalmente.

---

### Consulta não autorizada

Agora tente acessar uma tabela de outra camada, por exemplo:

```sql
SELECT * FROM dm_silver.beer LIMIT 10;
```

O Athena deve exibir a seguinte mensagem de erro:

> **"Insufficient permissions to execute the query"**

Isso confirma que o controle de acesso está funcionando corretamente e o usuário só tem visibilidade sobre os dados autorizados.

---

[Voltar para a página inicial](../README.md#documentação) | [Próximo: Roadmap Técnico e Melhorias Futuras](../docs/roadmap.md)