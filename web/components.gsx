package web

import "strings"

// Section is an <h2> with an anchor, and whatever follows it.
func Section(id string, title string, children ...Node) Node {
	return (
		<>
			<h2 id={id}>{title}</h2>
			{Group(children)}
		</>
	)
}

// Snippet is a labelled code block. The label says what language it is and,
// more usefully, which file it belongs in.
func Snippet(lang string, label string, code string) Node {
	body := <pre><code>{Raw(highlight(lang, strings.TrimSpace(code)))}</code></pre>

	if label == "" {
		return body
	}

	return (
		<div class="snippet">
			<div class="snippet-label">{label}</div>
			{body}
		</div>
	)
}

// Code is a labelled code block with the language as its label.
func Code(lang string, code string) Node {
	return Snippet(lang, lang, code)
}

// Note is an aside the reader should not skip.
func Note(title string, children ...Node) Node {
	return (
		<div class="note">
			{If(title != "", <strong class="note-title">{title}</strong>)}
			{Group(children)}
		</div>
	)
}

// Cols puts two blocks side by side, for a before and an after.
func Cols(left Node, right Node) Node {
	return (
		<div class="cols">
			{left}
			{right}
		</div>
	)
}

// Steps is a numbered list.
func Steps(items ...Node) Node {
	var lis []Node
	for _, item := range items {
		var body Node = item
		lis = append(lis, <li>{body}</li>)
	}
	return <ol class="steps">{Group(lis)}</ol>
}

// Pull is a sentence worth stopping on.
func Pull(children ...Node) Node {
	return <p class="pull">{Group(children)}</p>
}

// Table renders a header row and body rows.
func Table(headers []string, rows [][]Node) Node {
	var ths []Node
	for _, h := range headers {
		ths = append(ths, <th>{h}</th>)
	}

	var trs []Node
	for _, row := range rows {
		var tds []Node
		for _, cell := range row {
			var body Node = cell
			tds = append(tds, <td>{body}</td>)
		}
		trs = append(trs, <tr>{Group(tds)}</tr>)
	}

	return (
		<table>
			<thead><tr>{Group(ths)}</tr></thead>
			<tbody>{Group(trs)}</tbody>
		</table>
	)
}

// C is inline code.
func C(s string) Node { return <code>{s}</code> }

// Row is a shorthand for a table row.
func Row(cells ...Node) []Node { return cells }
