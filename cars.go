package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// scanner global para leitura única de stdin (evita conflitos com múltiplos scanners)
var inputScanner = bufio.NewScanner(os.Stdin)

// Carro representa um carro importado
type Carro struct {
	ID           string  `json:"id"`             // ID único baseado em timestamp
	Marca        string  `json:"marca"`          // Ex: Toyota, BMW
	Modelo       string  `json:"modelo"`         // Ex: Corolla, X5
	Ano          int     `json:"ano"`            // Ano de fabricação
	Cor          string  `json:"cor"`            // Ex: Prata, Preto
	Preco        float64 `json:"preco"`          // Preço em R$
	PaisOrigem   string  `json:"pais_origem"`    // Ex: Japão, Alemanha
	DataCadastro string  `json:"data_cadastro"`  // Data de cadastro (formato YYYY-MM-DD)
}

// CadastroCarros gerencia o banco temporário em memória
type CadastroCarros struct {
	carrosMap    map[string]Carro // Map para buscas rápidas por ID (banco principal)
	carros       []Carro          // Slice para listagem ordenada
	mu           sync.RWMutex     // Mutex para thread-safety
	arquivoJSON  string           // Caminho do arquivo JSON de persistência
}

// NewCadastroCarros cria um novo banco em memória
func NewCadastroCarros(nomeArquivo string) *CadastroCarros {
	return &CadastroCarros{
		carrosMap:   make(map[string]Carro),
		carros:      make([]Carro, 0),
		arquivoJSON: nomeArquivo,
	}
}

// AdicionarCarro adiciona um novo carro ao banco em memória com validações
func (c *CadastroCarros) AdicionarCarro() {
	// usa scanner global `inputScanner`
	fmt.Println("\n--- Cadastro de Novo Carro Importado ---")

	// Função helper para ler input com erro handling
	readInput := func(prompt string) (string, error) {
		fmt.Print(prompt)
		inputScanner.Scan() // usa scanner global
		if err := inputScanner.Err(); err != nil {
			return "", fmt.Errorf("erro no input: %v", err)
		}
		return strings.TrimSpace(inputScanner.Text()), nil
	}

	marca, err := readInput("Marca: ")
	if err != nil || marca == "" {
		fmt.Printf("Erro: %v. Marca não pode ser vazia.\n", err)
		return
	}

	modelo, err := readInput("Modelo: ")
	if err != nil || modelo == "" {
		fmt.Printf("Erro: %v. Modelo não pode ser vazio.\n", err)
		return
	}

	anoStr, err := readInput("Ano: ")
	if err != nil {
		fmt.Printf("Erro: %v\n", err)
		return
	}
	ano, err := strconv.Atoi(anoStr)
	if err != nil || ano <= 0 || ano > time.Now().Year()+1 {
		fmt.Println("Erro: Ano deve ser um número positivo válido (até", time.Now().Year()+1, ").")
		return
	}

	cor, _ := readInput("Cor: ") // Cor pode ser vazia

	precoStr, err := readInput("Preço (R$): ")
	if err != nil {
		fmt.Printf("Erro: %v\n", err)
		return
	}
	preco, err := strconv.ParseFloat(precoStr, 64)
	if err != nil || preco <= 0 {
		fmt.Println("Erro: Preço deve ser um número positivo válido.")
		return
	}

	paisOrigem, err := readInput("País de Origem: ")
	if err != nil || paisOrigem == "" {
		fmt.Printf("Erro: %v. País de origem não pode ser vazio.\n", err)
		return
	}

	// Gera ID único simples (timestamp nano) e data dinâmica
	id := fmt.Sprintf("car_%d", time.Now().UnixNano())
	dataCadastro := time.Now().Format("2006-01-02")

	novoCarro := Carro{
		ID:           id,
		Marca:        marca,
		Modelo:       modelo,
		Ano:          ano,
		Cor:          cor,
		Preco:        preco,
		PaisOrigem:   paisOrigem,
		DataCadastro: dataCadastro,
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.carrosMap[id] = novoCarro
	c.carros = append(c.carros, novoCarro)
	fmt.Printf("✅ Carro '%s %s' cadastrado no banco em memória com ID: %s\n", marca, modelo, id)
	
	// Salvar no JSON após adicionar
	if err := c.SalvarJSON(); err != nil {
		fmt.Printf("⚠️  Aviso: Falha ao salvar em JSON: %v\n", err)
	}
}

// ListarCarros exibe todos os carros do banco em memória
func (c *CadastroCarros) ListarCarros() {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.carros) == 0 {
		fmt.Println("\nNenhum carro cadastrado no banco em memória ainda.")
		return
	}

	fmt.Println("\n--- Lista de Carros Importados (Banco em Memória) ---")
	for _, carro := range c.carros {
		fmt.Printf("ID: %s | Marca: %s | Modelo: %s | Ano: %d | Cor: %s | Preço: R$ %.2f | Origem: %s | Cadastrado: %s\n",
			carro.ID, carro.Marca, carro.Modelo, carro.Ano, carro.Cor, carro.Preco, carro.PaisOrigem, carro.DataCadastro)
	}
}

// BuscarCarro busca um carro por ID no banco em memória
func (c *CadastroCarros) BuscarCarro(id string) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	carro, existe := c.carrosMap[id]
	if !existe {
		fmt.Printf("❌ Carro com ID '%s' não encontrado no banco em memória.\n", id)
		return
	}

	fmt.Printf("\n--- Carro Encontrado no Banco em Memória ---\n")
	fmt.Printf("ID: %s | Marca: %s | Modelo: %s | Ano: %d | Cor: %s | Preço: R$ %.2f | Origem: %s | Cadastrado: %s\n",
		carro.ID, carro.Marca, carro.Modelo, carro.Ano, carro.Cor, carro.Preco, carro.PaisOrigem, carro.DataCadastro)
}

// RemoverCarro remove um carro por ID do banco em memória (Deletar)
func (c *CadastroCarros) RemoverCarro(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, existe := c.carrosMap[id]
	if !existe {
		fmt.Printf("❌ Carro com ID '%s' não encontrado no banco em memória.\n", id)
		return
	}

	// Remove do map e do slice
	delete(c.carrosMap, id)
	var novosCarros []Carro
	for _, carro := range c.carros {
		if carro.ID != id {
			novosCarros = append(novosCarros, carro)
		}
	}
	c.carros = novosCarros
	fmt.Printf("✅ Carro com ID '%s' deletado (removido) do banco em memória.\n", id)
	
	// Salvar no JSON após remover
	if err := c.SalvarJSON(); err != nil {
		fmt.Printf("⚠️  Aviso: Falha ao salvar em JSON: %v\n", err)
	}
}

// AtualizarCarro atualiza um carro por ID no banco em memória
func (c *CadastroCarros) AtualizarCarro(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	carro, existe := c.carrosMap[id]
	if !existe {
		fmt.Printf("❌ Carro com ID '%s' não encontrado no banco em memória.\n", id)
		return
	}

	fmt.Printf("\n--- Atualização de Carro (ID: %s) ---\n", id)
	fmt.Printf("Dados atuais: Marca: %s, Modelo: %s, Ano: %d, Cor: %s, Preço: R$ %.2f, Origem: %s\n",
		carro.Marca, carro.Modelo, carro.Ano, carro.Cor, carro.Preco, carro.PaisOrigem)

	// Função helper para ler input com erro handling
	readInput := func(prompt string) (string, error) {
		fmt.Print(prompt)
		inputScanner.Scan()
		if err := inputScanner.Err(); err != nil {
			return "", fmt.Errorf("erro no input: %v", err)
		}
		return strings.TrimSpace(inputScanner.Text()), nil
	}

	// Atualiza campos opcionais (pergunta se quer mudar)
	updateOptional := func(current string, field string, prompt string, validator func(string) (string, error)) {
		opt, err := readInput(fmt.Sprintf("%s atual: %s. Novo %s (Enter para manter atual): ", field, current, field))
		if err != nil {
			fmt.Printf("Erro: %v. Mantendo atual.\n", err)
			return
		}
		if opt == "" {
			return // Mantém atual
		}
		newVal, err := validator(opt)
		if err != nil {
			fmt.Printf("Erro: %v. Mantendo atual.\n", err)
			return
		}
		switch field {
		case "Marca":
			carro.Marca = newVal
		case "Modelo":
			carro.Modelo = newVal
		case "Cor":
			carro.Cor = newVal
		case "País de Origem":
			carro.PaisOrigem = newVal
		}
	}

	updateOptional(carro.Marca, "Marca", "Marca", func(s string) (string, error) {
		if s == "" {
			return "", fmt.Errorf("marca não pode ser vazia")
		}
		return s, nil
	})

	updateOptional(carro.Modelo, "Modelo", "Modelo", func(s string) (string, error) {
		if s == "" {
			return "", fmt.Errorf("modelo não pode ser vazio")
		}
		return s, nil
	})

	// Ano
	anoStr, err := readInput(fmt.Sprintf("Ano atual: %d. Novo ano (Enter para manter): ", carro.Ano))
	if err == nil && anoStr != "" {
		ano, err := strconv.Atoi(anoStr)
		if err == nil && ano > 0 && ano <= time.Now().Year()+1 {
			carro.Ano = ano
		} else {
			fmt.Println("Ano inválido. Mantendo atual.")
		}
	}

	// Cor (opcional)
	updateOptional(carro.Cor, "Cor", "Cor", func(s string) (string, error) { return s, nil }) // Cor pode ser vazia

	// Preço
	precoStr, err := readInput(fmt.Sprintf("Preço atual: R$ %.2f. Novo preço (Enter para manter): ", carro.Preco))
	if err == nil && precoStr != "" {
		preco, err := strconv.ParseFloat(precoStr, 64)
		if err == nil && preco > 0 {
			carro.Preco = preco
		} else {
			fmt.Println("Preço inválido. Mantendo atual.")
		}
	}

	// País de Origem
	updateOptional(carro.PaisOrigem, "País de Origem", "País de Origem", func(s string) (string, error) {
		if s == "" {
			return "", fmt.Errorf("país de origem não pode ser vazio")
		}
		return s, nil
	})

	// Atualiza no map e no slice
	c.carrosMap[id] = carro
	var novosCarros []Carro
	for _, oldCarro := range c.carros {
		if oldCarro.ID == id {
			novosCarros = append(novosCarros, carro)
		} else {
			novosCarros = append(novosCarros, oldCarro)
		}
	}
	c.carros = novosCarros

	fmt.Printf("✅ Carro com ID '%s' atualizado no banco em memória.\n", id)
	
	// Salvar no JSON após atualizar
	if err := c.SalvarJSON(); err != nil {
		fmt.Printf("⚠️  Aviso: Falha ao salvar em JSON: %v\n", err)
	}
}

// SalvarJSON salva os carros em arquivo JSON
func (c *CadastroCarros) SalvarJSON() error {
	data, err := json.MarshalIndent(c.carros, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao serializar para JSON: %v", err)
	}

	err = os.WriteFile(c.arquivoJSON, data, 0644)
	if err != nil {
		return fmt.Errorf("erro ao escrever arquivo JSON: %v", err)
	}

	return nil
}

// CarregarJSON carrega os carros do arquivo JSON
func (c *CadastroCarros) CarregarJSON() error {
	data, err := os.ReadFile(c.arquivoJSON)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("erro ao ler arquivo JSON: %v", err)
	}

	var carros []Carro
	err = json.Unmarshal(data, &carros)
	if err != nil {
		return fmt.Errorf("erro ao desserializar JSON: %v", err)
	}

	// Reconstrói o map e o slice
	c.carros = carros
	c.carrosMap = make(map[string]Carro)
	for _, carro := range carros {
		c.carrosMap[carro.ID] = carro
	}

	return nil
}

// Menu principal interativo
func main() {
	cadastro := NewCadastroCarros("carros.json")
	
	// Carregar dados persistidos
	if err := cadastro.CarregarJSON(); err != nil {
		fmt.Printf("⚠️  Aviso ao carregar dados: %v\n", err)
	} else {
		cadastro.mu.RLock()
		if len(cadastro.carros) > 0 {
			fmt.Printf("✅ %d carro(s) carregado(s) do arquivo JSON.\n", len(cadastro.carros))
		}
		cadastro.mu.RUnlock()
	}

	fmt.Println("🚗 Bem-vindo ao Sistema de Cadastro de Carros Importados!")
	fmt.Println("Digite 'add' para adicionar, 'list' para listar, 'find <ID>' para buscar, 'remove <ID>' para deletar, 'update <ID>' para atualizar, ou 'exit' para sair.")

	// usa scanner global `inputScanner`
	for {
		fmt.Print("\n> ")
		if !inputScanner.Scan() {
			if err := inputScanner.Err(); err != nil {
				fmt.Printf("Erro de leitura: %v. Saindo...\n", err)
			}
			break
		}
		input := strings.ToLower(strings.TrimSpace(inputScanner.Text()))

		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}
		cmd := parts[0]

		switch cmd {
		case "add":
			cadastro.AdicionarCarro()
		case "list":
			cadastro.ListarCarros()
		case "find":
			if len(parts) < 2 {
				fmt.Println("Uso: find <ID>")
				continue
			}
			cadastro.BuscarCarro(parts[1])
		case "remove":
			if len(parts) < 2 {
				fmt.Println("Uso: remove <ID>")
				continue
			}
			cadastro.RemoverCarro(parts[1])
		case "update":
			if len(parts) < 2 {
				fmt.Println("Uso: update <ID>")
				continue
			}
			cadastro.AtualizarCarro(parts[1])
		case "exit":
			fmt.Println("Saindo do sistema. Dados do banco em memória perdidos (temporário). Até logo!")
			return
		default:
			fmt.Println("Comando inválido. Tente 'add', 'list', 'find <ID>', 'remove <ID>', 'update <ID>' ou 'exit'.")
		}
	}
}