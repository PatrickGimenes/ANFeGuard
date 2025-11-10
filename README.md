# ANFeGuard

## Tecnologias

- **Golang**
- **PostgreSQL**

## Funcionalidades

# Como Rodar o Projeto

## 1. Pré-requisitos

Antes de começar, certifique-se de ter o seguinte instalado em sua máquina:

- Go (versão 1.24.2 ou superior)
- Docker (para rodar o PostgreSQL de forma fácil)
- PostgreSQL (caso queira rodar localmente sem o Docker)

## 2. Clonando o Repositório

Clone o repositório para sua máquina local:

```bash
git clone https://github.com/PatrickGimenes/ANFeGuard
cd ANFeGuard
```

## 3. Instalação das Dependências

Instale as dependências do projeto utilizando o Go:

```bash
go mod tidy
```

## 4. Rodando o Servidor

Para rodar o servidor, execute:

```bash
go run cmd/main.go
```

O servidor irá rodar na porta 8080 por padrão. Você pode acessar a API no endereço:

```bash
http://localhost:8080
```

## Contribuindo

Se você deseja contribuir para o projeto, siga as etapas abaixo:

1. Faça o fork deste repositório.
2. Crie uma nova branch para sua feature ou correção.
3. Faça suas alterações.
4. Envie um pull request detalhando as mudanças feitas.

Feito com ❤️ por Patrick
