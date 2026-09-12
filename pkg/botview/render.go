package botview

import (
	"fmt"
	"reflect"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ── Рендеринг текстовой секции (@template) ───────────────────────────────────

// renderTemplate возвращает HTML-текст для Telegram.
func renderTemplate(nodes []node, scope evalScope) (string, error) {
	var sb strings.Builder
	for _, n := range nodes {
		s, err := renderNode(n, scope)
		if err != nil {
			return "", err
		}
		sb.WriteString(s)
	}
	return sb.String(), nil
}

func renderNode(n node, scope evalScope) (string, error) {
	switch nd := n.(type) {
	case *textNode:
		return nd.content, nil

	case *fieldNode:
		v, err := scope.resolve(nd.path)
		if err != nil {
			return "", fmt.Errorf("поле %q: %w", nd.path, err)
		}
		return toString(v), nil

	case *litNode:
		return nd.value, nil

	case *ifNode:
		ok, err := evalCond(nd.cond, scope)
		if err != nil {
			return "", fmt.Errorf("условие %q: %w", nd.cond, err)
		}
		if ok {
			return renderTemplate(nd.then, scope)
		}
		if nd.els != nil {
			return renderTemplate(nd.els, scope)
		}
		return "", nil

	case *eachNode:
		v, err := scope.resolve(nd.items)
		if err != nil {
			return "", fmt.Errorf("#each %q: %w", nd.items, err)
		}
		v = deref(v)
		if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
			return "", fmt.Errorf("#each %q: не является слайсом (тип %v)", nd.items, v.Kind())
		}

		var sb strings.Builder
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			inner := scope.push(nd.as, elem)
			s, err := renderTemplate(nd.body, inner)
			if err != nil {
				return "", fmt.Errorf("#each[%d]: %w", i, err)
			}
			sb.WriteString(s)
		}
		return sb.String(), nil
	}

	return "", fmt.Errorf("неизвестный тип узла: %T", n)
}

// ── Рендеринг клавиатурной секции (@keyboard) ────────────────────────────────

// renderKeyboard возвращает InlineKeyboardMarkup или nil.
func renderKeyboard(nodes []node, scope evalScope) (*tgbotapi.InlineKeyboardMarkup, error) {
	if len(nodes) == 0 {
		return nil, nil
	}

	rows, err := renderKbNodes(nodes, scope)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}

	kb := tgbotapi.NewInlineKeyboardMarkup(rows...)
	return &kb, nil
}

// renderKbNodes рендерит список keyboard-узлов в слайс строк кнопок.
func renderKbNodes(nodes []node, scope evalScope) ([][]tgbotapi.InlineKeyboardButton, error) {
	var rows [][]tgbotapi.InlineKeyboardButton

	for _, n := range nodes {
		switch nd := n.(type) {
		case *rowNode:
			row, err := renderRow(nd, scope)
			if err != nil {
				return nil, err
			}
			if len(row) > 0 {
				rows = append(rows, row)
			}

		case *ifNode:
			ok, err := evalCond(nd.cond, scope)
			if err != nil {
				return nil, fmt.Errorf("kb #if %q: %w", nd.cond, err)
			}
			var branch []node
			if ok {
				branch = nd.then
			} else {
				branch = nd.els
			}
			if branch != nil {
				subRows, err := renderKbNodes(branch, scope)
				if err != nil {
					return nil, err
				}
				rows = append(rows, subRows...)
			}

		case *eachNode:
			v, err := scope.resolve(nd.items)
			if err != nil {
				return nil, fmt.Errorf("kb #each %q: %w", nd.items, err)
			}
			v = deref(v)
			if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
				return nil, fmt.Errorf("kb #each %q: не слайс", nd.items)
			}
			for i := 0; i < v.Len(); i++ {
				inner := scope.push(nd.as, v.Index(i))
				subRows, err := renderKbNodes(nd.body, inner)
				if err != nil {
					return nil, fmt.Errorf("kb #each[%d]: %w", i, err)
				}
				rows = append(rows, subRows...)
			}
		}
	}

	return rows, nil
}

// renderRow рендерит [row] в одну строку кнопок.
func renderRow(nd *rowNode, scope evalScope) ([]tgbotapi.InlineKeyboardButton, error) {
	var btns []tgbotapi.InlineKeyboardButton

	for _, child := range nd.children {
		switch ch := child.(type) {
		case *btnNode:
			cb := interpolate(ch.rawCallback, scope)
			text := interpolate(ch.rawText, scope)
			btns = append(btns, tgbotapi.NewInlineKeyboardButtonData(text, cb))

		case *urlBtnNode:
			href := interpolate(ch.rawHref, scope)
			text := interpolate(ch.rawText, scope)
			btns = append(btns, tgbotapi.NewInlineKeyboardButtonURL(text, href))

		case *ifNode:
			ok, err := evalCond(ch.cond, scope)
			if err != nil {
				return nil, fmt.Errorf("row #if: %w", err)
			}
			var branch []node
			if ok {
				branch = ch.then
			} else {
				branch = ch.els
			}
			if branch != nil {
				subBtns, err := renderRow(&rowNode{children: branch}, scope)
				if err != nil {
					return nil, err
				}
				btns = append(btns, subBtns...)
			}

		case *eachNode:
			v, err := scope.resolve(ch.items)
			if err != nil {
				return nil, fmt.Errorf("row #each %q: %w", ch.items, err)
			}
			v = deref(v)
			if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
				return nil, fmt.Errorf("row #each %q: не слайс", ch.items)
			}
			for i := 0; i < v.Len(); i++ {
				inner := scope.push(ch.as, v.Index(i))
				subBtns, err := renderRow(&rowNode{children: ch.body}, inner)
				if err != nil {
					return nil, fmt.Errorf("row #each[%d]: %w", i, err)
				}
				btns = append(btns, subBtns...)
			}
		}
	}

	return btns, nil
}

// deref разыменовывает указатели и интерфейсы.
func deref(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	return v
}
