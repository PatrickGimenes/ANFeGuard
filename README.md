# ANFeGuard — Documentação do Projeto

## Visão Geral

**ANFeGuard** é uma ferramenta de monitoramento e gerenciamento escrita em **Go** que:

- Monitora **recursos do sistema** (CPU, memória, disco);
- Fornece uma **API web** e frontend estático para dashboards;
- Gerencia **serviços do Windows**;
- Envia **alertas por e-mail** quando limites são ultrapassados ou os serviços monitorados caem;
- Armazena métricas e logs em banco de dados.

O projeto foi refatorado para rodar como **serviço do Windows** usando `github.com/kardianos/service`.

---

## Stack

| Tecnologia | Finalidade |
|------------|------------|
| **Go** | Aplicação principal, serviço, API |
| **HTML/CSS/JS** | Interface frontend |
| **PostgreSQL** | Banco de dados (configurável via `.env`) |
| `kardianos/service` | Integração como serviço do Windows |
| `htmx.org` | Chamadas dinâmicas da interface para a API |

---

## Estrutura do Projeto

```
ANFeGuard/
├── cmd/                 # Entry point da aplicação
├── controllers/         # Handlers HTTP
├── database/            # Conexão e modelos do banco
├── email/               # Templates e envio de e-mails
├── logs/                # Manipulação de arquivos de log
├── monitor/             # Lógica de monitoramento
├── public/              # Frontend estático (HTML/CSS/JS)
├── router/              # Definição das rotas da API
├── sysinfo/             # Utilitários de informações do sistema
├── version/             # Constantes de versão
├── winservice/          # Integração com serviço Windows
├── .env_example         # Exemplo de variáveis de ambiente
├── README.md            # Instruções do projeto
├── go.mod
└── go.sum
```

---

## Módulos e Responsabilidades

### cmd

Contém a função principal que inicia o serviço e o servidor web. Configura a aplicação para rodar como **serviço do Windows**.

---

### controllers

Define todos os **handlers HTTP** usados pelo frontend:

| Rota | Handler |
|------|---------|
| `GET /api/servicos` | Lista serviços monitorados |
| `POST /api/servico` | Cria um novo serviço monitorado |
| `POST /api/servico/restart` | Reinicia serviço do Windows |
| `GET /api/portas` | Lista portas monitoradas |
| `POST /api/porta` | Adiciona uma porta monitorada |
| `GET /api/logs` | Lista logs do serviço |

---

### database

Responsável pela conexão e consultas ao banco de dados:

- Suporta **PostgreSQL** (configurado via `.env`);
- Contém modelos para logs e serviços monitorados.

---

### email

Gerencia envio de e-mails de alerta:

- Carrega templates HTML do disco;
- Envia e-mails de forma **assíncrona** em goroutines;
- Suporta diferentes templates dependendo do status do serviço.

---

### monitor

Contém a lógica de monitoramento contínuo:

- Verifica uso de **CPU**, **memória** e **disco**;
- Dispara alertas quando os limites são ultrapassados.

---

### logs

Responsável pela criação e abertura de arquivos de log:

- Cria diretório `Logs/` ao lado do executável;
- Escreve logs em arquivo (e no console durante desenvolvimento).

---

### public

Contém os assets do frontend:

- `index.html`
- `servico.html`
- `cadastro.html`
- `logs.html`
- CSS e JS

A interface utiliza **htmx** para atualizar partes da página via chamadas API.

---

### router

Mapeia URLs para os **handlers** definidos em `controllers`. Serves arquivos estáticos da pasta `public/`.

---

### sysinfo

Pacote utilitário para coleta de informações do sistema (disco, memória, CPU).

---

### winservice

Código relacionado ao registro e execução do ANFeGuard como **serviço do Windows**.

---

## Configuração

Variáveis de ambiente (via `.env`):

| Variável | Descrição |
|----------|-----------|
| `PERIOD` | Intervalo de monitoramento em segundos |
| `EMAIL_HOST` | Servidor SMTP |
| `EMAIL_PORT` | Porta SMTP |
| `EMAIL_USER` | Usuário SMTP |
| `EMAIL_PASS` | Senha SMTP |
| `NOTIFY_EMAILS` | Destinatários separados por vírgula |
| `MAX_RETRIES` | Máximo de tentativas para alertas |
| `THRESHOLD_WARNING` | Limite de recursos para alerta |
| `DISK` | Caminho do disco a ser monitorado |
| `API_PORT` | Porta do servidor HTTP |

> Pode-se copiar `.env_example` e ajustar os valores conforme necessário.

---

## Configuração Local

### 1. Clonar repositório

```sh
git clone https://github.com/PatrickGimenes/ANFeGuard.git
cd ANFeGuard
```

### 2. Instalar dependências

```sh
go mod tidy
```

### 3. Compilar

```sh
go build -o ANFeGuard.exe
```

### 4. Rodar em desenvolvimento

```sh
go run cmd/main.go
```

### 5. Rodar como serviço Windows

```sh
ANFeGuard.exe install
ANFeGuard.exe start
```

---

## Resumo da API

### Health

```sh
GET /api/health
```

Retorna: `{"status":"ok"}`

### Metrics

```sh
GET /api/metrics
```

Retorna métricas do sistema em JSON.

### Services

```sh
GET /api/servicos
```

Lista serviços monitorados.

### Criar Serviço

```sh
POST /api/servico
```

Recebe formulário com os dados do serviço.

---

## Testes e Validação

Não existem testes automatizados ainda, em desenvolvimento

---

## Observações

- O frontend usa `htmx.org` para chamadas AJAX sem SPA.  
- Todos os arquivos estáticos, templates e logs devem usar **paths absolutos** quando rodando como serviço.  
- Logs e templates devem ser carregados em relação ao diretório do executável.

---

## Resumo

ANFeGuard é um projeto modular e leve para:

Monitorar serviços e recursos do Windows  
Expor APIs para integração com frontend  
Enviar alertas por e-mail com templates customizados

## Contribuindo

Se você deseja contribuir para o projeto, siga as etapas abaixo:

1. Faça o fork deste repositório.
2. Crie uma nova branch para sua feature ou correção.
3. Faça suas alterações.
4. Envie um pull request detalhando as mudanças feitas.


### Padrão de branchs e commits

| Feat	| Nova funcionalidade|
|-------|--------------------|
| fix | Correção de bug |
| chore | Coisas internas (config, build, etc) |
| refactor | Melhoria sem mudar  comportamento |
| perf | Melhoria de performance |
| test | Testes |
| docs | Documentação |

Exemplos:
- feat(monitoring): add resource usage alert system
- refactor(alert): simplify alert control logic
- perf(monitoring): optimize resource usage checks
- chore(service): update service configuration


Feito com ❤️ por Patrick
