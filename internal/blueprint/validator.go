package blueprint

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type ValidationError struct {
	Field      string
	Rule       string
	Value      any
	Message    string
	Suggestion string
}

func (e ValidationError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("validasi gagal pada field %s untuk rule %s", e.Field, e.Rule)
}

var (
	regexCache = make(map[string]*regexp.Regexp)
	regexMu    sync.RWMutex
)

// ValidateInput memvalidasi satu pasangan (field, value) terhadap sebuah list aturan.
// Mengembalikan slice ValidationError jika ada validasi yang gagal.
func ValidateInput(field string, value any, rules []ValidationRule) []ValidationError {
	var errs []ValidationError

	for _, rule := range rules {
		if rule.Field != field {
			continue // rule ini bukan untuk field ini
		}

		if ruleErr := validateSingleRule(field, value, rule); ruleErr != nil {
			errs = append(errs, *ruleErr)
		}
	}

	return errs
}

func validateSingleRule(field string, value any, rule ValidationRule) *ValidationError {
	switch rule.Rule {
	case "required":
		if isEmpty(value) {
			msg := rule.Message
			if msg == "" {
				msg = fmt.Errorf("Field '%s' wajib diisi.", field).Error()
			}
			return &ValidationError{
				Field:      field,
				Rule:       rule.Rule,
				Value:      value,
				Message:    msg,
				Suggestion: fmt.Sprintf("Mohon masukkan nilai untuk %s", field),
			}
		}

	case "regex":
		strVal, ok := value.(string)
		if !ok || strVal == "" {
			return nil // kita hanya regex check pada string yang ada nilainya (jika require, pakai "required")
		}

		rx, err := getCompiledRegex(rule.Pattern)
		if err != nil {
			return &ValidationError{
				Field:      field,
				Rule:       rule.Rule,
				Value:      value,
				Message:    fmt.Sprintf("Konfigurasi template error: regex tidak valid. Pola: %s", rule.Pattern),
				Suggestion: "Hubungi author template.",
			}
		}

		if !rx.MatchString(strVal) {
			msg := rule.Message
			if msg == "" {
				msg = fmt.Sprintf("Nilai '%s' tidak memenuhi format yang diharapkan.", strVal)
			}
			return &ValidationError{
				Field:      field,
				Rule:       rule.Rule,
				Value:      value,
				Message:    msg,
				Suggestion: fmt.Sprintf("Harus cocok dengan format: %s", rule.Pattern),
			}
		}

	case "min_length":
		strVal, ok := value.(string)
		if !ok {
			return nil
		}
		minLen, err := strconv.Atoi(rule.Pattern)
		if err != nil {
			return nil // error konfigurasi pattern
		}
		if len(strVal) < minLen {
			msg := rule.Message
			if msg == "" {
				msg = fmt.Sprintf("Minimal %d karakter.", minLen)
			}
			return &ValidationError{
				Field:      field,
				Rule:       rule.Rule,
				Value:      value,
				Message:    msg,
				Suggestion: fmt.Sprintf("Tambahkan setidaknya %d karakter lagi.", minLen-len(strVal)),
			}
		}

	case "max_length":
		strVal, ok := value.(string)
		if !ok {
			return nil
		}
		maxLen, err := strconv.Atoi(rule.Pattern)
		if err != nil {
			return nil
		}
		if len(strVal) > maxLen {
			msg := rule.Message
			if msg == "" {
				msg = fmt.Sprintf("Maksimal %d karakter.", maxLen)
			}
			return &ValidationError{
				Field:      field,
				Rule:       rule.Rule,
				Value:      value,
				Message:    msg,
				Suggestion: fmt.Sprintf("Hapus setidaknya %d karakter.", len(strVal)-maxLen),
			}
		}
	}
	return nil
}

func isEmpty(v any) bool {
	if v == nil {
		return true
	}
	switch val := v.(type) {
	case string:
		return strings.TrimSpace(val) == ""
	case []any:
		return len(val) == 0
	case []string:
		return len(val) == 0
	}
	return false
}

func getCompiledRegex(pattern string) (*regexp.Regexp, error) {
	regexMu.RLock()
	rx, exists := regexCache[pattern]
	regexMu.RUnlock()

	if exists {
		return rx, nil
	}

	regexMu.Lock()
	defer regexMu.Unlock()

	// Double check setelah mendapat lock
	if rx, exists = regexCache[pattern]; exists {
		return rx, nil
	}

	c, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	regexCache[pattern] = c
	return c, nil
}
