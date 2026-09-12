package botview

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// evalScope — стек переменных (для переменных цикла {#each}).
type evalScope struct {
	vars []scopeVar
	root reflect.Value
}

type scopeVar struct {
	name string
	val  reflect.Value
}

// newScope создаёт scope с корневым объектом props.
func newScope(props any) evalScope {
	return evalScope{root: reflect.ValueOf(props)}
}

// push добавляет переменную цикла в scope.
func (s evalScope) push(name string, val reflect.Value) evalScope {
	newVars := make([]scopeVar, len(s.vars)+1)
	copy(newVars, s.vars)
	newVars[len(s.vars)] = scopeVar{name: name, val: val}
	return evalScope{vars: newVars, root: s.root}
}

// resolve возвращает reflect.Value для пути вида "Field", "nested.Field" или "loopVar.Field".
func (s evalScope) resolve(path string) (reflect.Value, error) {
	parts := strings.SplitN(path, ".", 2)
	head := parts[0]

	// Ищем в переменных scope (цикловые переменные)
	for i := len(s.vars) - 1; i >= 0; i-- {
		if s.vars[i].name == head {
			val := s.vars[i].val
			if len(parts) == 2 {
				return getFieldByPath(val, parts[1])
			}
			return val, nil
		}
	}

	// Ищем в корневом объекте
	return getFieldByPath(s.root, path)
}

// getFieldByPath обходит путь "A.B.C" по reflect.Value.
func getFieldByPath(v reflect.Value, path string) (reflect.Value, error) {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	for _, part := range strings.Split(path, ".") {
		if v.Kind() != reflect.Struct {
			return reflect.Value{}, fmt.Errorf("не структура при обращении к полю %q: %v", part, v.Kind())
		}
		v = v.FieldByName(part)
		if !v.IsValid() {
			return reflect.Value{}, fmt.Errorf("поле %q не найдено", part)
		}
		for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
			v = v.Elem()
		}
	}
	return v, nil
}

// toString конвертирует reflect.Value в строку для вывода в шаблоне.
func toString(v reflect.Value) string {
	if !v.IsValid() {
		return ""
	}
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	case reflect.Bool:
		if v.Bool() {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}

// isTruthy возвращает true, если значение "непустое".
func isTruthy(v reflect.Value) bool {
	if !v.IsValid() {
		return false
	}
	switch v.Kind() {
	case reflect.Bool:
		return v.Bool()
	case reflect.String:
		return v.String() != ""
	case reflect.Slice, reflect.Array, reflect.Map:
		return v.Len() > 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() != 0
	case reflect.Ptr, reflect.Interface:
		return !v.IsNil()
	default:
		return true
	}
}

// ── Условные выражения ───────────────────────────────────────────────────────

// evalCond вычисляет строковое условие из {#if ...}.
//
// Поддерживаемый синтаксис:
//
//	Field             → isTruthy(Field)
//	!Field            → !isTruthy(Field)
//	Field > N         → int/float сравнение
//	Field < N
//	Field >= N
//	Field <= N
//	Field == "str"    → строковое сравнение
//	Field != "str"
//	A && B            → логическое И
//	A || B            → логическое ИЛИ
func evalCond(cond string, scope evalScope) (bool, error) {
	// Нормализуем пробелы
	cond = strings.TrimSpace(cond)

	// Ищем || сначала (низший приоритет)
	if idx := findLogOp(cond, "||"); idx >= 0 {
		left := cond[:idx]
		right := cond[idx+2:]
		lv, err := evalCond(left, scope)
		if err != nil {
			return false, err
		}
		if lv {
			return true, nil
		}
		return evalCond(right, scope)
	}

	// Ищем &&
	if idx := findLogOp(cond, "&&"); idx >= 0 {
		left := cond[:idx]
		right := cond[idx+2:]
		lv, err := evalCond(left, scope)
		if err != nil {
			return false, err
		}
		if !lv {
			return false, nil
		}
		return evalCond(right, scope)
	}

	return evalSimpleCond(cond, scope)
}

// findLogOp ищет индекс логического оператора op вне кавычек.
func findLogOp(s, op string) int {
	inQuote := false
	for i := 0; i < len(s)-len(op)+1; i++ {
		if s[i] == '"' {
			inQuote = !inQuote
		}
		if !inQuote && strings.HasPrefix(s[i:], op) {
			return i
		}
	}
	return -1
}

// evalSimpleCond вычисляет простое условие без && и ||.
func evalSimpleCond(cond string, scope evalScope) (bool, error) {
	cond = strings.TrimSpace(cond)

	// Отрицание
	neg := false
	if strings.HasPrefix(cond, "!") {
		neg = true
		cond = strings.TrimSpace(cond[1:])
	}

	// Операторы сравнения
	for _, op := range []string{">=", "<=", "!=", "==", ">", "<"} {
		if idx := strings.Index(cond, op); idx >= 0 {
			left := strings.TrimSpace(cond[:idx])
			right := strings.TrimSpace(cond[idx+len(op):])

			lv, err := scope.resolve(left)
			if err != nil {
				return false, err
			}

			result, err := compare(lv, op, right)
			if err != nil {
				return false, err
			}
			if neg {
				return !result, nil
			}
			return result, nil
		}
	}

	// Простая truthy-проверка
	v, err := scope.resolve(cond)
	if err != nil {
		return false, err
	}
	result := isTruthy(v)
	if neg {
		return !result, nil
	}
	return result, nil
}

// compare сравнивает reflect.Value с правым операндом (строка или число).
func compare(lv reflect.Value, op, right string) (bool, error) {
	// Строковый литерал
	if strings.HasPrefix(right, `"`) && strings.HasSuffix(right, `"`) {
		rv := right[1 : len(right)-1]
		ls := toString(lv)
		switch op {
		case "==":
			return ls == rv, nil
		case "!=":
			return ls != rv, nil
		default:
			return false, fmt.Errorf("операция %q не применима к строкам", op)
		}
	}

	// Числовой литерал
	rn, err := strconv.ParseFloat(right, 64)
	if err != nil {
		return false, fmt.Errorf("неизвестный правый операнд %q", right)
	}

	var ln float64
	switch lv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ln = float64(lv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ln = float64(lv.Uint())
	case reflect.Float32, reflect.Float64:
		ln = lv.Float()
	default:
		return false, fmt.Errorf("поле не является числом для сравнения %q", op)
	}

	switch op {
	case ">":
		return ln > rn, nil
	case "<":
		return ln < rn, nil
	case ">=":
		return ln >= rn, nil
	case "<=":
		return ln <= rn, nil
	case "==":
		return ln == rn, nil
	case "!=":
		return ln != rn, nil
	}
	return false, fmt.Errorf("неизвестный оператор %q", op)
}

// interpolate подставляет {FieldPath} в строку атрибута.
// Используется для атрибутов кнопок: callback="sched:{Offset}".
func interpolate(s string, scope evalScope) string {
	var result strings.Builder
	i := 0
	runes := []rune(s)

	for i < len(runes) {
		if runes[i] == '{' {
			// Ищем закрывающую }
			j := i + 1
			for j < len(runes) && runes[j] != '}' {
				j++
			}
			if j < len(runes) {
				path := strings.TrimSpace(string(runes[i+1 : j]))
				if v, err := scope.resolve(path); err == nil {
					result.WriteString(toString(v))
				} else {
					// Оставляем как есть при ошибке
					result.WriteString(string(runes[i : j+1]))
				}
				i = j + 1
				continue
			}
		}
		result.WriteRune(runes[i])
		i++
	}

	return result.String()
}
