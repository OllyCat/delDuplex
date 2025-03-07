// Package main
package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type md5File struct {
	name    string
	md5     string
	modTime time.Time
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Using:\n\n%s <dir name> [<dir name> ...]\n", os.Args[0])
	}

	flag.Parse()
	if len(flag.Args()) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	var (
		wg    sync.WaitGroup
		ch    = make(chan md5File)
		done  = make(chan bool)
		limit = make(chan struct{}, runtime.NumCPU())
		files = make(map[string][]md5File)
	)

	go addFile(files, ch, done)

	for _, i := range flag.Args() {
		err := filepath.Walk(i, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				limit <- struct{}{}
				wg.Add(1)
				go md5sum(path, ch, limit, &wg)
			}
			return nil
		})
		if err != nil {
			log.Printf("ERROR: incorrect path: %v\n", i)
		}
	}

	wg.Wait()
	done <- true
	delDup(files)
}

func delDup(files map[string][]md5File) {
	for _, f := range files {
		if len(f) > 1 {
			v := sort(f)
			fmt.Printf("Trying delete duplicate of file:\t%v\n", v[0].name)
			for _, i := range v[1:] {
				if err := os.Remove(i.name); err != nil {
					log.Printf("\tERROR: delete error: %v\n", err)
				} else {
					fmt.Printf("\tFile deleted:\t%v\n", i.name)
				}
			}
		}
	}
}

func sort(f []md5File) []md5File {
	first := f[0].modTime
	for i := 1; i < len(f); i++ {
		if f[i].modTime.Before(first) {
			first = f[i].modTime
			f[0], f[i] = f[i], f[0]
		}
	}
	return f
}

func addFile(files map[string][]md5File, ch chan md5File, done chan bool) {
	for {
		select {
		case <-done:
			close(ch)
			close(done)
			return
		case f, ok := <-ch:
			if ok {
				files[f.md5] = append(files[f.md5], f)
			}
		}
	}
}

func md5sum(name string, ch chan md5File, limit <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		<-limit
	}()

	s, err := os.Lstat(name)
	if err != nil {
		log.Printf("ERROR: File %v stat error: %v\n", name, err)
		return
	}

	f, err := os.Open(name)
	if err != nil {
		log.Printf("ERROR: File %v opening error: %v\n", name, err)
		return
	}

	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Printf("ERROR: Sum error: %v\n", err)
		return
	}
	res := md5File{name, fmt.Sprintf("%x", h.Sum(nil)), s.ModTime()}
	ch <- res
}
