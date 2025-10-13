package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

//to run: go run ./pkg/converter/converter.go convert /home/lamao/Downloads/th_pt_BR.dat synonyms.txt

// ThesaurusConverter converte arquivo .dat para formato Elasticsearch
type ThesaurusConverter struct {
	minSynonyms int // Mínimo de sinônimos para incluir
	maxWords    int // Máximo de palavras por linha
}

// ConvertDatToSynonyms lê th_pt_BR.dat e gera synonyms.txt
func (tc *ThesaurusConverter) ConvertDatToSynonyms(inputFile, outputFile string) error {
	input, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo: %v", err)
	}
	defer func() {
		_ = input.Close()
	}()

	output, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo de saída: %v", err)
	}
	defer func() {
		_ = output.Close()
	}()

	scanner := bufio.NewScanner(input)
	writer := bufio.NewWriter(output)
	defer func() {
		_ = writer.Flush()
	}()

	var currentWord string
	var synonyms []string
	lineCount := 0
	entryCount := 0

	// Regex para limpar caracteres especiais
	cleanRegex := regexp.MustCompile(`[^\p{L}\s]`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineCount++

		// Pula linhas vazias ou encoding
		if line == "" || strings.HasPrefix(line, "ISO") || strings.HasPrefix(line, "UTF") {
			continue
		}

		// Formato do arquivo .dat:
		// palavra|número_de_sinônimos
		// (sinônimo1|tipo)
		// (sinônimo2|tipo)
		// ...

		if strings.Contains(line, "|") && !strings.HasPrefix(line, "(") {
			// Processa entrada anterior se existir
			if currentWord != "" && len(synonyms) >= tc.minSynonyms {
				tc.writeSynonymLine(writer, currentWord, synonyms, cleanRegex)
				entryCount++
			}

			// Nova entrada
			parts := strings.Split(line, "|")
			currentWord = strings.ToLower(strings.TrimSpace(parts[0]))
			synonyms = []string{}

		} else if strings.HasPrefix(line, "(") {
			// Linha de sinônimo: (palavra|tipo)
			line = strings.Trim(line, "()")
			parts := strings.Split(line, "|")
			if len(parts) > 0 {
				synonym := strings.ToLower(strings.TrimSpace(parts[0]))
				// Remove marcadores como "sinônimo", "antônimo", etc
				synonym = strings.ReplaceAll(synonym, "sinônimo", "")
				synonym = strings.ReplaceAll(synonym, "antônimo", "")
				synonym = strings.ReplaceAll(synonym, "termo", "")
				synonym = strings.TrimSpace(synonym)

				if synonym != "" && synonym != currentWord && len(synonym) > 2 {
					synonyms = append(synonyms, synonym)
				}
			}
		}

		// Limita progresso no console
		if lineCount%10000 == 0 {
			fmt.Printf("Processadas %d linhas, %d entradas geradas...\n", lineCount, entryCount)
		}
	}

	// Processa última entrada
	if currentWord != "" && len(synonyms) >= tc.minSynonyms {
		tc.writeSynonymLine(writer, currentWord, synonyms, cleanRegex)
		entryCount++
	}

	fmt.Printf("\n✅ Conversão concluída!\n")
	fmt.Printf("📊 Total de linhas processadas: %d\n", lineCount)
	fmt.Printf("📝 Total de entradas geradas: %d\n", entryCount)

	return scanner.Err()
}

// writeSynonymLine escreve uma linha de sinônimos no formato Elasticsearch
func (tc *ThesaurusConverter) writeSynonymLine(writer *bufio.Writer, word string, synonyms []string, cleanRegex *regexp.Regexp) {
	// Remove duplicatas
	uniqueSyns := tc.removeDuplicates(synonyms)

	// Limita quantidade de palavras
	if len(uniqueSyns) > tc.maxWords {
		uniqueSyns = uniqueSyns[:tc.maxWords]
	}

	// Adiciona palavra original
	allWords := append([]string{word}, uniqueSyns...)

	// Limpa caracteres especiais e marcadores
	cleaned := make([]string, 0, len(allWords))
	for _, w := range allWords {
		// Remove marcadores textuais
		clean := strings.ReplaceAll(w, "sinônimo", "")
		clean = strings.ReplaceAll(clean, "antônimo", "")
		clean = strings.ReplaceAll(clean, "termo", "")
		clean = strings.ReplaceAll(clean, "relacionado", "")

		// Remove caracteres especiais
		clean = cleanRegex.ReplaceAllString(clean, "")
		clean = strings.TrimSpace(clean)

		if clean != "" && len(clean) > 2 { // Ignora palavras muito curtas
			cleaned = append(cleaned, clean)
		}
	}

	if len(cleaned) >= 2 {
		line := strings.Join(cleaned, ", ")
		_, _ = writer.WriteString(line + "\n")
	}
}

// removeDuplicates remove sinônimos duplicados
func (tc *ThesaurusConverter) removeDuplicates(synonyms []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, syn := range synonyms {
		if !seen[syn] {
			seen[syn] = true
			result = append(result, syn)
		}
	}

	return result
}

// FilterByDomain filtra sinônimos para domínio específico (suporte técnico)
func (tc *ThesaurusConverter) FilterByDomain(inputFile, outputFile string, keywords []string) error {
	input, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer func() {
		_ = input.Close()
	}()

	output, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer func() {
		_ = output.Close()
	}()

	scanner := bufio.NewScanner(input)
	writer := bufio.NewWriter(output)
	defer func() {
		_ = writer.Flush()
	}()

	keywordMap := make(map[string]bool)
	for _, kw := range keywords {
		keywordMap[strings.ToLower(kw)] = true
	}

	filteredCount := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Verifica se a linha contém alguma palavra-chave
		words := strings.Split(line, ",")
		for i, word := range words {
			words[i] = strings.TrimSpace(word)
		}

		shouldInclude := false
		for _, word := range words {
			if keywordMap[strings.ToLower(word)] {
				shouldInclude = true
				break
			}
		}

		if shouldInclude {
			_, _ = writer.WriteString(line + "\n")
			filteredCount++
		}
	}

	fmt.Printf("✅ Filtro aplicado! %d linhas mantidas.\n", filteredCount)
	return scanner.Err()
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("📖 Uso:")
		fmt.Println("  Converter: go run converter.go convert th_pt_BR.dat synonyms.txt")
		fmt.Println("  Filtrar:   go run converter.go filter synonyms.txt synonyms_suporte.txt")
		return
	}

	command := os.Args[1]
	converter := &ThesaurusConverter{
		minSynonyms: 2,  // Mínimo 2 sinônimos por palavra
		maxWords:    15, // Máximo 15 palavras por linha (performance)
	}

	switch command {
	case "convert":
		if len(os.Args) < 4 {
			fmt.Println("❌ Erro: especifique arquivo de entrada e saída")
			return
		}
		inputFile := os.Args[2]
		outputFile := os.Args[3]

		fmt.Printf("🔄 Convertendo %s para %s...\n", inputFile, outputFile)
		if err := converter.ConvertDatToSynonyms(inputFile, outputFile); err != nil {
			fmt.Printf("❌ Erro: %v\n", err)
			return
		}

	case "filter":
		if len(os.Args) < 4 {
			fmt.Println("❌ Erro: especifique arquivo de entrada e saída")
			return
		}
		inputFile := os.Args[2]
		outputFile := os.Args[3]

		// Palavras-chave para suporte técnico
		keywords := []string{
			"erro", "problema", "falha", "bug", "defeito",
			"login", "senha", "acesso", "entrar",
			"lento", "travado", "parado", "quebrado",
			"ajuda", "suporte", "dúvida",
			"cancelar", "reembolso", "devolução",
			"pagamento", "boleto", "cartão",
			"produto", "compra", "pedido", "entrega",
			"email", "mensagem", "notificação",
			"instalar", "configurar", "atualizar",
			"conta", "cadastro", "perfil",
			"internet", "conexão", "wifi",
			"segurança", "vírus", "proteção",
		}

		fmt.Printf("🔍 Filtrando %s para %s...\n", inputFile, outputFile)
		if err := converter.FilterByDomain(inputFile, outputFile, keywords); err != nil {
			fmt.Printf("❌ Erro: %v\n", err)
			return
		}

	default:
		fmt.Println("❌ Comando inválido. Use 'convert' ou 'filter'")
	}
}
