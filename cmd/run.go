package main

import (
    "flag"
    "html/template"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "path"
    "strings"
)

type Linkable struct {
    Link string
}

type Dir struct {
    Linkable
}

type Picture struct {
    Linkable
}

type ListData struct {
    Name string
    Pictures []Picture
    Dirs []Dir
}

const defaultGalleryPath = "/home/mvala/Pictures"
var supportedFiles = []string{".jpg", ".jpeg", ".png"}

type Conf struct {
    galleryPath string
}

func main() {
    var conf = &Conf{}
    flag.StringVar(&conf.galleryPath, "photosdir", defaultGalleryPath, "default path to the gallery")
    flag.Parse()

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        //fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
        if strings.HasPrefix(r.URL.Path, conf.galleryPath) && isSupportedFile(r.URL.Path) {
            if content, err := ioutil.ReadFile(r.URL.Path); err == nil {
                if _, wErr := w.Write(content); wErr != nil {
                    log.Println(wErr)
                }
            }
        } else {
            reqPath := conf.galleryPath + r.URL.Path
            reqPath = strings.TrimRight(reqPath, "/")
            if fi, err := os.Stat(reqPath); err != nil && fi != nil {
                if fi.IsDir() {
                    renderTemplate(w, r, conf.galleryPath)
                }
            } else {
                log.Println(reqPath, "is not dir")
            }
        }
    })

    log.Fatal(http.ListenAndServe(":8080", nil))
}

func listFiles(dirPath string) ([]Dir, []Picture) {

    files, err := ioutil.ReadDir(dirPath)
    if err != nil {
        log.Fatal(err)
    }

    pictures := make([]Picture, 0)
    dirs := make([]Dir, 0)
    for _, file := range files {
        if file.IsDir() {
            dirs = append(dirs, Dir{Linkable{Link: file.Name()}})
        } else {
            if isSupportedFile(file.Name()) {
                pictures = append(pictures, Picture{Linkable{
                    Link: path.Join(dirPath, file.Name())},
                })
            }
        }
    }
    return dirs, pictures
}

func isSupportedFile(filename string) bool {
    for _, ext := range supportedFiles {
        if strings.HasSuffix(filename, ext) {
            return true
        }
    }
    return false
}

func renderTemplate(w http.ResponseWriter, r *http.Request, galleryPath string) {
    parsedTemplate, _ := template.ParseFiles("templates/index.html")
    dirs, pictures := listFiles(galleryPath + r.URL.Path)
    err := parsedTemplate.Execute(w, ListData{Name: galleryPath, Dirs: dirs, Pictures: pictures})
    if err != nil {
        log.Println("Error executing template :", err)
        return
    }
}
