// LaTeX rendering backend

package mmark

import (
	"bytes"
	"fmt"
	"strconv"
)

// LaTeX renderer configuration options.
const (
	LATEX_STANDALONE = 1 << iota // create standalone document
)

// Latex is a type that implements the Renderer interface for LaTeX output.
//
// Do not create this directly, instead use the LatexRenderer function.
type Latex struct {
	flags int    // LATEX_* options
	head  string // optionial latex file to be included

	// store the IAL we see for this block element
	ial *inlineAttr

	// titleBlock in TOML
	titleBlock *title

	appendix bool

	// index, map idx to id
	index      map[idx][]string
	indexCount int

	// (@good) example list group counter
	group map[string]int
}

// LatexRenderer creates and configures a Latex object, which
// satisfies the Renderer interface.
//
// flags is a set of LATEX_* options ORed together (currently no such options
// are defined).
func LatexRenderer(flags int) Renderer {
	return &Latex{}
}

func (options *Latex) GetFlags() int {
	return 0
}

// render code chunks using verbatim, or listings if we have a language
func (options *Latex) BlockCode(out *bytes.Buffer, text []byte, lang string, caption []byte, subfigure, callout bool) {
	if lang == "" {
		out.WriteString("\n\\begin{verbatim}\n")
	} else {
		out.WriteString("\n\\begin{lstlisting}[language=")
		out.WriteString(lang)
		out.WriteString("]\n")
	}
	out.Write(text)
	if lang == "" {
		out.WriteString("\n\\end{verbatim}\n")
	} else {
		out.WriteString("\n\\end{lstlisting}\n")
	}
}

func (options *Latex) Aside(out *bytes.Buffer, text []byte) {
}

func (options *Latex) TitleBlock(out *bytes.Buffer, text []byte) {

}

func (options *Latex) BlockQuote(out *bytes.Buffer, text, attribution []byte) {
	out.WriteString("\n\\begin{quotation}\n")
	out.Write(text)
	out.WriteString("\n\\end{quotation}\n")
}

func (options *Latex) CommentHtml(out *bytes.Buffer, text []byte) {
	doubleSpace(out)
	out.Write(text)
	out.WriteByte('\n')
}

func (options *Latex) BlockHtml(out *bytes.Buffer, text []byte) {
	// a pretty lame thing to do...
	out.WriteString("\n\\begin{verbatim}\n")
	out.Write(text)
	out.WriteString("\n\\end{verbatim}\n")
}

func (options *Latex) Flags() int {
	return options.flags
}

func (options *Latex) TitleBlockTOML(out *bytes.Buffer, block *title) {
}

func (options *Latex) Part(out *bytes.Buffer, text func() bool, id string) {
	if id != "" {
		out.WriteString(fmt.Sprintf("<h1 class=\"part\" id=\"%s\">", id))
	} else {
		out.WriteString(fmt.Sprintf("<h1 class=\"part\""))
	}
	text()
	out.WriteString(fmt.Sprintf("</h1>\n"))
}

// Section without numbering
func (options *Latex) Note(out *bytes.Buffer, text func() bool, id string) {
	options.inlineAttr() //reset the IAL
	if id != "" {
		out.WriteString(fmt.Sprintf("<h1 class=\"note\" id=\"%s\">", id))
	} else {
		out.WriteString(fmt.Sprintf("<h1 class=\"note\""))
	}
	text()
	out.WriteString(fmt.Sprintf("</h1>\n"))
}

func (options *Latex) SpecialHeader(out *bytes.Buffer, what []byte, text func() bool, id string) {
	if string(what) == "preface" {
		printf(nil, "handling preface like abstract")
		what = []byte("abstract")
	}
	ial := options.inlineAttr()

	out.WriteString("\n<abstract" + ial.String() + ">\n")
	return
}

func (options *Latex) Header(out *bytes.Buffer, text func() bool, level int, id string) {
	marker := out.Len()

	switch level {
	case 1:
		out.WriteString("\n\\section{")
	case 2:
		out.WriteString("\n\\subsection{")
	case 3:
		out.WriteString("\n\\subsubsection{")
	case 4:
		out.WriteString("\n\\paragraph{")
	case 5:
		out.WriteString("\n\\subparagraph{")
	case 6:
		out.WriteString("\n\\textbf{")
	}
	if !text() {
		out.Truncate(marker)
		return
	}
	out.WriteString("}\n")
}

func (options *Latex) HRule(out *bytes.Buffer) {
	out.WriteString("\n\\HRule\n")
}

func (options *Latex) CalloutCode(out *bytes.Buffer, index, id string) {
	out.WriteString("<span class=\"callout\">")
	out.WriteString(index)
	out.WriteString("</span>")
	return
}

func (options *Latex) CalloutText(out *bytes.Buffer, id string, ids []string) {
	for i, k := range ids {
		out.WriteString("<span class=\"callout\">")
		out.WriteString(k)
		out.WriteString("</span>")
		if i < len(ids)-1 {
			out.WriteString(" ")
		}
	}
}

func (options *Latex) List(out *bytes.Buffer, text func() bool, flags, start int, group []byte) {
	marker := out.Len()
	if flags&_LIST_TYPE_ORDERED != 0 {
		out.WriteString("\n\\begin{enumerate}\n")
	} else {
		out.WriteString("\n\\begin{itemize}\n")
	}
	if !text() {
		out.Truncate(marker)
		return
	}
	if flags&_LIST_TYPE_ORDERED != 0 {
		out.WriteString("\n\\end{enumerate}\n")
	} else {
		out.WriteString("\n\\end{itemize}\n")
	}
}

func (options *Latex) ListItem(out *bytes.Buffer, text []byte, flags int) {
	out.WriteString("\n\\item ")
	out.Write(text)
}

func (options *Latex) Example(out *bytes.Buffer, index int) {
	out.WriteByte('(')
	out.WriteString(strconv.Itoa(index))
	out.WriteByte(')')
}

func (options *Latex) Paragraph(out *bytes.Buffer, text func() bool, flags int) {
	marker := out.Len()
	out.WriteString("\n")
	if !text() {
		out.Truncate(marker)
		return
	}
	out.WriteString("\n")
}

func (options *Latex) Table(out *bytes.Buffer, header, body, footer []byte, columnData []int, caption []byte) {
	out.WriteString("\n\\begin{tabular}{")
	for _, elt := range columnData {
		switch elt {
		case _TABLE_ALIGNMENT_LEFT:
			out.WriteByte('l')
		case _TABLE_ALIGNMENT_RIGHT:
			out.WriteByte('r')
		default:
			out.WriteByte('c')
		}
	}
	out.WriteString("}\n")
	out.Write(header)
	out.WriteString(" \\\\\n\\hline\n")
	out.Write(body)
	out.WriteString("\n\\end{tabular}\n")
}

func (options *Latex) TableRow(out *bytes.Buffer, text []byte) {
	if out.Len() > 0 {
		out.WriteString(" \\\\\n")
	}
	out.Write(text)
}

func (options *Latex) TableHeaderCell(out *bytes.Buffer, text []byte, align, colspan int) {
	if out.Len() > 0 {
		out.WriteString(" & ")
	}
	out.Write(text)
}

func (options *Latex) TableCell(out *bytes.Buffer, text []byte, align, colspan int) {
	if out.Len() > 0 {
		out.WriteString(" & ")
	}
	out.Write(text)
}

// TODO: this
func (options *Latex) Footnotes(out *bytes.Buffer, text func() bool) {

}

func (options *Latex) FootnoteItem(out *bytes.Buffer, name, text []byte, flags int) {

}

func (options *Latex) Math(out *bytes.Buffer, text []byte, display bool) {
	ial := options.inlineAttr()
	s := ial.String()
	s = s
}

func (options *Latex) AutoLink(out *bytes.Buffer, link []byte, kind int) {
	out.WriteString("\\href{")
	if kind == _LINK_TYPE_EMAIL {
		out.WriteString("mailto:")
	}
	out.Write(link)
	out.WriteString("}{")
	out.Write(link)
	out.WriteString("}")
}

func (options *Latex) CodeSpan(out *bytes.Buffer, text []byte) {
	out.WriteString("\\texttt{")
	escapeSpecialChars(out, text)
	out.WriteString("}")
}

func (options *Latex) DoubleEmphasis(out *bytes.Buffer, text []byte) {
	out.WriteString("\\textbf{")
	out.Write(text)
	out.WriteString("}")
}

func (options *Latex) Emphasis(out *bytes.Buffer, text []byte) {
	out.WriteString("\\textit{")
	out.Write(text)
	out.WriteString("}")
}

func (options *Latex) Subscript(out *bytes.Buffer, text []byte) {
	out.WriteString("<sub>")
	out.Write(text)
	out.WriteString("</sub>")
}

func (options *Latex) Superscript(out *bytes.Buffer, text []byte) {
	out.WriteString("<sup>")
	out.Write(text)
	out.WriteString("</sup>")
}

func (options *Latex) Figure(out *bytes.Buffer, text []byte, caption []byte) {
	s := options.inlineAttr().String()
	out.WriteString("<figure role=\"group\"" + s + ">\n")
	out.WriteString("<figcaption>")
	out.Write(caption)
	out.WriteString("</figcaption>\n")
	out.Write(text)
	out.WriteString("</figure>\n")
}

func (options *Latex) Image(out *bytes.Buffer, link []byte, title []byte, alt []byte, subfigure bool) {
	if bytes.HasPrefix(link, []byte("http://")) || bytes.HasPrefix(link, []byte("https://")) {
		// treat it like a link
		out.WriteString("\\href{")
		out.Write(link)
		out.WriteString("}{")
		out.Write(alt)
		out.WriteString("}")
	} else {
		out.WriteString("\\includegraphics{")
		out.Write(link)
		out.WriteString("}")
	}
}

func (options *Latex) LineBreak(out *bytes.Buffer) {
	out.WriteString(" \\\\\n")
}

func (options *Latex) Link(out *bytes.Buffer, link []byte, title []byte, content []byte) {
	out.WriteString("\\href{")
	out.Write(link)
	out.WriteString("}{")
	out.Write(content)
	out.WriteString("}")
}

func (options *Latex) Abbreviation(out *bytes.Buffer, abbr, title []byte) {
	out.Write(abbr)
}

func (options *Latex) RawHtmlTag(out *bytes.Buffer, tag []byte) {
}

func (options *Latex) TripleEmphasis(out *bytes.Buffer, text []byte) {
	out.WriteString("\\textbf{\\textit{")
	out.Write(text)
	out.WriteString("}}")
}

func (options *Latex) StrikeThrough(out *bytes.Buffer, text []byte) {
	out.WriteString("\\sout{")
	out.Write(text)
	out.WriteString("}")
}

func (options *Latex) FootnoteRef(out *bytes.Buffer, ref []byte, id int) {

}

func (options *Latex) Index(out *bytes.Buffer, primary, secondary []byte, prim bool) {
	idx := idx{string(primary), string(secondary)}
	id := ""
	if ids, ok := options.index[idx]; ok {
		// write id out and add it to the list
		id = fmt.Sprintf("#idxref:%d-%d", options.indexCount, len(ids))
		options.index[idx] = append(options.index[idx], id)
	} else {
		id = fmt.Sprintf("#idxref:%d-0", options.indexCount)
		options.index[idx] = []string{id}
	}
	out.WriteString("<span class=\"index-ref\" id=\"" + id[1:] + "\"></span>")

	options.indexCount++
}

func (options *Latex) Citation(out *bytes.Buffer, link, title []byte) {
	out.WriteString("<a class=\"cite\" href=\"#")
	out.Write(bytes.ToLower(link))
	out.WriteString("\">")
	out.Write(title)
	out.WriteString("</a>")
}

func needsBackslash(c byte) bool {
	for _, r := range []byte("_{}%$&\\~#") {
		if c == r {
			return true
		}
	}
	return false
}

func escapeSpecialChars(out *bytes.Buffer, text []byte) {
	for i := 0; i < len(text); i++ {
		// directly copy normal characters
		org := i

		for i < len(text) && !needsBackslash(text[i]) {
			i++
		}
		if i > org {
			out.Write(text[org:i])
		}

		// escape a character
		if i >= len(text) {
			break
		}
		out.WriteByte('\\')
		out.WriteByte(text[i])
	}
}

func (options *Latex) Entity(out *bytes.Buffer, entity []byte) {
	// TODO: convert this into a unicode character or something
	out.Write(entity)
}

func (options *Latex) NormalText(out *bytes.Buffer, text []byte) {
	escapeSpecialChars(out, text)
}

// header and footer
func (options *Latex) DocumentHeader(out *bytes.Buffer, first bool) {
	if !first {
		return
	}
	out.WriteString("\\documentclass{article}\n")
	out.WriteString("\n")
	out.WriteString("\\usepackage{graphicx}\n")
	out.WriteString("\\usepackage{listings}\n")
	out.WriteString("\\usepackage[margin=1in]{geometry}\n")
	out.WriteString("\\usepackage[utf8]{inputenc}\n")
	out.WriteString("\\usepackage{verbatim}\n")
	out.WriteString("\\usepackage[normalem]{ulem}\n")
	out.WriteString("\\usepackage{hyperref}\n")
	out.WriteString("\n")
	out.WriteString("\\hypersetup{colorlinks,%\n")
	out.WriteString("  citecolor=black,%\n")
	out.WriteString("  filecolor=black,%\n")
	out.WriteString("  linkcolor=black,%\n")
	out.WriteString("  urlcolor=black,%\n")
	out.WriteString("  pdfstartview=FitH,%\n")
	out.WriteString("  breaklinks=true,%\n")
	out.WriteString("  pdfauthor={Blackfriday Markdown Processor v")
	out.WriteString(Version)
	out.WriteString("}}\n")
	out.WriteString("\n")
	out.WriteString("\\newcommand{\\HRule}{\\rule{\\linewidth}{0.5mm}}\n")
	out.WriteString("\\addtolength{\\parskip}{0.5\\baselineskip}\n")
	out.WriteString("\\parindent=0pt\n")
	out.WriteString("\n")
	out.WriteString("\\begin{document}\n")
}

func (options *Latex) DocumentFooter(out *bytes.Buffer, first bool) {
	if !first {
		return
	}
	out.WriteString("\n\\end{document}\n")
}

func (options *Latex) DocumentMatter(out *bytes.Buffer, matter int) {
	if matter == _DOC_BACK_MATTER {
		options.appendix = true
	}
}

func (options *Latex) References(out *bytes.Buffer, citations map[string]*citation) {
}

func (options *Latex) SetInlineAttr(i *inlineAttr) {
	options.ial = i
}

func (options *Latex) inlineAttr() *inlineAttr {
	if options.ial == nil {
		return newInlineAttr()
	}
	return options.ial
}
