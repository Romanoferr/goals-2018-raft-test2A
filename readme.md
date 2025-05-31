# **Relatório sobre o trabalho de Raft**

## A principal ideia do trabalho é implementar a eleição de líder, conforme [especificação do Raft.](https://raft.github.io/raft.pdf)

## Integrantes:
- **Antônio Romano Carvalho Ferreira**
- **Eduardo**

## O objetivo é:
- **Um líder seja eleito**
- **O líder continue sendo o líder se não houver falhas**
- **Um novo líder assuma o controle se o antigo líder falhar ou se os pacotes de/para o antigo líder forem perdidos**

## Ferramentas usadas:
- VS Code
- Github
- Foruns Go


## Dificuldades encontradas:
- Preparação do ambiente - Dificuldades de preparação de ambiente para desenvolvimento em Go. A documentação do go ajudou a resolver os problemas para executar o código do repositório base

> Problemas configuração ambiente go
>
> * Como já possui o go instalado, tive algumas dificuldades de encontrar sua ROOT. 
> * Não houve necessidade de configurar as exportações do GO pois minha máquina já estava configurada
>
> Com o repositório aberto, segui os comandos para a execução da versão base. Acessei o diretório `src/raft` e executei o comando `go test -run 2A` eu tomei os seguintes erros:
> * go: cannot find main module, but found .git/config in C:\Projetos\goals-2018 to create a module there, run:
>
> * `cd ..\.. && go mod init`
>
> Para esse erro, eu voltei a root do repositório e rodei
> * `go mod init <goals-2018/src`
> * `go mod tidy`
>
>Após a correta configuração do Go Module, ao tentar executar de novo o `go test -run 2A`, deu o seguinte erro agora:
> * config.go:11:8: package labrpc is not in `std` (`C:\Program Files\Go\src\labrpc`)
>
>Não sei se é essa é a unica solução, mas bastou colar a pasta `\goals-2018\src\labrpc` para o diretório que a saída da execução mostra (local de instalação do GO) -> `C:\Program Files\Go\src`
>
>Após isso, o `go test -run 2A` rodou com sucesso como mostrado no enunciado do trabalho, exibindo os erros pois a implementação ainda não estava realizada
>
>A partir de agora, criarei um fork do projeto para realizar as alterações necessárias para o desenvolvimento da atividade