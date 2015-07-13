package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
)

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options] [PATH]\n", os.Args[0])
	flag.CommandLine.VisitAll(func(flag *flag.Flag) {
		fmt.Fprintf(os.Stderr, " -%s\t%s, default %s\n",
			flag.Name, flag.Usage, flag.DefValue)
	})
}

func parseArgs(port *int, path *string) {
	flag.Usage = printUsage
	flag.IntVar(port, "p", 3232, "Port to listen on")
	flag.Parse()

	if len(flag.Args()) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatalln("Failed to get current working directory")
		}
		*path = cwd
	} else {
		*path = flag.Args()[0]
	}
}

func logHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s from %s", r.Method, r.URL, r.RemoteAddr)
		handler.ServeHTTP(w, r)
	})
}

func getAddr(port int) string {
	return fmt.Sprintf(":%d", port)
}

func getHandler(src string) (http.Handler, error) {
	fstat, err := os.Stat(src)
	if os.IsNotExist(err) {
		return nil, err
	}
	var h http.Handler
	switch mode := fstat.Mode(); {
	case mode.IsDir():
		h = http.FileServer(http.Dir(src))
	case mode.IsRegular():
		h = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", "application/octet-stream")
			w.Header().Set("content-disposition", "filename="+path.Base(src))
			f, err := os.Open(src)
			if err != nil {
				log.Fatalf("cannot open file: %s", src)
			}
			defer f.Close()
			io.Copy(w, f)
		})
	}
	return h, nil
}

func main() {
	var port int
	var path string

	parseArgs(&port, &path)

	log.Printf("try to bind to 0.0.0.0:%d", port)
	log.Printf("the shared path is %s", path)

	handler, err := getHandler(path)
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/", logHandler(handler))
	log.Fatal(http.ListenAndServe(getAddr(port), nil))
}
