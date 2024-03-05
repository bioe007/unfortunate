package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const (
	dataPath                   = "/usr/share/fortune/wisdom"
	cachePath                  = "./unfortunate.cache"
	delim                      = "%\n"
	cacheOffsetPath            = 0
	cacheOffsetPathLength      = 5
	cacheOffsetFortunesEntries = 10
	cacheOffsetNumFortunes     = 14
)

type fortuneEntry struct {
	offset int64
	length int32
}

type fortuneCache struct {
	path     string
	fortunes []fortuneEntry
}

// fields in binary file layout -
// 0..4 - int32 location of path name
// 5..8 - int32 length of path name
// 9 - reserved
// 10..13 - int32 location of start of fortune entries
// 14..17 - int32 number of fortune entries
// +content as described
// [] string of pathlen length
// [] array of fortuneentries
type cacheLayout struct {
	pathOffset         int32
	pathLenOffset      int32
	fortuneStartOffset int32
	fortuneCountOffset int32
}

func NewCacheLayout() cacheLayout {
	cl := new(cacheLayout)
	return *cl
}

func main() {
	if _, err := os.Stat(cachePath); err == nil {
		fmt.Println("cache is ready do something")
	} else if errors.Is(err, os.ErrNotExist) {
		log.Fatal("No cache file found! building dataset.")
		err = buildFortuneCache(dataPath)
		if err != nil {
			log.Fatal("Still coudln't build cache..", err)
		}
		print("try again...")
		os.Exit(1)
	} else {
		log.Fatal("fixme: unhandled erorr checking for cache, exiting.", err)
	}

	fcache, err := os.Open(cachePath)
	if err != nil {
		log.Fatal("can't open cache", err)
	}
	defer fcache.Close()

	// f := getFortune()
	// fmt.Printf(f)
}

func buildFortuneCache(fpath string) error {
	f, err := os.Open(fpath)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	fc := new(fortuneCache)
	fc.path = fpath
	tmp_ftn := strings.Builder{}
	for scanner.Scan() {
		line := scanner.Text()
		if line == "%\n" {
			pos, err := f.Seek(0, io.SeekCurrent)
			if err != nil {
				return err
			}
			fe := fortuneEntry{
				pos,
				int32(tmp_ftn.Len()),
			}
			fc.fortunes = append(fc.fortunes, fe)
		} else {
			tmp_ftn.WriteString(line)
		}
	}

	fbin, err := os.Create(cachePath)
	if err != nil {
		log.Fatal(err)
	}
	defer fbin.Close()

	return nil
}
