package main

import (
    "fmt"
    "flag"
    "io"
    "log"
    "net/http"
    "net/url"
    "strings"
    "slices"
    "regexp"
    "os"
    "os/exec"

    "github.com/CBaier33/html-parser"
)

type level struct{
    Name string
    Children []string
}

func main() {
    urlFlag := flag.String("url", "https://www.avemaria.edu", "the url to build sitemap")
    maxDepth := flag.Int("depth", 20, "the maximum number of link levels to recurse.")
    verbose := flag.Bool("v", false, "Print verbose output")
    flag.Parse()

    
    fmt.Println("Parsing Site.")
    pages := bfs(*urlFlag, *maxDepth)
    fmt.Println("Parsed.")

    fmt.Println("Creating Map...")
    map1 := levels(pages)

    // Create .dot file output
	cmd := exec.Command("echo", "digraph site {\nrankdir=LR;\nsize=100\nlayout=sfdp\noverlap=prism\n#beautify=true\n#smoothing=triangle\n#quadtree=fast\npack=false")
    file, _ := os.Create("sitemap.dot")
    cmd.Stdout = file
    err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

    loaded := make(map[string]struct{})
    for _, level := range map1 {
        a := "\"" + level.Name + "\""
        for _, child := range level.Children {

            b := "\""+ child + "\""

            if _, ok := loaded[fmt.Sprintf("%s-%s", a, b)]; ok {
                continue
            } else {
                loaded[fmt.Sprintf("%s-%s", a, b)] = struct{}{}
                str := fmt.Sprintf("%s -> %s;", a, b )
                cmd := exec.Command("echo", str)
                cmd.Stdout = file
                err := cmd.Run()
                if err != nil {
                    log.Fatal(err)
                }
            }
        }
    }
    cmd = exec.Command("echo", "}")
    cmd.Stdout = file
    err = cmd.Run()
    if err != nil {
        log.Fatal(err)
    }

    front := regexp.MustCompile("http(.*)://")
    head1 := head(front.ReplaceAllString(*urlFlag, ""))

    if *verbose {
        print(map1[head1], map1, "")
    }
}

//func defluff(s string) string {
//    noDots := regexp.MustCompile("[&%=?#.-]")
//    noNum := regexp.MustCompile("^([0-9])")
//
//    ret := noDots.ReplaceAllString(s, "_")
//    if noNum.Match([]byte(s)) {
//        ret = "_" + s
//    }
//
//    return ret
//}

// Create site structure from list of all internal site links
func levels(pages []string) map[string]level {
    front := regexp.MustCompile("http(.*)://")
    pages = Map(pages, func(url string) string {return front.ReplaceAllString(url, "")})

    levels := make(map[string]level)
    for len(pages) != 0 {
        page := pop(&pages)
        if (page == "") {
            continue
        }

        base := head(page)
        head := head(tail(page))

        if link, ok := levels[base]; ok {
            if (head != "" && head != " ") {
                link.Children = append(link.Children, head)
                levels[head] = level{Name: head}
                levels[base] = link
            }
        } else {
            levels[base] = level{Name: base}
        }
        page = tail(page)
        if page != "" {
            pages = append(pages, page)
        }
    }
    return levels
}

// recursively follow a page's links
func bfs(urlStr string, maxDepth int) []string {
    seen := make(map[string]struct{})
    var q map[string]struct{}
    nq := map[string]struct{}{
        urlStr: struct{}{},
    }

    for i := 0; i < maxDepth; i++ {
        q, nq = nq, make(map[string]struct{})
        for url, _ := range q {
            if _, ok := seen[url]; ok {
                continue
            }
            seen[url] = struct{}{}
            
            for _, link := range get(url) {
                nq[link] = struct{}{}
            }
        }
    }
    ret := make([]string, 0, len(seen))
    for url, _ := range seen {
        ret = append(ret, url)
    }
    // return ret
    return ret

}

// GET a link
func get(urlStr string) []string {

    resp, err := http.Get(urlStr)
    if err != nil {
        return []string{}
    }

    defer resp.Body.Close()

    reqUrl := resp.Request.URL
    baseUrl := &url.URL{
        Scheme: reqUrl.Scheme,
        Host:   reqUrl.Host,
    }
    base := baseUrl.String()
    return toSet(filter(hrefs(resp.Body, base), withPrefix(base)))
}

// Parse HTML for links
func hrefs(r io.Reader, base string) []string {
    links, _ := htmlParser.Parse(r)
    var ret []string
    for _, l := range links {
        switch{
        case strings.HasPrefix(l.Href, "/"):
            ret = append (ret, base+l.Href)
        case strings.HasPrefix(l.Href, "http"):
            ret = append(ret, l.Href)
        }
    }
    return ret
}

// filter a list of strings by a HOF
func filter(links []string, valid func(string) bool) []string {
    var ret []string
    for _, link := range links {
        if valid(link) {
            ret = append(ret, link)
        }
    }
    return ret
}

// recursivley print the hierarchy of all pages. Still needs work
func print(head level, dict map[string]level, space string) {
    fmt.Println(space + head.Name)
    for _, child := range head.Children {
        if child == "" || child == " " || child == head.Name {
            continue
        }
        if _, ok := dict[child]; ok {
            print(dict[child], dict, "  "+space)
        } else {
            fmt.Println(space+"  "+child)
        }
    }
}

func withPrefix(pfx string) func(string) bool {
    return func(link string) bool {
        return strings.HasPrefix(link, pfx) && !strings.HasSuffix(link, "/")
    }
}

func head(input string) string {
    re := regexp.MustCompile("/.*")
    return re.ReplaceAllString(input, "")
}

func tail(input string) string {
    re := regexp.MustCompile("^(.*?)/")
    if !(strings.Contains(input, "/")) {
        return ""
    }
    return re.ReplaceAllString(input, "")
}

func toSet(list []string) []string {
    slices.Sort(list)
    return slices.Compact(list)
}

func convertSet(list []string) []string {
    seen := make(map[string]struct{})
    ret := []string{}
    for _, item := range list {
        if _, ok := seen[item]; ok { 
            continue
        } else {
            seen[item] = struct{}{}
            ret = append(ret, item)
        }
    }
    return ret
}
	
func Map(vs []string, f func(string) string) []string {
    vsm := make([]string, len(vs))
    for i, v := range vs {
        vsm[i] = f(v)
    }
    return vsm
}

func pop(alist *[]string) string {
   rv:=(*alist)[0]
   *alist=(*alist)[1:]
   return rv
}
