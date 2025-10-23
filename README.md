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


## Licença e Uso

_Este Software é distribuídos sob licença GNU GENERAL PUBLIC LICENSE Version 3 onde é permitido o uso, modificação e distribuição, mantendo o padrão de
software livre. O autor deste software NÃO OFERECE NENHUMA GARANTIA DE USO E FUNCIONAMENTO e o seu uso deve ser feito por conta e risco sem qualquer
responsabilidade do autor. O seu uso determina que reconhece e aceita os Termos de Uso aqui aplicados. Licensa GPL3 https://www.gnu.org/licenses/gpl-3.0.pt-br.html_
