package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

const (
	// dataPath                   = "/usr/share/fortune/wisdom"
	dataPath                   = "./fakefortune.txt"
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
		fmt.Println("No cache file found! building dataset.")
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

func writeCache(fc *fortuneCache) error {
	cachefile, err := os.Open(fc.path)
	if err != nil {
		log.Fatal(err)
	}

	buf := new(bytes.Buffer)
	fmt.Println("writing to buf")
	err = binary.Write(buf, binary.BigEndian, int8(42))
	fmt.Println("written to buf")
	if err != nil {
		log.Fatal("deadbeef :", err)
	}
	fmt.Println("writing to buf")
	err = binary.Write(buf, binary.BigEndian, fc.fortunes[0])
	fmt.Println("written to buf")
	if err != nil {
		log.Fatal("fc.fortunes", err)
	}
	fmt.Println("writing to cachefile")
	cachefile.Write(buf.Bytes())

	return nil
}

func buildFortuneCache(fpath string) error {
	f, err := os.Open(fpath)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewReader(f)
	fc := new(fortuneCache)
	fc.path = fpath
	line, err := scanner.ReadBytes('\n')
	if err != nil {
		log.Fatal(err)
	}
	tmp_ftn := new(bytes.Buffer)
	ftn_len := int32(0)
	for len(line) != 0 {
		if string(line) == "%\n" {
			pos, err := f.Seek(0, io.SeekCurrent)
			// TODO - this is not what you think it is...
			fmt.Printf("fortune term at position: %d\t", pos)
			if err != nil {
				return err
			}
			fe := fortuneEntry{
				pos,
				ftn_len,
			}
			fmt.Printf(
				"added fortune %+x - offset=%d, length=%d  -> %s\n",
				fe,
				fe.offset,
				fe.length,
				tmp_ftn.String(),
			)
			fc.fortunes = append(fc.fortunes, fe)
			tmp_ftn.Reset()
		} else {
			tmp_ftn.Write(line)
			ftn_len += int32(len(line))
		}
		line, err = scanner.ReadBytes('\n')
		if err != nil {
			log.Fatal(err)
		}
	}
	fmt.Printf("fc cache %+x\n", fc)

	err = writeCache(fc)
	if err != nil {
		return err
	}
	return nil
}
