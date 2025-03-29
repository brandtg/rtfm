package core

import (
	"fmt"
	"log/slog"
	"net/http"
	"os/exec"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

/*

def parse_standard_library_classnames(java_version):
    """
    Finds links to Javadoc for the JDK from the web.
    """
    base_url = (
        "https://docs.oracle.com/en/java/javase/" + str(java_version) + "/docs/api/"
    )
    all_classes_url = base_url + "allclasses-index.html"
    with request.urlopen(all_classes_url) as response:
        logging.info("GET %s => %s", all_classes_url, response.getcode())
        html_content = response.read().decode("utf-8")
        parser = LinkExtractor(include_external=True)
        parser.feed(html_content)
        return [
            dict(
                name=link.replace(".html", "").replace("/", "."),
                path=base_url + link,
                jar=JDK,
            )
            for link in parser.get_links()
        ]
*/

func findJavaVersion() (string, error) {
	cmd := exec.Command("java", "-version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	re := regexp.MustCompile(`version "(\d+)\.`)
	match := re.FindStringSubmatch(string(out))
	if len(match) < 2 {
		return "", fmt.Errorf("could not find Java version in output: %s", string(out))
	}
	javaVersion := match[1]
	slog.Info("Java version found", "version", javaVersion)
	return javaVersion, nil
}

type Link struct {
	title string
	href  string
}

type JavaClass struct {
	name string
	path string
	jar  string
}

func javaClassFromLink(baseUrl string, link Link) JavaClass {
	name := strings.ReplaceAll(link.href, ".html", "")
	name = strings.ReplaceAll(name, "/", ".")
	return JavaClass{
		name: name,
		path: baseUrl + link.href,
		jar:  "JDK",
	}
}

func findJDKClasses() ([]JavaClass, error) {
	javaVersion, err := findJavaVersion()
	if err != nil {
		slog.Error("Error finding Java version", "error", err)
		return nil, err
	}
	url := fmt.Sprintf("https://docs.oracle.com/en/java/javase/%s/docs/api/", javaVersion)
	slog.Info("Fetching JDK classes from URL", "url", url)
	res, err := http.Get(url + "allclasses-index.html")
	if err != nil {
		slog.Error("Error fetching URL", "url", url, "error", err)
		return nil, err
	}
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		slog.Error("Error parsing HTML", "url", url, "error", err)
		return nil, err
	}
	slog.Info("Parsed document", "title", doc.Find("title").Text())
	links := []Link{}
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, hasHref := s.Attr("href")
		if hasHref {
			title, hasTitle := s.Attr("title")
			if hasTitle && (strings.Contains(title, "class in") || strings.Contains(title, "interface in")) {
				links = append(links, Link{title, href})
			}
		}
	})
	slog.Info("Extracted links", "count", len(links))
	javaClasses := []JavaClass{}
	for _, link := range links {
        javaClasses = append(javaClasses, javaClassFromLink(url, link))
	}
	return javaClasses, nil
}
