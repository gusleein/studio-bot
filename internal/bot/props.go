package bot

import (
	"fmt"
	"time"

	"github.com/yourstudio/studio-bot/internal/domain"
)

// ── Props structs (входные данные для botview шаблонов) ─────────────────────

// WelcomeProps — пропсы для шаблона welcome.botview.
type WelcomeProps struct {
	FirstName string
}

// ScheduleLessonProps — одно занятие в расписании.
type ScheduleLessonProps struct {
	StartsAt string // "15:04"
	EndsAt   string // "15:04"
	Room     string
	Notes    string
}

// ScheduleDayProps — группа занятий за один день.
type ScheduleDayProps struct {
	Label   string // "Понедельник, 14 апреля"
	Lessons []ScheduleLessonProps
}

// ScheduleProps — пропсы для шаблона schedule.botview.
type ScheduleProps struct {
	WeekTitle      string
	IsEmpty        bool
	Days           []ScheduleDayProps
	PrevCallback   string // "sched:-1" или "sched_my:-1"
	NextCallback   string // "sched:1" или "sched_my:1"
	ToggleCallback string // "sched:0" или "sched_my:0"
	ToggleLabel    string // "📋 Всё расписание" или "📌 Моё расписание"
}

// SubscriptionItemProps — одна активная подписка.
type SubscriptionItemProps struct {
	TypeName  string // "🎵 Sample Packs"
	ExpiresAt string // "02.01.2006"
}

// SubLinkProps — ссылка на оформление подписки.
type SubLinkProps struct {
	Label string // "🎵 Оформить Sample Packs"
	URL   string
}

// SubscriptionProps — пропсы для шаблона subscription.botview.
type SubscriptionProps struct {
	HasSubs bool
	Subs    []SubscriptionItemProps
	Links   []SubLinkProps
}

// ── Построители пропсов ──────────────────────────────────────────────────────

// buildScheduleProps собирает ScheduleProps из domain-объектов.
func buildScheduleProps(schedules []*domain.Schedule, offset int, isMy bool) ScheduleProps {
	// Группировка по датам
	type dayGroup struct {
		date    time.Time
		key     string
		entries []*domain.Schedule
	}

	var groups []dayGroup
	groupIndex := map[string]int{}

	for _, s := range schedules {
		key := s.StartsAt.Format("2006-01-02")
		if idx, ok := groupIndex[key]; ok {
			groups[idx].entries = append(groups[idx].entries, s)
		} else {
			groupIndex[key] = len(groups)
			groups = append(groups, dayGroup{
				date:    s.StartsAt,
				key:     key,
				entries: []*domain.Schedule{s},
			})
		}
	}

	// Строим дни
	days := make([]ScheduleDayProps, 0, len(groups))
	for _, g := range groups {
		lessons := make([]ScheduleLessonProps, 0, len(g.entries))
		for _, s := range g.entries {
			lessons = append(lessons, ScheduleLessonProps{
				StartsAt: s.StartsAt.Format("15:04"),
				EndsAt:   s.EndsAt.Format("15:04"),
				Notes:    s.Notes,
			})
		}
		days = append(days, ScheduleDayProps{
			Label:   formatDate(g.date),
			Lessons: lessons,
		})
	}

	// Callback-строки для навигации
	prefix := "sched"
	if isMy {
		prefix = "sched_my"
	}
	prevCb := fmt.Sprintf("%s:%d", prefix, offset-1)
	nextCb := fmt.Sprintf("%s:%d", prefix, offset+1)

	var toggleCb, toggleLabel string
	if isMy {
		toggleCb = fmt.Sprintf("sched:%d", offset)
		toggleLabel = "📋 Всё расписание"
	} else {
		toggleCb = fmt.Sprintf("sched_my:%d", offset)
		toggleLabel = "📌 Моё расписание"
	}

	return ScheduleProps{
		WeekTitle:      weekTitle(offset),
		IsEmpty:        len(schedules) == 0,
		Days:           days,
		PrevCallback:   prevCb,
		NextCallback:   nextCb,
		ToggleCallback: toggleCb,
		ToggleLabel:    toggleLabel,
	}
}

// buildSubscriptionProps собирает SubscriptionProps из domain-объектов.
func buildSubscriptionProps(subs []*domain.Subscription, links []SubLinkProps) SubscriptionProps {
	typeNames := map[domain.SubscriptionType]string{
		domain.SubscriptionTypeSamplePacks: "🎵 Sample Packs",
		domain.SubscriptionTypePresets:     "🎛 Presets",
		domain.SubscriptionTypeBundle:      "📦 Bundle (всё включено)",
	}

	items := make([]SubscriptionItemProps, 0, len(subs))
	for _, sub := range subs {
		name, ok := typeNames[sub.Type]
		if !ok {
			name = string(sub.Type)
		}
		items = append(items, SubscriptionItemProps{
			TypeName:  name,
			ExpiresAt: sub.ExpiresAt.Format("02.01.2006"),
		})
	}

	return SubscriptionProps{
		HasSubs: len(subs) > 0,
		Subs:    items,
		Links:   links,
	}
}
