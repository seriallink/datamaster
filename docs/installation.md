# Instalação do Data Master CLI

O **Data Master CLI** é o ponto de entrada para toda a experiência deste projeto. Em vez de depender de interfaces gráficas ou scripts manuais, **todo o fluxo é guiado por uma interface de linha de comando interativa**, projetada para ser simples, clara e totalmente automatizada.

> A proposta é oferecer uma jornada estruturada e realista, simulando o trabalho de um engenheiro de dados em um ambiente moderno baseado em nuvem.

### Por que uma CLI?

* **Reprodutibilidade**: evita passos manuais e garante consistência entre execuções.
* **Automação total**: provisionamento, autenticação, deploy e processamento — tudo controlado por comandos.
* **Foco na experiência técnica**: ideal para simulações hands-on e uso real por engenheiros.

---

Com os [requisitos atendidos](pre-requirements.md), siga os passos abaixo para instalar o CLI localmente:

### 1. Clone o repositório

```bash
git clone https://github.com/seriallink/datamaster.git
cd datamaster
```

---

### 2. Compile e instale o CLI

Execute o comando abaixo para compilar o projeto e instalar o binário:

```bash
go install ./...
```

Esse comando compila o CLI e o instala no diretório padrão de binários do Go:

```
$GOPATH/bin
```

> Por padrão, o Go instala os binários em `~/go/bin` no Linux/macOS, e em `%USERPROFILE%\go\bin` no Windows.

---

### 3. Adicione o Go binário ao seu `PATH`

#### Linux/macOS:

Se ainda não estiver no `PATH`, adicione o diretório ao seu shell (`.bashrc`, `.zshrc`, etc.):

```bash
export PATH="$HOME/go/bin:$PATH"
```

E para tornar isso permanente:

```bash
echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.bashrc
```

Substitua `.bashrc` por `.zshrc` se estiver usando Zsh.

---

#### Windows:

O diretório geralmente é:

```
C:\Users\<SeuUsuário>\go\bin
```

Para adicioná-lo ao `PATH`:

1. Abra o **Menu Iniciar** e busque por `variáveis de ambiente`.
2. Clique em **Editar variáveis de ambiente do sistema**.
3. Clique em **Variáveis de Ambiente...**.
4. Em **Variáveis do usuário**, selecione `Path` e clique em **Editar**.
5. Adicione o caminho: `C:\Users\<SeuUsuário>\go\bin`
6. Clique em **OK** em todas as janelas.

> Após isso, reinicie o terminal (PowerShell, CMD ou Git Bash) para que a mudança tenha efeito.

---

### 4. Verifique se o CLI está funcionando

Abra um terminal e execute:

```bash
datamaster
```

---

### 5. Alternativa: execução direta com `go run`

Se preferir, você também pode executar o CLI diretamente a partir do diretório do projeto, sem precisar instalar o binário no `PATH`:

```bash
go run ./main.go
```

--- 

Se tudo estiver certo, você verá a interface interativa do Data Master CLI.

![cli-welcome-screen.png](../assets/cli-welcome.png)

---

[Voltar para a página inicial](../README.md#documentação) | [Próximo: Utilização do Data Master CLI](how-to-use.md)