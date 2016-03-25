package util

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type ByBasename []string

func (a ByBasename) Len() int           { return len(a) }
func (a ByBasename) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByBasename) Less(i, j int) bool { return filepath.Base(a[i]) < filepath.Base(a[j]) }

type ByContent []string

func (a ByContent) Len() int           { return len(a) }
func (a ByContent) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByContent) Less(i, j int) bool { return a[i] < a[j] }

func TSStr(ts time.Time) string {
	return ts.Format("20060102_150405")
}

func Safe_Realm(realm string) string {
	return strings.Replace(realm, ":", "-", -1)
}

func Make_FName(realm string, ts time.Time, zipped bool) string {
	n := fmt.Sprintf("%s-%s.json", Safe_Realm(realm), TSStr(ts.UTC()))
	if zipped {
		n = n + ".gz"
	}
	return n
}

func Parse_FName(fname string) (realm string, ts time.Time, good bool) {
	good = false
	// log.Printf("Parse_FName(%s)", fname)
	name := filepath.Base(fname)
	rx := regexp.MustCompile("^([^-]+)-([^-]+)-(\\d{8}_\\d{6})\\.json\\.gz$")
	v := rx.FindStringSubmatch(name)
	if v == nil {
		//log.Panicf("... not matched")
		return
	}
	// log.Printf("... matched, v=%v", v)
	realm = v[1] + ":" + v[2]
	ts, err := time.Parse("20060102_150405", v[3])
	if err != nil {
		//log.Panicf("time not parsed: %s", v[3])
		return
	}
	good = true
	return
}

// получить полный путь до исполняемого файла
func ExeName() string {
	exe, err := filepath.Abs(os.Args[0])
	if err != nil {
		log.Fatal(err)
	}
	return exe
}

// получить каталог приложения
func AppDir() string {
	dir, err := filepath.Abs(filepath.Dir(ExeName()))
	if err != nil {
		log.Fatal(err)
	}
	return dir
}

func AppBaseFileName() string { 
/* 
name with full pathname but without extension
c:\Apps\MyFile.exe -> c:\Apps\MyFile
/tmp/zzzz -> /tmp/zzz
*/
	r, _ := regexp.Compile("^(.*?)(?:\\.exe|\\.EXE|)$")
	return r.FindStringSubmatch(ExeName())[1]
}

// проверить или создать каталог
func CheckDir(path string) {
	log.Println("check for directory: ", path)
	if err := os.MkdirAll(path, 0755); err != nil {
		log.Fatalf("mkDirAll(%s) got error: %s", path, err)
	}
}

// Проверить на наличие файла. Если это - не файл - подохнуть
func CheckFile(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	if info.IsDir() {
		log.Fatalln("not a flie: ", path)
	}
	return true
}

// Проверить возможность парсинга JSON-блока
func CheckJSON(data []byte) {
	var r interface{}
	if err := json.Unmarshal(data, &r); err != nil {
		log.Fatal("broken")
	}
	log.Print("=== postmortem ===  %f", data)
}

// сжать данные gzip-ом
func Zip(data []byte) []byte {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write(data)
	gz.Close()
	return buf.Bytes()
}

// распаковать gzip-данные
func Unzip(zdata []byte) []byte {
	z := bytes.NewReader(zdata)
	zreader, err := gzip.NewReader(z)
	if err != nil {
		log.Fatalf(".. create gzip reader failed: %s", err)
	}
	defer zreader.Close()
	ubody, err := ioutil.ReadAll(zreader)
	if err != nil {
		log.Fatalf(".. gunzip failed: %s", err)
	}
	return ubody
}

// загрузить (и распаковать) данные
func LoadData(zdata []byte) []byte {
	z := bytes.NewReader(zdata)
	zreader, err := gzip.NewReader(z)
	if err != nil {
		log.Fatalf(".. create gzip reader failed: %s", err)
	}
	defer zreader.Close()
	ubody, err := ioutil.ReadAll(zreader)
	if err != nil {
		log.Fatalf(".. gunzip failed: %s", err)
	}
	return ubody
}

func Store(fname string, data []byte) error {
	return ioutil.WriteFile(fname, data, 0644)
}

func Load(fname string) (data []byte, err error) {

	if gzipped, _ := regexp.MatchString("\\.gz$", fname); gzipped { // gunzip it
		fi, err := os.Open(fname)
		if err != nil {
			return nil, err
		}
		defer fi.Close()

		fz, err := gzip.NewReader(fi)
		if err != nil {
			return nil, err
		}
		defer fz.Close()

		s, err := ioutil.ReadAll(fz)
		if err != nil {
			return nil, err
		}
		return s, nil
	}
	return ioutil.ReadFile(fname)
}
