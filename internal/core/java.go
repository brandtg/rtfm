package core

import (
	"fmt"
	"log/slog"
	"net/http"
	"os/exec"
	"regexp"
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

func findJDKClasses() {
	javaVersion, err := findJavaVersion()
	if err != nil {
		panic(err)
	}
	url := fmt.Sprintf("https://docs.oracle.com/en/java/javase/%s/docs/api/allclasses-index.html", javaVersion)
	slog.Info("Fetching JDK classes from URL", "url", url)
	res, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	slog.Info("Parsed document", "title", doc.Find("title").Text())
}
