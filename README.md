# gprice
Google Cloud Platform Price List from SKU

Ferramenta de extração dos dados da Price List do Google GCP para análise e validação.

## Overview

Este utilitário realiza o processamento de dados a partir do relatório de Cost Table (Tabela de Custo de mês fechado) do Google GCP e valida o preço unitário a partir da Price List online do Google contido em https://cloud.google.com/skus?currency=USD


## Instalação

Este código foi escrito em GOLANG, para instalação do Go siga o passo a passo disponível em [https://go.dev/doc/tutorial/getting-started#install](go.dev).

Download e preparação
```
git clone https://github.com/mitvix/gprice
cd gprice
go build -o gprice main.go
sudo mv gprice /usr/local/bin/gprice
```

## Uso

```
gprice arquivo_table_cost.csv valor_do_dólar

ou

go run main.go arquivo_table_cost.csv valor_do_dólar
```

### Exemplo

```
gprice Table_Price_GCP_2025-08.csv 5,7004
```
