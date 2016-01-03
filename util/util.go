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
	"strings"
	"time"
)

func TSStr(ts time.Time) string {
	return ts.Format("20060102_150405")
}

func Make_FName(realm string, ts time.Time) string {
	v := strings.Split(realm, ":")
	return fmt.Sprintf("%s-%s-%s.json.gz", v[0], v[1], TSStr(ts.UTC()))
}

// получить каталог приложения
func AppDir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return dir
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
func Unzip(data []byte) []byte {
	/*
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		gz.Write(data)
		gz.Close()
		return buf.Bytes()
		reader, err = gzip.NewReader(data)
		if err != nil {
			log.Fatalf(".. create gzip reader failed: %s", url, err)
		}
		defer reader.Close()
		ubody, err = ioutil.ReadAll(reader)
		if err != nil {
			log.Fatalf(".. gunzip failed: %s", url, err)
		}
		return ubody
	*/
	return data // FIXME
}

func Store(fname string, data []byte) error {
	return ioutil.WriteFile(fname, data, 0644)
}

func Load(fname string) (data []byte, err error) {
	return ioutil.ReadFile(fname)
}
