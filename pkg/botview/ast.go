package botview

// node — базовый интерфейс узла AST.
type node interface{ isNode() }

// ── Template nodes ──────────────────────────────────────────────────────────

// textNode — литеральный текст (включая HTML теги <b>, </b> и т.д.)
type textNode struct{ content string }

// fieldNode — подстановка поля: {FieldName} или {nested.Field}
type fieldNode struct{ path string }

// litNode — строковый литерал: {"\n"} → \n
type litNode struct{ value string }

// ifNode — условный блок: {#if Cond}...{#else}...{/if}
type ifNode struct {
	cond string // сырое выражение условия
	then []node
	els  []node // может быть nil
}

// eachNode — цикл: {#each Items as item}...{/each}
type eachNode struct {
	items string // путь к полю-слайсу
	as    string // имя переменной цикла
	body  []node
}

// ── Keyboard nodes ──────────────────────────────────────────────────────────

// rowNode — строка кнопок: [row]...[/row]
type rowNode struct{ children []node }

// btnNode — callback-кнопка: [btn callback="sched:{Offset}"] Текст [/btn]
type btnNode struct {
	rawCallback string // шаблон callback_data (может содержать {Field})
	rawText     string // текст кнопки
}

// urlBtnNode — URL-кнопка: [url href="{Link}"] Текст [/url]
type urlBtnNode struct {
	rawHref string // шаблон URL
	rawText string // текст кнопки
}

func (n *textNode)   isNode() {}
func (n *fieldNode)  isNode() {}
func (n *litNode)    isNode() {}
func (n *ifNode)     isNode() {}
func (n *eachNode)   isNode() {}
func (n *rowNode)    isNode() {}
func (n *btnNode)    isNode() {}
func (n *urlBtnNode) isNode() {}

// document — разобранный .botview файл.
type document struct {
	propsType string // имя Go-структуры (из @props)
	template  []node // узлы секции @template
	keyboard  []node // узлы секции @keyboard (rowNode, ifNode, eachNode)
}
