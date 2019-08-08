package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type Md5File struct {
	name string
	md5  string
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		usage(os.Args[0])
		os.Exit(1)
	}

	var (
		wg    sync.WaitGroup
		ch    = make(chan Md5File)
		done  = make(chan bool)
		files = make(map[string][]string)
	)

	go addFile(files, ch, done)

	for _, i := range args {
		err := filepath.Walk(i, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				wg.Add(1)
				go md5sum(path, ch, &wg)
			}
			return nil
		})
		if err != nil {
			log.Printf("Ошибка чтения пути: %v\n", i)
		}
	}

	wg.Wait()
	done <- true
	for k, v := range files {
		fmt.Printf("Key = %v, valume = %#v\n", k, v)
	}
	delDup(files)
}

func delDup(files map[string][]string) {
}

func usage(prog string) {
	fmt.Printf("Using:\n\n%s <dir name> [<dir name> ...]\n", prog)
}

func addFile(files map[string][]string, ch chan Md5File, done chan bool) {
	for {
		select {
		case <-done:
			close(ch)
			close(done)
			return
		case f, ok := <-ch:
			if ok {
				files[f.md5] = append(files[f.md5], f.name)
			}
		}
	}
}

func md5sum(name string, ch chan Md5File, wg *sync.WaitGroup) {
	defer wg.Done()
	f, err := os.Open(name)
	if err != nil {
		log.Printf("File %v open is filed.\n", name)
		return
	}
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Printf("Sum error: %v\n", err)
		return
	}
	res := Md5File{name, fmt.Sprintf("%x", h.Sum(nil))}
	ch <- res
}
