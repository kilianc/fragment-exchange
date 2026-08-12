package web

import (
	"html"
	"strings"
)

// Syntax highlighting, in about a hundred lines.
//
// Every code sample on this site is HTML, JavaScript or Go, and the thing a
// reader most needs to pick out is an fx-* attribute — so those get the accent
// colour and everything else gets the usual four. A real parser would be more
// correct and would also be the largest thing in the repository.

// highlight returns HTML-escaped source wrapped in <span class="c-*"> tags.
// An unknown language is escaped and left alone.
func highlight(lang, src string) string {
	switch lang {
	case "html":
		return highlightHTML(src)
	case "js", "javascript":
		return highlightCode(src, jsKeywords, true)
	case "go":
		return highlightCode(src, goKeywords, true)
	case "http":
		return highlightHTTP(src)
	default:
		return html.EscapeString(src)
	}
}

var jsKeywords = words(`async await break case catch class const continue default delete do else export
	extends finally for from function if import in instanceof let new of return static super switch this
	throw try typeof var void while yield true false null undefined`)

var goKeywords = words(`break case chan const continue default defer else fallthrough for func go goto if
	import interface map package range return select struct switch type var true false nil string int bool
	byte error rune`)

func words(s string) map[string]bool {
	m := map[string]bool{}
	for _, w := range strings.Fields(s) {
		m[w] = true
	}
	return m
}

func span(class, text string) string {
	return `<span class="c-` + class + `">` + html.EscapeString(text) + `</span>`
}

// highlightHTML colours tags, attribute names and quoted values, and gives
// fx-* attributes the accent so they stand out in a wall of markup.
func highlightHTML(src string) string {
	var b strings.Builder
	i := 0

	for i < len(src) {
		if strings.HasPrefix(src[i:], "<!--") {
			end := strings.Index(src[i:], "-->")
			if end < 0 {
				end = len(src) - i
			} else {
				end += 3
			}
			b.WriteString(span("comment", src[i:i+end]))
			i += end
			continue
		}

		if src[i] != '<' {
			j := strings.IndexByte(src[i:], '<')
			if j < 0 {
				b.WriteString(html.EscapeString(src[i:]))
				break
			}
			b.WriteString(html.EscapeString(src[i : i+j]))
			i += j
			continue
		}

		// A tag: <name attr="value" ...>
		end := strings.IndexByte(src[i:], '>')
		if end < 0 {
			b.WriteString(html.EscapeString(src[i:]))
			break
		}
		b.WriteString(highlightTag(src[i : i+end+1]))
		i += end + 1
	}

	return b.String()
}

func highlightTag(tag string) string {
	var b strings.Builder
	i := 0

	// "<" plus the tag name, or "</" plus the tag name.
	b.WriteString(span("tag", "<"))
	i++
	if i < len(tag) && tag[i] == '/' {
		b.WriteString(span("tag", "/"))
		i++
	}
	start := i
	for i < len(tag) && (isName(tag[i])) {
		i++
	}
	b.WriteString(span("tag", tag[start:i]))

	for i < len(tag) {
		switch {
		case tag[i] == '>' || tag[i] == '/':
			b.WriteString(span("tag", string(tag[i])))
			i++
		case isSpace(tag[i]):
			b.WriteByte(tag[i])
			i++
		case isName(tag[i]):
			start := i
			for i < len(tag) && isName(tag[i]) {
				i++
			}
			name := tag[start:i]
			class := "attr"
			if strings.HasPrefix(name, "fx-") || name == "fx-refresh" {
				class = "fx"
			}
			b.WriteString(span(class, name))
		case tag[i] == '=':
			b.WriteString("=")
			i++
		case tag[i] == '"' || tag[i] == '\'':
			quote := tag[i]
			start := i
			i++
			for i < len(tag) && tag[i] != quote {
				i++
			}
			if i < len(tag) {
				i++
			}
			b.WriteString(span("string", tag[start:i]))
		default:
			b.WriteString(html.EscapeString(string(tag[i])))
			i++
		}
	}

	return b.String()
}

// highlightCode handles the C-family shape shared by JavaScript and Go:
// // and /* */ comments, quoted strings, numbers and a keyword list.
func highlightCode(src string, keywords map[string]bool, backticks bool) string {
	var b strings.Builder
	i := 0

	for i < len(src) {
		switch {
		case strings.HasPrefix(src[i:], "//"):
			end := strings.IndexByte(src[i:], '\n')
			if end < 0 {
				end = len(src) - i
			}
			b.WriteString(span("comment", src[i:i+end]))
			i += end

		case strings.HasPrefix(src[i:], "/*"):
			end := strings.Index(src[i:], "*/")
			if end < 0 {
				end = len(src) - i
			} else {
				end += 2
			}
			b.WriteString(span("comment", src[i:i+end]))
			i += end

		case src[i] == '"' || src[i] == '\'' || (backticks && src[i] == '`'):
			quote := src[i]
			start := i
			i++
			for i < len(src) && src[i] != quote {
				if src[i] == '\\' && i+1 < len(src) {
					i++
				}
				i++
			}
			if i < len(src) {
				i++
			}
			b.WriteString(span("string", src[start:i]))

		case isLetter(src[i]):
			start := i
			for i < len(src) && isName(src[i]) {
				i++
			}
			word := src[start:i]
			if keywords[word] {
				b.WriteString(span("keyword", word))
			} else {
				b.WriteString(html.EscapeString(word))
			}

		case isDigit(src[i]):
			start := i
			for i < len(src) && (isDigit(src[i]) || src[i] == '.' || src[i] == '_') {
				i++
			}
			b.WriteString(span("number", src[start:i]))

		default:
			b.WriteString(html.EscapeString(string(src[i])))
			i++
		}
	}

	return b.String()
}

// highlightHTTP colours a request or response transcript: the start line, then
// Header: value pairs.
func highlightHTTP(src string) string {
	var b strings.Builder

	for n, line := range strings.Split(src, "\n") {
		if n > 0 {
			b.WriteByte('\n')
		}
		switch {
		case strings.HasPrefix(line, "#"):
			b.WriteString(span("comment", line))
		case strings.HasPrefix(line, "GET "), strings.HasPrefix(line, "POST "), strings.HasPrefix(line, "HTTP/"):
			b.WriteString(span("keyword", line))
		default:
			if name, value, ok := strings.Cut(line, ":"); ok && !strings.HasPrefix(line, " ") {
				class := "attr"
				if strings.EqualFold(strings.TrimSpace(name), "fx-target") {
					class = "fx"
				}
				b.WriteString(span(class, name) + ":" + html.EscapeString(value))
				continue
			}
			b.WriteString(html.EscapeString(line))
		}
	}

	return b.String()
}

func isSpace(c byte) bool  { return c == ' ' || c == '\t' || c == '\n' || c == '\r' }
func isDigit(c byte) bool  { return c >= '0' && c <= '9' }
func isLetter(c byte) bool { return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_' }
func isName(c byte) bool   { return isLetter(c) || isDigit(c) || c == '-' || c == '.' || c == ':' }
