package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func fmtText(text string) (string, error) {
	ret := ""
	scanner := bufio.NewScanner(strings.NewReader(text))

	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.Replace(line, "  ", "", -1)
		line = strings.Replace(line, "\t", "", -1)
		if len(line) > 0 &&
			!strings.HasPrefix(line, "<img") &&
			!strings.Contains(line, "※ご利用のPCやブラウザにより") &&
			!strings.Contains(line, "Copyright © KANJIDIC2") {
			//fmt.Println(line)
			ret += line + "\n"
		}
	}
	if err := scanner.Err(); err != nil {
		errMsg := errors.New("读取时发生错误:" + err.Error())
		return "", errMsg
	}

	return ret, nil
}

func main() {
	searchQuery := "かまってちゃん"

	searchURL := fmt.Sprintf("https://www.weblio.jp/content/%s", url.QueryEscape(searchQuery))

	proxyURL, err := url.Parse("http://127.0.0.1:10809")
	if err != nil {
		log.Printf("Failed to parse proxy URL: %v", err)
	}
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept-Language", "ja,en-US;q=0.7,en;q=0.3")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to fetch URL: %s", resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("Failed to parse HTML: %v", err)
	}

	var results []map[string]interface{}
	discNames := []string{}

	doc.Find("div .kiji").Each(func(i int, s *goquery.Selection) {
		discName, err := fmtText(s.Text())
		if err != nil {
			log.Printf("Failed to fmt: %v", err)
			return
		}
		discNames = append(discNames, discName)
	})
	doc.Find(".pbarTL").Each(func(i int, s *goquery.Selection) {
		dicName, err := fmtText(s.Text())
		if err != nil {
			log.Printf("Failed to fmt: %v", err)
			return
		}
		item := map[string]interface{}{
			"dict": dicName,
			"data": discNames[i],
		}
		results = append(results, item)

	})
	retStr := ""
	for v := range results {
		dictInV := results[v]["dict"].(string)
		if dictInV == "Weblio日本語例文用例辞書\n" || dictInV == "ウィキペディア\n" {
			continue
		}
		dataInV := results[v]["data"].(string)
		retStr += "===========================\n"
		retStr += "### " + dictInV + "\n"
		retStr += dataInV + "\n"

	}
	fmt.Println(retStr)
}
