# Cotação USD/BRL

Projeto simples em Go com dois programas:

* `server.go`: expõe `/cotacao`, busca a cotação e salva no SQLite
* `client.go`: chama o server e grava o valor em um arquivo

## Como rodar

### 1. Subir o server

```bash
go run server.go
```

Vai subir em:

```text
http://localhost:8080/cotacao
```

---

### 2. Rodar o client

Em outro terminal:

```bash
go run client.go
```

Isso vai criar um arquivo:

```text
cotacao.txt
```

Exemplo de conteúdo:

```text
Dólar: 5.12
```
