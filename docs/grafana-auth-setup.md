# Grafana Authentication Setup (AWS IAM Identity Center / SSO)

Este documento descreve os passos necessários para habilitar o login via AWS IAM Identity Center (SSO) no AWS Managed Grafana para o projeto Data Master.

---

## 📍 Pré-requisitos

- AWS Managed Grafana Workspace já criado
- IAM Identity Center habilitado na conta (região: `us-east-1`)
- Usuário com permissões administrativas para configurar SSO

---

## 🔐 Etapas para configuração do SSO (IAM Identity Center)

### 1. Acessar o IAM Identity Center

Console: https://us-east-1.console.aws.amazon.com/iamidentitycenter/

Verifique se sua instância está ativa e a região correta está selecionada (`us-east-1`).

---

### 2. Criar um novo usuário

- Menu lateral → **Users** → **Add user**
- Preencha os dados:
    - Email: (ex: `marcelo@example.com`)
    - First name: Marcelo
    - Last name: Data Master
- Clique em **Create user**

---

### 3. (Opcional) Criar um grupo

- Menu lateral → **Groups** → **Create group**
- Nome do grupo: `grafana-admins`
- Adicione o usuário criado ao grupo

---

### 4. Atribuir acesso ao Grafana

- Acesse o console do Grafana: https://console.aws.amazon.com/grafana/
- Entre no workspace `dm-grafana-workspace`
- Aba **Authentication** → clique em **Assign users or groups**
- Selecione o usuário ou grupo
- Atribua a permissão **Admin**

---

### 5. Acessar o Grafana via SSO

- Vá para o portal do AWS IAM Identity Center:
    - Exemplo: `https://d-xxxxxxxxxx.awsapps.com/start`
    - URL disponível no painel do IAM Identity Center → seção **Settings summary**
- Faça login com o email criado
- Clique no ícone do Grafana para acessar o workspace autenticado

---

## ⚙️ Observações finais

- Essa configuração permite login com SSO (sem credenciais IAM).
- Pode-se automatizar dashboards com token de API posteriormente.
- Ideal para ambientes temporários ou apresentações (como no caso do projeto Data Master).

---

Marcelo Monaco  
Data Master 2025  
