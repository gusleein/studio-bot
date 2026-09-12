package botview

import (
	"fmt"
	"strings"
	"unicode"
)

// ── Секционный разбор файла ─────────────────────────────────────────────────

// parseFile разбирает .botview файл на документ.
func parseFile(src string) (*document, error) {
	propsType, tmplSrc, kbSrc, err := splitSections(src)
	if err != nil {
		return nil, err
	}

	tmplNodes, err := parseTemplateSection(tmplSrc)
	if err != nil {
		return nil, fmt.Errorf("@template: %w", err)
	}

	kbNodes, err := parseKeyboardSection(kbSrc)
	if err != nil {
		return nil, fmt.Errorf("@keyboard: %w", err)
	}

	return &document{
		propsType: propsType,
		template:  tmplNodes,
		keyboard:  kbNodes,
	}, nil
}

// splitSections делит файл на части по маркерам @props, @template, @keyboard.
func splitSections(src string) (propsType, tmplSrc, kbSrc string, err error) {
	tmplIdx := strings.Index(src, "@template")
	if tmplIdx < 0 {
		return "", "", "", fmt.Errorf("не найдена секция @template")
	}

	header := src[:tmplIdx]
	if pi := strings.Index(header, "@props"); pi >= 0 {
		line := header[pi+6:]
		if nl := strings.IndexByte(line, '\n'); nl >= 0 {
			line = line[:nl]
		}
		propsType = strings.TrimSpace(line)
	}

	afterTemplate := src[tmplIdx+len("@template"):]
	// Пропускаем newline сразу после маркера
	if strings.HasPrefix(afterTemplate, "\n") {
		afterTemplate = afterTemplate[1:]
	}

	if kbIdx := strings.Index(afterTemplate, "@keyboard"); kbIdx >= 0 {
		tmplSrc = afterTemplate[:kbIdx]
		kbSrc = afterTemplate[kbIdx+len("@keyboard"):]
		if strings.HasPrefix(kbSrc, "\n") {
			kbSrc = kbSrc[1:]
		}
	} else {
		tmplSrc = afterTemplate
	}

	return propsType, tmplSrc, kbSrc, nil
}

// ── Template parser ─────────────────────────────────────────────────────────

const (
	stopNone   = ""
	stopElse   = "else"
	stopEndIf  = "endif"
	stopEndEach = "endeach"
)

type templateParser struct {
	src      []rune
	pos      int
	lastStop string // что остановило последний parseNodes: "", "else", "endif", "endeach"
}

func parseTemplateSection(src string) ([]node, error) {
	p := &templateParser{src: []rune(src)}
	nodes, err := p.parseNodes()
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

// parseNodes читает узлы до конца или до управляющего закрывающего токена.
// Управляющий токен ({/if}, {/each}, {#else}) сохраняется в p.lastStop.
func (p *templateParser) parseNodes() ([]node, error) {
	var nodes []node

	for p.pos < len(p.src) {
		if p.src[p.pos] != '{' {
			// Обычный текст — читаем до следующего {
			start := p.pos
			for p.pos < len(p.src) && p.src[p.pos] != '{' {
				p.pos++
			}
			if p.pos > start {
				nodes = append(nodes, &textNode{content: string(p.src[start:p.pos])})
			}
			continue
		}

		nd, stop, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if stop != stopNone {
			p.lastStop = stop
			return nodes, nil
		}
		if nd != nil {
			nodes = append(nodes, nd)
		}
	}

	p.lastStop = stopNone
	return nodes, nil
}

// parseExpr разбирает {...} выражение, начиная с позиции '{'.
// Возвращает (узел, stopReason, ошибка).
func (p *templateParser) parseExpr() (node, string, error) {
	p.pos++ // consume '{'

	start := p.pos
	for p.pos < len(p.src) && p.src[p.pos] != '}' {
		p.pos++
	}
	inner := strings.TrimSpace(string(p.src[start:p.pos]))
	if p.pos < len(p.src) {
		p.pos++ // consume '}'
	}

	// Управляющие токены потребляют trailing newline
	isCtrl := strings.HasPrefix(inner, "#") || strings.HasPrefix(inner, "/")
	if isCtrl {
		p.skipNewline()
	}

	switch {
	case inner == "/if":
		return nil, stopEndIf, nil

	case inner == "/each":
		return nil, stopEndEach, nil

	case inner == "#else":
		return nil, stopElse, nil

	case strings.HasPrefix(inner, "#if "):
		cond := strings.TrimSpace(inner[4:])
		then, err := p.parseNodes()
		if err != nil {
			return nil, stopNone, fmt.Errorf("#if: %w", err)
		}

		var els []node
		if p.lastStop == stopElse {
			els, err = p.parseNodes()
			if err != nil {
				return nil, stopNone, fmt.Errorf("#else: %w", err)
			}
		}
		// Теперь p.lastStop должен быть stopEndIf

		return &ifNode{cond: cond, then: then, els: els}, stopNone, nil

	case strings.HasPrefix(inner, "#each "):
		rest := strings.TrimSpace(inner[6:])
		parts := strings.SplitN(rest, " as ", 2)
		if len(parts) != 2 {
			return nil, stopNone, fmt.Errorf("#each: ожидается 'Items as varName', получено: %q", rest)
		}
		items := strings.TrimSpace(parts[0])
		as := strings.TrimSpace(parts[1])

		body, err := p.parseNodes()
		if err != nil {
			return nil, stopNone, fmt.Errorf("#each: %w", err)
		}

		return &eachNode{items: items, as: as, body: body}, stopNone, nil

	case strings.HasPrefix(inner, `"`):
		val, err := parseGoStringLiteral(inner)
		if err != nil {
			return nil, stopNone, fmt.Errorf("литерал: %w", err)
		}
		return &litNode{value: val}, stopNone, nil

	default:
		if isValidPath(inner) {
			return &fieldNode{path: inner}, stopNone, nil
		}
		return nil, stopNone, fmt.Errorf("неизвестное выражение: %q", inner)
	}
}

func (p *templateParser) skipNewline() {
	if p.pos < len(p.src) && p.src[p.pos] == '\n' {
		p.pos++
	}
}

// ── Keyboard parser ─────────────────────────────────────────────────────────

type kbParser struct {
	src      []rune
	pos      int
	lastStop string
}

func parseKeyboardSection(src string) ([]node, error) {
	if strings.TrimSpace(src) == "" {
		return nil, nil
	}
	p := &kbParser{src: []rune(src)}
	return p.parseNodes()
}

func (p *kbParser) parseNodes() ([]node, error) {
	var nodes []node

	for p.pos < len(p.src) {
		ch := p.src[p.pos]

		// Пропускаем пробельные символы
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			p.pos++
			continue
		}

		if ch == '[' {
			nd, stop, err := p.parseKbTag()
			if err != nil {
				return nil, err
			}
			if stop != stopNone {
				p.lastStop = stop
				return nodes, nil
			}
			if nd != nil {
				nodes = append(nodes, nd)
			}
			continue
		}

		if ch == '{' {
			nd, stop, err := p.parseCtrl()
			if err != nil {
				return nil, err
			}
			if stop != stopNone {
				p.lastStop = stop
				return nodes, nil
			}
			if nd != nil {
				nodes = append(nodes, nd)
			}
			continue
		}

		p.pos++ // пропускаем неизвестные символы
	}

	p.lastStop = stopNone
	return nodes, nil
}

// parseKbTag разбирает [tag attrs] ... [/tag].
func (p *kbParser) parseKbTag() (node, string, error) {
	p.pos++ // consume '['

	// Закрывающий тег [/...]
	if p.pos < len(p.src) && p.src[p.pos] == '/' {
		for p.pos < len(p.src) && p.src[p.pos] != ']' {
			p.pos++
		}
		p.pos++ // consume ']'
		// Закрывающий [/row], [/btn], [/url] — стоп для текущего блока
		return nil, stopEndIf, nil // используем stopEndIf как универсальный "закрывающий"
	}

	start := p.pos
	for p.pos < len(p.src) && p.src[p.pos] != ']' {
		p.pos++
	}
	tagContent := strings.TrimSpace(string(p.src[start:p.pos]))
	p.pos++ // consume ']'

	tagName, attrs := parseTagContent(tagContent)

	switch tagName {
	case "row":
		children, err := p.parseNodes()
		if err != nil {
			return nil, stopNone, fmt.Errorf("[row]: %w", err)
		}
		return &rowNode{children: children}, stopNone, nil

	case "btn":
		cb := attrs["callback"]
		text := p.readTextUntilClose("btn")
		return &btnNode{
			rawCallback: cb,
			rawText:     strings.TrimSpace(text),
		}, stopNone, nil

	case "url":
		href := attrs["href"]
		text := p.readTextUntilClose("url")
		return &urlBtnNode{
			rawHref: href,
			rawText: strings.TrimSpace(text),
		}, stopNone, nil
	}

	return nil, stopNone, fmt.Errorf("неизвестный тег клавиатуры: %q", tagName)
}

// readTextUntilClose читает текст до [/tagName].
func (p *kbParser) readTextUntilClose(tag string) string {
	closeTag := []rune("[/" + tag + "]")
	start := p.pos

	for p.pos < len(p.src) {
		if p.pos+len(closeTag) <= len(p.src) {
			match := true
			for i, r := range closeTag {
				if p.src[p.pos+i] != r {
					match = false
					break
				}
			}
			if match {
				text := string(p.src[start:p.pos])
				p.pos += len(closeTag)
				return text
			}
		}
		p.pos++
	}
	return string(p.src[start:])
}

// parseCtrl разбирает {#if}, {/if}, {#each}, {/each}, {#else} в keyboard секции.
func (p *kbParser) parseCtrl() (node, string, error) {
	p.pos++ // consume '{'

	start := p.pos
	for p.pos < len(p.src) && p.src[p.pos] != '}' {
		p.pos++
	}
	inner := strings.TrimSpace(string(p.src[start:p.pos]))
	if p.pos < len(p.src) {
		p.pos++ // consume '}'
	}

	isCtrl := strings.HasPrefix(inner, "#") || strings.HasPrefix(inner, "/")
	if isCtrl {
		p.skipNewline()
	}

	switch {
	case inner == "/if":
		return nil, stopEndIf, nil

	case inner == "/each":
		return nil, stopEndEach, nil

	case inner == "#else":
		return nil, stopElse, nil

	case strings.HasPrefix(inner, "#if "):
		cond := strings.TrimSpace(inner[4:])
		then, err := p.parseNodes()
		if err != nil {
			return nil, stopNone, fmt.Errorf("#if: %w", err)
		}

		var els []node
		if p.lastStop == stopElse {
			els, err = p.parseNodes()
			if err != nil {
				return nil, stopNone, fmt.Errorf("#else: %w", err)
			}
		}

		return &ifNode{cond: cond, then: then, els: els}, stopNone, nil

	case strings.HasPrefix(inner, "#each "):
		rest := strings.TrimSpace(inner[6:])
		parts := strings.SplitN(rest, " as ", 2)
		if len(parts) != 2 {
			return nil, stopNone, fmt.Errorf("#each: ожидается 'Items as varName', получено: %q", rest)
		}

		body, err := p.parseNodes()
		if err != nil {
			return nil, stopNone, fmt.Errorf("#each: %w", err)
		}

		return &eachNode{
			items: strings.TrimSpace(parts[0]),
			as:    strings.TrimSpace(parts[1]),
			body:  body,
		}, stopNone, nil
	}

	return nil, stopNone, nil
}

func (p *kbParser) skipNewline() {
	if p.pos < len(p.src) && p.src[p.pos] == '\n' {
		p.pos++
	}
}

// ── Вспомогательные функции ─────────────────────────────────────────────────

// parseTagContent разбирает "btn callback=\"val\"" → ("btn", {"callback":"val"}).
func parseTagContent(s string) (name string, attrs map[string]string) {
	attrs = make(map[string]string)
	s = strings.TrimSpace(s)

	spaceIdx := strings.IndexByte(s, ' ')
	if spaceIdx < 0 {
		return s, attrs
	}
	name = s[:spaceIdx]
	rest := strings.TrimSpace(s[spaceIdx+1:])

	for len(rest) > 0 {
		eqIdx := strings.IndexByte(rest, '=')
		if eqIdx < 0 {
			break
		}
		key := strings.TrimSpace(rest[:eqIdx])
		rest = rest[eqIdx+1:]

		var val string
		if strings.HasPrefix(rest, `"`) {
			rest = rest[1:]
			endIdx := strings.IndexByte(rest, '"')
			if endIdx < 0 {
				val = rest
				rest = ""
			} else {
				val = rest[:endIdx]
				rest = strings.TrimSpace(rest[endIdx+1:])
			}
		} else {
			spIdx := strings.IndexByte(rest, ' ')
			if spIdx < 0 {
				val = rest
				rest = ""
			} else {
				val = rest[:spIdx]
				rest = strings.TrimSpace(rest[spIdx+1:])
			}
		}
		attrs[key] = val
	}

	return name, attrs
}

// isValidPath проверяет, что строка — допустимый путь к полю.
func isValidPath(s string) bool {
	if s == "" {
		return false
	}
	for _, part := range strings.Split(s, ".") {
		if part == "" {
			return false
		}
		for i, r := range part {
			if i == 0 && !unicode.IsLetter(r) && r != '_' {
				return false
			}
			if i > 0 && !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
				return false
			}
		}
	}
	return true
}

// parseGoStringLiteral разбирает Go-строковый литерал: "text\nmore".
func parseGoStringLiteral(s string) (string, error) {
	s = strings.TrimSpace(s)
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return "", fmt.Errorf("не является строковым литералом: %q", s)
	}
	inner := s[1 : len(s)-1]

	var result strings.Builder
	i := 0
	for i < len(inner) {
		if inner[i] == '\\' && i+1 < len(inner) {
			switch inner[i+1] {
			case 'n':
				result.WriteByte('\n')
			case 't':
				result.WriteByte('\t')
			case 'r':
				result.WriteByte('\r')
			case '"':
				result.WriteByte('"')
			case '\\':
				result.WriteByte('\\')
			default:
				result.WriteByte('\\')
				result.WriteByte(inner[i+1])
			}
			i += 2
		} else {
			result.WriteByte(inner[i])
			i++
		}
	}
	return result.String(), nil
}
