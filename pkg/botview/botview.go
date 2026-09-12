// Package botview реализует легковесный шаблонизатор для Telegram-бота.
//
// Шаблоны описываются в файлах .botview и содержат три секции:
//
//	@props TypeName      — имя Go-структуры пропсов (для документации)
//	@template            — текст сообщения с подстановками {Field}, {#if}, {#each}
//	@keyboard            — inline-клавиатура с [row], [btn], [url]
//
// Пример использования:
//
//	renderer, _ := botview.New("internal/bot/templates")
//	result, _ := renderer.Render("schedule", props)
//	msg.Text = result.Text
//	msg.ReplyMarkup = result.Keyboard
package botview

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// RenderResult содержит результат рендеринга шаблона.
type RenderResult struct {
	// Text — HTML-текст для отправки в Telegram (ParseMode = HTML).
	Text string
	// Keyboard — inline-клавиатура или nil, если секции @keyboard нет.
	Keyboard *tgbotapi.InlineKeyboardMarkup
}

// Renderer загружает и кэширует .botview шаблоны из директории.
type Renderer struct {
	dir   string
	mu    sync.RWMutex
	cache map[string]*document
}

// New создаёт Renderer, читающий шаблоны из заданной директории.
func New(dir string) *Renderer {
	return &Renderer{
		dir:   dir,
		cache: make(map[string]*document),
	}
}

// Render рендерит шаблон name с заданными пропсами.
//
// name — имя файла без расширения (например, "schedule" → "schedule.botview").
// props — Go-структура с данными, поля которой доступны в шаблоне.
func (r *Renderer) Render(name string, props any) (*RenderResult, error) {
	doc, err := r.load(name)
	if err != nil {
		return nil, fmt.Errorf("botview Render(%q): %w", name, err)
	}

	scope := newScope(props)

	text, err := renderTemplate(doc.template, scope)
	if err != nil {
		return nil, fmt.Errorf("botview Render(%q) template: %w", name, err)
	}

	kb, err := renderKeyboard(doc.keyboard, scope)
	if err != nil {
		return nil, fmt.Errorf("botview Render(%q) keyboard: %w", name, err)
	}

	return &RenderResult{
		Text:     text,
		Keyboard: kb,
	}, nil
}

// Reload сбрасывает кэш шаблонов (удобно при разработке).
func (r *Renderer) Reload() {
	r.mu.Lock()
	r.cache = make(map[string]*document)
	r.mu.Unlock()
}

// load возвращает кэшированный или загруженный документ.
func (r *Renderer) load(name string) (*document, error) {
	r.mu.RLock()
	doc, ok := r.cache[name]
	r.mu.RUnlock()
	if ok {
		return doc, nil
	}

	path := filepath.Join(r.dir, name+".botview")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать шаблон %q: %w", path, err)
	}

	doc, err = parseFile(string(data))
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга шаблона %q: %w", path, err)
	}

	r.mu.Lock()
	r.cache[name] = doc
	r.mu.Unlock()

	return doc, nil
}
