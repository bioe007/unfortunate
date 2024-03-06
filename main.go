package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
)

const (
	// dataPath                   = "/usr/share/fortune/wisdom"
	dataPath                 = "./fakefortune.txt"
	cachePath                = "./unfortunate.cache"
	delim                    = "%\n"
	cacheOffsetNumFortunes   = 0
	cacheOffsetPathLength    = 4
	cacheOffsetPath          = 8
	cacheOffsetFortunesStart = -1 // define this when writing by looking at pathlength
)

type fortuneEntry struct {
	offset int64
	length int32
}

type fortuneCache struct {
	path     string
	fortunes []fortuneEntry
}

// TODO - this would be a 'fancy' version of the cache
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
	fortuneCountOffset int32
	pathLenOffset      int32
	pathOffset         int32
	fortuneStartOffset int32
}

// TODO - iff fancy version
// Just a helper to keep file offsets
func newCacheLayout(fc fortuneCache) cacheLayout {
	cl := new(cacheLayout)
	cl.pathOffset = cacheOffsetPath
	cl.pathLenOffset = cacheOffsetPathLength
	cl.fortuneCountOffset = cacheOffsetNumFortunes
	cl.fortuneStartOffset = cacheOffsetPath + int32(len(fc.path))
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
		log.Fatal("fixme: unhandled error checking for cache, exiting.", err)
	}

	fcache, err := os.Open(cachePath)
	if err != nil {
		log.Fatal("can't open cache", err)
	}
	defer fcache.Close()
	fortune_num := rand.Intn(getFortuneCountFromCache())
	fmt.Println("fortune is: ", getFortuneByIndex(fortune_num))
	fmt.Println("-----------------------------------------------------")
	fmt.Println("fortune is: ", getFortuneByIndex(1))
	fmt.Println("fortune is: ", getFortuneByIndex(2))
	fmt.Println("fortune is: ", getFortuneByIndex(3))
	fmt.Println("fortune is: ", getFortuneByIndex(4))
	fmt.Println("fortune is: ", getFortuneByIndex(5))
	fmt.Println("fortune is: ", getFortuneByIndex(6))
	fmt.Println("fortune is: ", getFortuneByIndex(7))
}

func getFortuneByIndex(idx int) string {
	// TODO - sizeof(struct) in go isn't a thing
	// skip the num_fortunes value and scale by sizeof(fortuneentry)
	idxtocache := 8 + (idx-1)*(8+4)
	var filestartidx int64
	cache, err := os.Open(cachePath)
	if err != nil {
		log.Fatal("couldn't open cache file", err)
	}
	defer cache.Close()

	cache.Seek(int64(idxtocache), 0)
	binary.Read(cache, binary.LittleEndian, &filestartidx)

	f, err := os.Open(dataPath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// debugging
	fmt.Printf("fortune num: %d \tidxtocache: %d\toffset: %d\n", idx, idxtocache, filestartidx)

	f.Seek(int64(filestartidx)-2, 0)
	scanner := bufio.NewScanner(f)
	var s strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		// fmt.Println("line", line)
		if line == "%" {
			// fmt.Println("found fortune", s.String(), line)
			break // fortune terminator TODO - this should be constant
		} else {
			s.WriteString(line)
			s.WriteString("\n")
		}
	}

	return s.String()
}

func writeCache(fc *fortuneCache) {
	fmt.Println("creating cache file: ", cachePath) // debugging
	cachefile, err := os.Create(cachePath)
	if err != nil {
		log.Fatal(err)
	}
	defer cachefile.Close()

	err = binary.Write(cachefile, binary.LittleEndian, int64(len(fc.fortunes)))
	if err != nil {
		log.Fatal("Failed writing fortune counts: ", err)
	}

	for i, v := range fc.fortunes {
		// debugging
		fmt.Printf(
			"Writing fortune: index=%d\toffset=%d\tlen=%d\tsize=%d\n",
			i,
			v.offset,
			v.length,
			binary.Size(v),
		)
		err = binary.Write(cachefile, binary.LittleEndian, v)
		if err != nil {
			log.Fatal("Failed writing fortune entry: ", fc, "\t", err)
		}
	}
}

func getFortuneCountFromCache() int {
	cachefile, err := os.Open(cachePath)
	if errors.Is(err, os.ErrNotExist) {
		log.Fatal("No cache file at %s", cachePath, err)
	}

	var num_fortunes int64
	binary.Read(cachefile, binary.LittleEndian, &num_fortunes)
	fmt.Println("dbug num_fortunes", num_fortunes)

	return int(num_fortunes)
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

	writeCache(fc)
	return nil
}
