package java

import (
	"log/slog"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/mitchellh/go-wordwrap"
)

type Header struct {
	Title    string
	SubTitle []string
}

func cleanText(text string) string {
	text = strings.ReplaceAll(text, "\u00a0", " ")
	text = strings.TrimSpace(text)
	return text
}

func normalizeWhitespace(text string) string {
	text = strings.ReplaceAll(text, "\u00a0", " ")
	text = strings.ReplaceAll(text, "\u200b", "")
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)
	return text
}

func parseHeader(doc *goquery.Document) Header {
	subTitle := make([]string, 0)
	header := doc.Find(".header").First()
	header.Find(".subTitle").Each(func(i int, s *goquery.Selection) {
		text := cleanText(s.Text())
		if text != "" {
			subTitle = append(subTitle, text)
		}
	})
	title := cleanText(header.Find(".title").First().Text())
	return Header{
		Title:    title,
		SubTitle: subTitle,
	}
}

func parseInheritance(doc *goquery.Document) []string {
	acc := make([]string, 0)
	doc.Find("ul.inheritance > li:first-child").Each(func(i int, s *goquery.Selection) {
		text := cleanText(s.Text())
		if text != "" {
			acc = append(acc, text)
		}
	})
	return acc
}

func parseDescription(doc *goquery.Document) string {
	description := doc.Find(".description").First().Text()
	acc := make([]string, 0)
	for line := range strings.SplitSeq(description, "\n") {
		text := cleanText(line)
		if text != "" {
			acc = append(acc, text)
		}
	}
	return strings.Join(acc, "\n")
}

type Constructor struct {
	Name        string
	Description string
}

func parseConstructorSummary(doc *goquery.Document) []Constructor {
	acc := make([]Constructor, 0)
	header := doc.Find("h3").FilterFunction(func(i int, s *goquery.Selection) bool {
		return s.Text() == "Constructor Summary"
	}).First()
	table := header.Next()
	table.Find("tr").Each(func(i int, s *goquery.Selection) {
		name := s.Find(".colConstructorName").First().Text()
		description := s.Find(".colLast").First().Text()
		if name != "" {
			acc = append(acc, Constructor{
				Name:        normalizeWhitespace(name),
				Description: normalizeWhitespace(description),
			})
		}
	})
	return acc
}

type Method struct {
	Modifier    string
	Name        string
	Description string
}

func parseMethodSummary(doc *goquery.Document) []Method {
	acc := make([]Method, 0)
	header := doc.Find("h3").FilterFunction(func(i int, s *goquery.Selection) bool {
		return s.Text() == "Method Summary"
	}).First()
	table := header.Next()
	table.Find("tr").Each(func(i int, s *goquery.Selection) {
		modifier := s.Find(".colFirst").First().Text()
		name := s.Find(".colSecond").First().Text()
		description := s.Find(".colLast").First().Text()
		if name != "" && name != "Method" {
			acc = append(acc, Method{
				Modifier:    normalizeWhitespace(modifier),
				Name:        normalizeWhitespace(name),
				Description: normalizeWhitespace(description),
			})
		}
	})
	return acc
}

func formatMarkdownHeader(text string) string {
	lines := make([]string, 0)
	lines = append(lines, text)
	lines = append(lines, strings.Repeat("=", len(text)))
	lines = append(lines, "")
	return strings.Join(lines, "\n")
}

const INDENT = "    "

func formatMarkdownDocument(
	title string,
	header Header,
	inheritance []string,
	description string,
	constructors []Constructor,
	methods []Method,
) string {
	lines := make([]string, 0)
	lines = append(lines, formatMarkdownHeader(title))
	lines = append(lines, header.SubTitle...)
	lines = append(lines, header.Title)
	lines = append(lines, "")
	if len(inheritance) > 0 {
		lines = append(lines, formatMarkdownHeader("Inheritance"))
		for i, line := range inheritance {
			lines = append(lines, strings.Repeat(INDENT, i)+line)
		}
		lines = append(lines, "")
	}
	lines = append(lines, formatMarkdownHeader("Description"))
	lines = append(lines, wordwrap.WrapString(description, 80))
	lines = append(lines, "")
	if len(constructors) > 0 {
		lines = append(lines, formatMarkdownHeader("Constructors"))
		for _, constructor := range constructors {
			lines = append(lines, constructor.Name)
			if constructor.Description != "" {
				lines = append(lines, INDENT+constructor.Description)
			}
			lines = append(lines, "")
		}
	}
	if len(methods) > 0 {
		lines = append(lines, formatMarkdownHeader("Methods"))
		for _, method := range methods {
			lines = append(lines, method.Modifier+" "+method.Name)
			if method.Description != "" {
				wrappedLines := strings.SplitSeq(
					wordwrap.WrapString(method.Description, 80), "\n")
				for line := range wrappedLines {
					lines = append(lines, INDENT+line)
				}
			}
			lines = append(lines, "")
		}
	}
	return strings.Join(lines, "\n")
}

func FormatMarkdown(filename string) string {
	// Open the file
	r, err := os.Open(filename)
	if err != nil {
		slog.Error("Error opening file", "filename", filename, "error", err)
		os.Exit(1)
	}
	defer r.Close()
	// Parse HTML document
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		slog.Error("Error parsing HTML", "filename", filename, "error", err)
		os.Exit(1)
	}
	title := doc.Find("title").Text()
	header := parseHeader(doc)
	inheritance := parseInheritance(doc)
	description := parseDescription(doc)
	constructors := parseConstructorSummary(doc)
	methods := parseMethodSummary(doc)
	return formatMarkdownDocument(title, header, inheritance, description, constructors, methods)
}
