package main

/*
 * Código simples em Go para obtenção do preço de lista por SKU a
 * partir do relatório de "Table Cost" da Console do Google GCP.
 * Google SKU URI: https://cloud.google.com/skus?currency=USD
 * Licença: GNU GENERAL PUBLIC LICENSE Version 3 (GPL3)
 * O autor deste software NÃO OFERECE NENHUMA GARANTIA DE USO E FUNCIONAMENTO
 * Repositório: https://github.com/mitvix/gprice
 */

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// URI para requisição via cURL
const googleCloudUrl = "https://cloud.google.com/_/GoogleCloudUxWebAppCgcUi/data/batchexecute?rpcids=jBDUmc&source-path=%2Fskus&f.sid=711971036412359606&bl=boq_cloud-ux-webapp-cgc-ui_20251015.07_p0&hl=pt_br&soc-app=1&soc-platform=1&soc-device=1&_reqid=839130&rt=c"

// Payload (com %s para substituição com ID da SKU)
const dataRawFormat = "f.req=%%5B%%5B%%5B%%22jBDUmc%%22%%2C%%22%%5B%%5C%%22%s%%5C%%22%%2C%%5C%%22USD%%5C%%22%%2Cnull%%2C30%%5D%%22%%2Cnull%%2C%%22generic%%22%%5D%%5D%%5D&"
const refererFormat = "https://cloud.google.com/skus?currency=USD&filter=%s&hl=pt_br"

// Regex para encontrar o preço a partir da resposta do servidor em json. O padrão busca "número.número USD" ex: 0.075 USD per 1 hour
var priceRegex = regexp.MustCompile(`"([0-9]+\.[0-9]+\sUSD\s[^"]+)"`) // "([0-9]+\.[0-9]+\sUSD\sper\s[^"]+)"

// Executa o curl externo usando os/exec.Command (Pré-req: curl instalado - Repositório Oficial: https://github.com/curl/curl)
func fetchPriceUsingCurlCommand(sku string) (string, error) {
	// 1. Constrói o payload e o referer dinâmicos
	dataRaw := fmt.Sprintf(dataRawFormat, sku)
	referer := fmt.Sprintf(refererFormat, sku)

	// 2. Define os argumentos para o comando 'curl'
	args := []string{
		"-s",         // Silencioso
		"-X", "POST", // Método POST
		googleCloudUrl,

		// Headers
		"-H", "accept: */*",
		"-H", "content-type: application/x-www-form-urlencoded;charset=UTF-8",
		"-H", "origin: https://cloud.google.com",
		"-H", "referer: " + referer,
		"-H", "x-same-domain: 1",

		// Payload Raw
		"--data-raw", dataRaw,
	}

	// 3. Cria o comando
	cmd := exec.Command("curl", args...)

	// 4. Executa o comando e captura a saída (stdout)
	output, err := cmd.Output()

	if err != nil {
		return string(output), fmt.Errorf("falha ao executar curl para SKU %s: %w", sku, err)
	}

	// 5. Retorna a saída
	return string(output), nil
}

// Aplica o regex para extrair o valor de preço
func extractPrice(rawData string) string {
	// Procura as correspondências usando a regex
	matches := priceRegex.FindStringSubmatch(rawData)

	if len(matches) > 1 {
		return matches[1]
	}

	return "Preço não encontrado/Regex falhou"
}

func main() {

	// Define número de argumentos Padrão: run_program nome_do_arquivo.csv valor_do_dolar_do_mês
	// IMPORTANTE! O valor de dólar do mês de referência deve ser repassado como 3º argumento
	if len(os.Args) < 3 {
		fmt.Println("Ferramenta de extração dos dados da Price List do Google GCP para análise e validação.")
		fmt.Printf("Use: %s <Nome_do_Arquivo.csv> <dolar_google>\n", os.Args[0])
		os.Exit(0)
	}

	// Variáveis para reescrita do novo CSV com dados complementares
	var inputFilePath = os.Args[1]                           // 1º Argumento nome do arquivo
	var inputDolar = os.Args[2]                              // 2º Argumento valor do dólar google
	var targetColumn = "Código SKU"                          // Usado para encontrar a coluna no CSV
	var newColumnPriceDolar = "Dólar Google"                 // Novo cabeçalho
	var newColumnCustoUSD = "Custo (USD)"                    // Novo cabeçalho
	var newColumnPriceUnit = "Preço de Lista"                // Novo cabeçalho
	var newColumnPriceModel = "Modelo de cobrança"           // Novo cabeçalho
	var newColumnPriceLocal = "Preço unitário cobrado (USD)" // Novo cabeçalho

	// 1. Abre o arquivo CSV da fatura fechada (Export de Origem Console GCP > Table Cost)
	fRead, err := os.Open(inputFilePath)
	if err != nil {
		log.Fatalf("Erro ao abrir o arquivo %s para leitura: %v", inputFilePath, err)
	}
	defer fRead.Close()

	// 2. Cria um leitor e joga todos os registros do CSV para a memória
	reader := csv.NewReader(fRead)
	allRecords, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Erro ao ler todos os registros do CSV: %v", err)
	}

	if len(allRecords) == 0 {
		log.Fatalf("O arquivo CSV está vazio.")
	}

	// 3. Processa o cabeçalho e encontra o índice da coluna "Código SKU"
	header := allRecords[0]
	skuColumnIndex := -1

	for i, colName := range header {
		trimmedColName := strings.TrimSpace(colName)
		if trimmedColName == targetColumn {
			skuColumnIndex = i
			break
		}
	}

	if skuColumnIndex == -1 {
		log.Fatalf("Erro: Coluna \"%s\" não encontrada no arquivo CSV.", targetColumn)
	}

	// Adiciona os novos cabeçalhos ao final da primeira linha do CSV
	allRecords[0] = append(allRecords[0], newColumnPriceDolar)
	allRecords[0] = append(allRecords[0], newColumnCustoUSD)
	allRecords[0] = append(allRecords[0], newColumnPriceUnit)
	allRecords[0] = append(allRecords[0], newColumnPriceModel)
	allRecords[0] = append(allRecords[0], newColumnPriceLocal)

	fmt.Printf("--- Processando SKUs e Inserindo Preços no CSV ---\n")

	// 4. Itera sobre os registros de dados (a partir do índice 1)
	for i := 1; i < len(allRecords); i++ {
		record := allRecords[i]
		// [i] Posição na coluna
		// [08] Código SKU
		// [14] Quantidade de uso
		// [16] Custo não arredondado R$
		// [17] Custo arredondado R$

		qtdUso_s := record[14]
		custoNaoArredondado_s := record[16]

		// Tratamento da string (qtdUso) antes da conversão para float
		qtdUso_s = strings.ReplaceAll(qtdUso_s, ".", "")
		qtdUso_s = strings.ReplaceAll(qtdUso_s, ",", ".")
		qtdUso, _ := strconv.ParseFloat(qtdUso_s, 64)

		// Tratamento da string (dolar) antes da conversão para float
		inputDolar = strings.ReplaceAll(inputDolar, ",", ".")
		dolar, _ := strconv.ParseFloat(inputDolar, 64)

		// Tratamento da string (custo não arredondado) antes da conversão para float
		custoNaoArredondado_s = strings.ReplaceAll(custoNaoArredondado_s, ",", ".")
		custoNaoArredondado, _ := strconv.ParseFloat(custoNaoArredondado_s, 64)

		// Conversão do custo de Reais para Dólar
		totalCustoUSD_f := custoNaoArredondado / dolar

		// Encontra o Custo Unitário de cada item em Dólar
		totalCustoUnit_f := totalCustoUSD_f / qtdUso

		// Valida se os valores encontrados são válidos (+Inf || IsNaN)
		if math.IsInf(totalCustoUSD_f, 1) || math.IsNaN(totalCustoUSD_f) {
			totalCustoUSD_f = 0.0
		}
		if math.IsInf(totalCustoUnit_f, 1) || math.IsNaN(totalCustoUnit_f) {
			totalCustoUnit_f = 0.0
		}

		// Reverte Custo em Dólar e Unitário para string novamente
		totalCustoUSD_s := fmt.Sprintf("%f", totalCustoUSD_f)
		totalCustoLocal := fmt.Sprintf("%f", totalCustoUnit_f)

		// Adiciona dados nas novas colunas do arquivo CSV (em memória)
		if skuColumnIndex < len(record) {
			skuValue := strings.TrimSpace(record[skuColumnIndex])

			// Se o valor do SKU for válido, executa o comando curl
			if skuValue != "" {
				// Chama a função que executa o comando curl
				priceData, err := fetchPriceUsingCurlCommand(skuValue)

				var extractedPriceModel string

				// Extrai preço float (digito.digito)
				var extractedPriceUnit string
				re := regexp.MustCompile(`^\d*\.\d+`)

				if err != nil {
					// Em caso de erro, define uma mensagem de erro na coluna
					log.Printf("Falha ao processar SKU %s: %v. Usando 'ERRO DE REQUISIÇÃO'.", skuValue, err)
					extractedPriceModel = "ERRO DE REQUISIÇÃO"
				} else {
					// Extrai o preço formatado da resposta bruta
					extractedPriceModel = extractPrice(priceData)
					extractedPriceModel = strings.ReplaceAll(extractedPriceModel, "\\", "")
					extractedPriceUnit = re.FindString(extractedPriceModel)

				}

				// Adiciona o preço unitário extraído como uma nova coluna em memória
				allRecords[i] = append(record, inputDolar, totalCustoUSD_s, extractedPriceUnit, extractedPriceModel, totalCustoLocal)

				fmt.Printf("SKU: %s -> Preço: %s Modelo: %s Dólar %s TotalCost %s\n", skuValue, extractedPriceUnit, extractedPriceModel, inputDolar, totalCustoLocal)
			} else {
				// Se o SKU estiver vazio, adiciona uma coluna vazia
				allRecords[i] = append(record, "")
				fmt.Printf("Linha %d ignorada (SKU vazio).\n", i)
			}
		} else {
			// Caso a linha esteja incompleta, adiciona uma coluna vazia
			allRecords[i] = append(record, "")
			log.Printf("Aviso: Linha %d incompleta. Adicionando coluna vazia.", i)
		}
	}

	// 5. Escreve todos os dados que estão em memória no novo arquivo CSV

	// Abre o arquivo para escrita (cria/trunca o arquivo/escreve)
	newFileName := "validado_" + inputFilePath                                      // define um novo nome para não alterar o original e cria um novo arquivo
	fWrite, err := os.OpenFile(newFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644) // os.O_TRUNC limpa os dados anteriores
	if err != nil {
		log.Fatalf("Erro ao abrir o arquivo %s para escrita: %v", inputFilePath, err)
	}
	defer fWrite.Close()

	// Cria o escritor de dados CSV
	writer := csv.NewWriter(fWrite)

	// Escreve todos os registros (cabeçalho + dados) de volta ao arquivo
	err = writer.WriteAll(allRecords)
	if err != nil {
		log.Fatalf("Erro ao escrever os dados atualizados para o CSV: %v", err)
	}

	// Garante que os dados sejam gravados no disco
	writer.Flush()

	fmt.Printf("\nProcesso finalizado.\n Arquivo %s atualizado com as novas colunas\n", newFileName)
}
