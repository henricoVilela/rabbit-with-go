# rabbit-with-go
Projeto para ler e enviar mensagem no rabbitmq

## Iniciar o container do rabbit
- Ter o [Docker](https://docs.docker.com/desktop/install/linux-install/) instalado
- Navegar a ate a pasta do projeto e executar ```docker compose up -d```
- A pasta ```./data``` é onde fica armazenado as informações do rabbit caso derrube o container

## Usar o writer
- Ter o [GO](https://go.dev/doc/install)
- Foi configurado para criar se não tiver, duas filas: ```sys1.v1.messages``` e ```sys2.v1.messages```.
- Em seguida criar um exchange chamada ```notify``` e faz o bind das filas com as key (sys1.message e sys2.message).
- Para enviar a mensagem para fila executar ```go run ./writer/main.go "Key" "Texto da mensagem"```

## Usar o reader
- Ter o [GO](https://go.dev/doc/install)
- Para ler as mensagens para fila executar ```go run ./reader/main.go "nome da fila"```
