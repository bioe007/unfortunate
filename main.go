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
	fmt.Println("creating file", cachePath)
	cachefile, err := os.Create(cachePath)
	if err != nil {
		log.Fatal(err)
	}
	defer cachefile.Close()

	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.BigEndian, int8(42))
	if err != nil {
		log.Fatal("deadbeef :", err)
	}

	fmt.Println("how many entries?", len(fc.fortunes))

	err = binary.Write(buf, binary.BigEndian, fc.fortunes[0])
	if err != nil {
		log.Fatal("fc.fortunes", err)
	}
	fmt.Println("writing to cachefile", len(buf.Bytes()))
	cachefile.Write(buf.Bytes())

	return nil
}

func buildFortuneCache(fpath string) error {
	f, err := os.Open(fpath)
	if err != nil {
		return err
	}
	defer f.Close()

	// TODO: Reader isn't really much better as it also can't track
	// the file position in a helpful way. Might as well switch
	// back to a scanner
	scanner := bufio.NewReader(f)

	fc := new(fortuneCache)
	fc.path = fpath

	// Have to count bytes read because the buffered reader can't
	// track the position of what's been read from the buffer only
	// the actual file pointer which will always be at the bufsize
	// boundary. see f.Seek(0, io.SeekCurrent))
	byte_count := int64(0)       // number of bytes read from file
	ftn_len := int32(0)          // length of current fortune
	tmp_ftn := new(bytes.Buffer) // debugging
	for {
		// We need to know how many bytes into the file for each fortune.
		line, err := scanner.ReadBytes('\n')
		if err == io.EOF {
			// Without a terminator the tail of the file is just considered as comments.
			break
		} else if err != nil {
			log.Fatal(err)
		}
		byte_count += int64(len(line)) // increment byte count

		// NOTE: debugging - buffered read doesn't show where e.g. Text() is at
		// pos, err := f.Seek(0, io.SeekCurrent)

		if string(line) == "%\n" {
			fmt.Printf("ftn terminated at byte_count: %d\t", byte_count)
			if err != nil {
				return err
			}

			fe := fortuneEntry{
				// backup location by length of the current fortune
				byte_count - int64(ftn_len),
				ftn_len,
			}

			// debugging
			fmt.Printf(
				"added fortune %+x - offset=%d, length=%d  -> %s",
				fe,
				fe.offset,
				fe.length,
				tmp_ftn.String(),
			)
			fc.fortunes = append(fc.fortunes, fe)

			// reset trackers
			tmp_ftn.Reset()
			ftn_len = 0

		} else {
			tmp_ftn.Write(line) // debugging
			ftn_len += int32(len(line))
		}
	}
	fmt.Printf("fc cache %+x\n", fc)

	err = writeCache(fc)
	if err != nil {
		return err
	}
	return nil
}
