package botview

import (
	"strings"
	"testing"
)

// ── Тесты парсера ────────────────────────────────────────────────────────────

func TestSplitSections(t *testing.T) {
	src := `@props MyProps

@template
Hello {Name}

@keyboard
[row]
[btn callback="go"] OK [/btn]
[/row]
`
	propsType, tmpl, kb, err := splitSections(src)
	if err != nil {
		t.Fatalf("splitSections: %v", err)
	}
	if propsType != "MyProps" {
		t.Errorf("propsType = %q, хотим %q", propsType, "MyProps")
	}
	if !strings.Contains(tmpl, "Hello {Name}") {
		t.Errorf("tmpl не содержит 'Hello {Name}': %q", tmpl)
	}
	if !strings.Contains(kb, "[row]") {
		t.Errorf("kb не содержит '[row]': %q", kb)
	}
}

func TestParseFile_NoKeyboard(t *testing.T) {
	src := `@props WelcomeProps

@template
Привет, {FirstName}!
`
	doc, err := parseFile(src)
	if err != nil {
		t.Fatalf("parseFile: %v", err)
	}
	if doc.propsType != "WelcomeProps" {
		t.Errorf("propsType = %q", doc.propsType)
	}
	if len(doc.template) == 0 {
		t.Error("template пуст")
	}
	if len(doc.keyboard) != 0 {
		t.Errorf("keyboard должен быть пустым, получили %d узлов", len(doc.keyboard))
	}
}

func TestParseFile_WithIf(t *testing.T) {
	src := `@props P

@template
{#if Flag}YES{/if}
`
	doc, err := parseFile(src)
	if err != nil {
		t.Fatalf("parseFile: %v", err)
	}
	if len(doc.template) == 0 {
		t.Fatal("template пуст")
	}
	ifN, ok := doc.template[0].(*ifNode)
	if !ok {
		t.Fatalf("ожидался ifNode, получили %T", doc.template[0])
	}
	if ifN.cond != "Flag" {
		t.Errorf("cond = %q, хотим %q", ifN.cond, "Flag")
	}
}

func TestParseFile_WithEach(t *testing.T) {
	src := `@props P

@template
{#each Items as item}
{item.Name}
{/each}
`
	doc, err := parseFile(src)
	if err != nil {
		t.Fatalf("parseFile: %v", err)
	}
	// Первый узел — eachNode
	var eachN *eachNode
	for _, n := range doc.template {
		if e, ok := n.(*eachNode); ok {
			eachN = e
			break
		}
	}
	if eachN == nil {
		t.Fatal("eachNode не найден")
	}
	if eachN.items != "Items" || eachN.as != "item" {
		t.Errorf("each: items=%q as=%q", eachN.items, eachN.as)
	}
}

// ── Тесты рендерера ──────────────────────────────────────────────────────────

type welcomeTestProps struct {
	FirstName string
}

func TestRender_SimpleField(t *testing.T) {
	src := `@props welcomeTestProps

@template
Привет, {FirstName}!
`
	doc, err := parseFile(src)
	if err != nil {
		t.Fatalf("parseFile: %v", err)
	}

	scope := newScope(welcomeTestProps{FirstName: "Алекс"})
	text, err := renderTemplate(doc.template, scope)
	if err != nil {
		t.Fatalf("renderTemplate: %v", err)
	}
	if !strings.Contains(text, "Алекс") {
		t.Errorf("текст не содержит 'Алекс': %q", text)
	}
}

type schedTestProps struct {
	IsEmpty bool
	Title   string
}

func TestRender_IfFalse(t *testing.T) {
	src := `@props schedTestProps

@template
{Title}{#if IsEmpty}EMPTY{/if}
`
	doc, _ := parseFile(src)
	scope := newScope(schedTestProps{IsEmpty: false, Title: "Week"})
	text, err := renderTemplate(doc.template, scope)
	if err != nil {
		t.Fatalf("renderTemplate: %v", err)
	}
	if strings.Contains(text, "EMPTY") {
		t.Errorf("не должно содержать EMPTY: %q", text)
	}
	if !strings.Contains(text, "Week") {
		t.Errorf("должно содержать 'Week': %q", text)
	}
}

func TestRender_IfTrue(t *testing.T) {
	src := `@props schedTestProps

@template
{#if IsEmpty}EMPTY{/if}
`
	doc, _ := parseFile(src)
	scope := newScope(schedTestProps{IsEmpty: true})
	text, err := renderTemplate(doc.template, scope)
	if err != nil {
		t.Fatalf("renderTemplate: %v", err)
	}
	if !strings.Contains(text, "EMPTY") {
		t.Errorf("должно содержать EMPTY: %q", text)
	}
}

type eachTestProps struct {
	Items []eachItem
}

type eachItem struct {
	Name  string
	Value int
}

func TestRender_Each(t *testing.T) {
	src := `@props eachTestProps

@template
{#each Items as it}
{it.Name}:{it.Value}
{/each}
`
	doc, err := parseFile(src)
	if err != nil {
		t.Fatalf("parseFile: %v", err)
	}

	props := eachTestProps{
		Items: []eachItem{
			{Name: "Alpha", Value: 1},
			{Name: "Beta", Value: 2},
		},
	}

	scope := newScope(props)
	text, err := renderTemplate(doc.template, scope)
	if err != nil {
		t.Fatalf("renderTemplate: %v", err)
	}
	if !strings.Contains(text, "Alpha:1") {
		t.Errorf("нет Alpha:1 в %q", text)
	}
	if !strings.Contains(text, "Beta:2") {
		t.Errorf("нет Beta:2 в %q", text)
	}
}

func TestRender_Literal(t *testing.T) {
	src := "@props P\n\n@template\nline1{\"\n\"}line2"
	doc, err := parseFile(src)
	if err != nil {
		t.Fatalf("parseFile: %v", err)
	}

	type P struct{}
	scope := newScope(P{})
	text, err := renderTemplate(doc.template, scope)
	if err != nil {
		t.Fatalf("renderTemplate: %v", err)
	}
	if !strings.Contains(text, "line1\nline2") {
		t.Errorf("нет переноса строки: %q", text)
	}
}

// ── Тесты условий ────────────────────────────────────────────────────────────

type condProps struct {
	Count  int
	Name   string
	Active bool
}

func TestEvalCond_Truthy(t *testing.T) {
	scope := newScope(condProps{Count: 5})
	ok, err := evalCond("Count", scope)
	if err != nil || !ok {
		t.Errorf("Count=5 должен быть truthy, err=%v ok=%v", err, ok)
	}
}

func TestEvalCond_Not(t *testing.T) {
	scope := newScope(condProps{Active: false})
	ok, err := evalCond("!Active", scope)
	if err != nil || !ok {
		t.Errorf("!Active=false должен быть true, err=%v ok=%v", err, ok)
	}
}

func TestEvalCond_IntGt(t *testing.T) {
	scope := newScope(condProps{Count: 3})
	ok, err := evalCond("Count > 0", scope)
	if err != nil || !ok {
		t.Errorf("Count=3 > 0 должен быть true, err=%v ok=%v", err, ok)
	}
}

func TestEvalCond_StrEq(t *testing.T) {
	scope := newScope(condProps{Name: "bundle"})
	ok, err := evalCond(`Name == "bundle"`, scope)
	if err != nil || !ok {
		t.Errorf(`Name=="bundle" должен быть true, err=%v ok=%v`, err, ok)
	}
}

func TestEvalCond_And(t *testing.T) {
	scope := newScope(condProps{Count: 1, Active: true})
	ok, err := evalCond("Count > 0 && Active", scope)
	if err != nil || !ok {
		t.Errorf("Count>0 && Active должен быть true, err=%v ok=%v", err, ok)
	}
}

func TestEvalCond_Or_False(t *testing.T) {
	scope := newScope(condProps{Count: 0, Active: false})
	ok, err := evalCond("Count > 0 || Active", scope)
	if err != nil || ok {
		t.Errorf("Count=0>0 || Active=false должен быть false, err=%v ok=%v", err, ok)
	}
}

// ── Тесты keyboard ───────────────────────────────────────────────────────────

type kbTestProps struct {
	Offset     int
	PrevOffset int
	NextOffset int
	IsMy       bool
}

func TestRender_Keyboard_Buttons(t *testing.T) {
	src := `@props kbTestProps

@template
Text

@keyboard
[row]
[btn callback="sched:{PrevOffset}"] ◀ [/btn]
[btn callback="sched:{NextOffset}"] ▶ [/btn]
[/row]
`
	doc, err := parseFile(src)
	if err != nil {
		t.Fatalf("parseFile: %v", err)
	}

	scope := newScope(kbTestProps{PrevOffset: -1, NextOffset: 1})
	kb, err := renderKeyboard(doc.keyboard, scope)
	if err != nil {
		t.Fatalf("renderKeyboard: %v", err)
	}
	if kb == nil {
		t.Fatal("keyboard не должен быть nil")
	}
	if len(kb.InlineKeyboard) != 1 {
		t.Fatalf("ожидаем 1 строку, получили %d", len(kb.InlineKeyboard))
	}
	row := kb.InlineKeyboard[0]
	if len(row) != 2 {
		t.Fatalf("ожидаем 2 кнопки, получили %d", len(row))
	}
	if row[0].CallbackData == nil || *row[0].CallbackData != "sched:-1" {
		t.Errorf("callback[0] = %v, хотим %q", row[0].CallbackData, "sched:-1")
	}
	if row[1].CallbackData == nil || *row[1].CallbackData != "sched:1" {
		t.Errorf("callback[1] = %v, хотим %q", row[1].CallbackData, "sched:1")
	}
}

func TestRender_Keyboard_IfInRow(t *testing.T) {
	src := `@props kbTestProps

@template
T

@keyboard
[row]
{#if IsMy}
[btn callback="all"] Все [/btn]
{#else}
[btn callback="my"] Моё [/btn]
{/if}
[/row]
`
	doc, err := parseFile(src)
	if err != nil {
		t.Fatalf("parseFile: %v", err)
	}

	// isMy=true
	scope := newScope(kbTestProps{IsMy: true})
	kb, err := renderKeyboard(doc.keyboard, scope)
	if err != nil {
		t.Fatalf("renderKeyboard: %v", err)
	}
	if kb == nil || len(kb.InlineKeyboard) == 0 {
		t.Fatal("keyboard пуст")
	}
	btn := kb.InlineKeyboard[0][0]
	if btn.CallbackData == nil || *btn.CallbackData != "all" {
		t.Errorf("ожидали 'all', получили %v", btn.CallbackData)
	}

	// isMy=false
	scope2 := newScope(kbTestProps{IsMy: false})
	kb2, _ := renderKeyboard(doc.keyboard, scope2)
	btn2 := kb2.InlineKeyboard[0][0]
	if btn2.CallbackData == nil || *btn2.CallbackData != "my" {
		t.Errorf("ожидали 'my', получили %v", btn2.CallbackData)
	}
}

type eachKbProps struct {
	Links []kbLink
}

type kbLink struct {
	URL   string
	Label string
}

func TestRender_Keyboard_EachRows(t *testing.T) {
	src := `@props eachKbProps

@template
T

@keyboard
{#each Links as link}
[row]
[url href="{link.URL}"] {link.Label} [/url]
[/row]
{/each}
`
	doc, err := parseFile(src)
	if err != nil {
		t.Fatalf("parseFile: %v", err)
	}

	props := eachKbProps{
		Links: []kbLink{
			{URL: "https://a.com", Label: "A"},
			{URL: "https://b.com", Label: "B"},
		},
	}

	scope := newScope(props)
	kb, err := renderKeyboard(doc.keyboard, scope)
	if err != nil {
		t.Fatalf("renderKeyboard: %v", err)
	}
	if kb == nil {
		t.Fatal("keyboard nil")
	}
	if len(kb.InlineKeyboard) != 2 {
		t.Fatalf("ожидаем 2 строки, получили %d", len(kb.InlineKeyboard))
	}
	if kb.InlineKeyboard[0][0].URL == nil || *kb.InlineKeyboard[0][0].URL != "https://a.com" {
		t.Errorf("url[0] = %v", kb.InlineKeyboard[0][0].URL)
	}
	if kb.InlineKeyboard[1][0].URL == nil || *kb.InlineKeyboard[1][0].URL != "https://b.com" {
		t.Errorf("url[1] = %v", kb.InlineKeyboard[1][0].URL)
	}
}
