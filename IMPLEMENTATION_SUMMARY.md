# Implementação do Collection Picker Interativo

## ✅ Funcionalidades Implementadas

### 1. Seletor Interativo de Collections (Bubbletea)
- Interface TUI completa com busca filtrada em tempo real
- Navegação com setas (↑↓) ou teclas vim (j/k)
- Busca case-insensitive que filtra conforme o usuário digita
- Exibição do número de notas em cada collection
- Suporte para cancelar com ESC

### 2. Criação de Novas Collections
- Quando o usuário digita um nome que não existe, aparece opção: `✨ Criar nova: "nome"`
- Normalização automática do nome (remoção de acentos, conversão para slug)
- Criação do diretório ao confirmar com Enter

### 3. Comando Novo Simplificado
- **Antes**: `margi new [collection] [title]` (obrigatório informar collection)
- **Agora**: 
  - `margi new [title]` → Abre o seletor interativo
  - `margi new [collection] [title]` → Mantém comportamento tradicional

### 4. Comando para Listar Collections
- Novo comando: `margi collections`
- Mostra todas as collections disponíveis com contagem de notas

## 📁 Arquivos Criados/Modificados

### Novos Arquivos
1. `internal/collection/service.go` - Serviço para gerenciar collections
2. `internal/collection/service_test.go` - Testes do serviço
3. `internal/ui/picker.go` - Interface Bubbletea do seletor
4. `internal/ui/picker_test.go` - Testes da interface

### Arquivos Modificados
1. `cmd/margi/main.go` - Integração do picker e novo comando collections
2. `internal/slug/service.go` - Exportação da função MakeSlug
3. `go.mod` / `go.sum` - Adicionadas dependências bubbletea e lipgloss

## 🎯 Como Usar

### Criar nova nota com seletor interativo
```bash
$ margi new "Título da minha nota"
```

Abrirá uma interface assim:

```
Selecionar Collection

Buscar: █

  ▸ blog (12 notas)
    drafts (5 notas)
    essays (3 notas)
    journal (45 notas)

[↑↓/jk] navegar • [Enter] selecionar • [Esc] cancelar
```

### Filtrar collections
Digite para filtrar em tempo real:

```
Selecionar Collection

Buscar: jo█

  ▸ journal (45 notas)

[↑↓/jk] navegar • [Enter] selecionar • [Esc] cancelar
```

### Criar nova collection
Digite um nome que não existe:

```
Selecionar Collection

Buscar: ideas█

  ▸ ✨ Criar nova: "ideas"

[Enter] criar [Esc] cancelar
```

### Listar todas as collections
```bash
$ margi collections
Collections disponíveis:

  • blog (12 notas)
  • drafts (5 notas)
  • essays (3 notas)
  • journal (45 notas)
```

### Método tradicional (ainda funciona)
```bash
$ margi new blog "Meu post sobre Go"
```

## 🧪 Testes

Todos os testes passam:

```bash
$ go test ./...
✓ TestListCollections
✓ TestCreateCollection
✓ TestCollectionExists
✓ TestGetCollectionStats
✓ TestNewPickerModel
✓ TestFilteredItems
✓ TestCreateNewOption
✓ TestNavigation
✓ TestView
✓ TestCancellation
```

## 🎨 Detalhes de UX

### Cores e Estilo
- Título em azul claro (cor 39)
- Input do usuário em rosa/magenta (cor 205)
- Item selecionado em roxo claro (cor 170)
- Opção de criar nova em verde (cor 86)
- Ajuda/instruções em cinza claro (cor 241)
- Erros em vermelho (cor 196)

### Navegação
- **↑ / k**: Mover para cima
- **↓ / j**: Mover para baixo
- **Enter**: Selecionar collection ou criar nova
- **Esc / Ctrl+C**: Cancelar operação
- **Qualquer letra**: Adicionar ao filtro
- **Backspace**: Remover do filtro

### Comportamento de Filtro
- Filtro em tempo real enquanto digita
- Case-insensitive (maiúsculas/minúsculas ignoradas)
- Match em qualquer parte do nome da collection
- Se não houver match exato, mostra opção de criar nova
- Cursor volta ao topo ao filtrar

### Scroll para Listas Longas
- Mostra até 10 itens por vez
- Scroll automático quando cursor se aproxima das bordas
- Indicador de quantos itens estão sendo mostrados

## 🏗️ Arquitetura

```
┌─────────────────┐
│   cmd/margi     │  ← Interface CLI
│   main.go       │
└────────┬────────┘
         │
         ├──────────┐
         │          │
    ┌────▼───┐  ┌──▼──────────┐
    │   ui   │  │ collection  │
    │picker.go│  │ service.go  │
    └────────┘  └──────┬──────┘
                       │
                  ┌────▼────┐
                  │ storage │
                  │  fs.go  │
                  └─────────┘
```

## ✨ Funcionalidades Extras Implementadas

Além do plano original, foram adicionados:

1. **Testes completos** - Cobertura de teste para collection service e UI picker
2. **Pluralização inteligente** - "1 nota" vs "2 notas"
3. **Scroll para listas longas** - Interface adaptável para muitas collections
4. **Validação de nomes** - Normalização automática usando slug
5. **Mensagens em português** - Interface totalmente localizada
6. **Comando collections** - Visualização rápida de todas as collections

## 🚀 Build e Instalação

```bash
# Build local
make build
./dist/margi

# Instalar no sistema
make install
margi
```
