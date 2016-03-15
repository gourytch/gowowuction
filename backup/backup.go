package backup

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	util "github.com/gourytch/gowowuction/util"
)

func MakeTarball(tarname string, fnames []string) error {
	log.Printf("tarring %d entrires to %s ...", len(fnames), tarname)
	tarfile, err := os.Create(tarname)
	if err != nil {
		return err
	}
	defer tarfile.Close()
	var tarwriter *tar.Writer
	if strings.HasSuffix(tarname, ".gz") {
		zipper := gzip.NewWriter(tarfile)
		defer zipper.Close()
		tarwriter = tar.NewWriter(zipper)
		/*
			} else if strings.HasSuffix(tarname, ".xz") {
				p := xz.WriterDefaults
				p.DictCap = 1 << 24
				zipper, err := xz.NewWriterParams(tarfile, &p) //xz.NewWriter(tarfile)
				if err != nil {
					return err
				}
				defer zipper.Close()
				tarwriter = tar.NewWriter(zipper)
		*/
	} else {
		tarwriter = tar.NewWriter(tarfile)
	}
	defer tarwriter.Close()

	for _, fname := range fnames {
		realm, ts, good := util.Parse_FName(fname)
		if !good {
			log.Printf("warning: skip ill-named file '%s'", fname)
			continue // skip
		}
		data, err := util.Load(fname)
		if err != nil {
			return err
		}

		hdr := new(tar.Header)
		hdr.Name = util.Make_FName(realm, ts, false)
		hdr.Size = int64(len(data))
		hdr.ModTime = ts
		hdr.Mode = 0644
		err = tarwriter.WriteHeader(hdr)
		if err != nil {
			return err
		}
		log.Printf("tar %d bytes for file %s", hdr.Size, hdr.Name)
		_, err = tarwriter.Write(data)
		if err != nil {
			return err
		}
	}
	log.Printf("%s tarred without errors", tarname)
	return nil
}

func MakeZip(zipname string, fnames []string) error {
	log.Printf("zipping %d entrires to %s ...", len(fnames), zipname)
	zipfile, err := os.Create(zipname)
	if err != nil {
		return err
	}
	defer zipfile.Close()
	zipwriter := zip.NewWriter(zipfile)
	defer zipwriter.Close()

	for _, fname := range fnames {
		realm, ts, good := util.Parse_FName(fname)
		if !good {
			log.Printf("warning: skip ill-named file '%s'", fname)
			continue // skip
		}
		data, err := util.Load(fname)
		if err != nil {
			return err
		}
		name := util.Make_FName(realm, ts, false)
		f, err := zipwriter.Create(name)
		if err != nil {
			return err
		}
		log.Printf("zip %d bytes for file %s", len(data), name)
		_, err = f.Write(data)
		if err != nil {
			return err
		}
	}
	log.Printf("%s zipped without errors", zipname)
	return nil
}

func Backup(srcdir, dstdir, timeformat, ext string) {
	// Backup("/opt/wowauc/download", "/opt/wowauc/backup", "20060102", ".tar.gz")
	fnames, err := filepath.Glob(srcdir + "/*.json.gz")
	if err != nil {
		log.Fatalln("glob failed:", err)
	}
	log.Printf("... %d entries collected", len(fnames))

	rmap := make(map[string][]string)

	for _, fname := range fnames {
		// realm, ts, good := util.Parse_FName(fname)
		realm, ts, good := util.Parse_FName(fname)
		if good {
			// log.Printf("fname %s -> %s, %v", fname, realm, ts)
			key := strings.Replace(realm, ":", "-", -1) + "-" + ts.Format(timeformat)
			rmap[key] = append(rmap[key], fname)
		} else {
			// log.Printf("skip fname %s", fname)
		}
	}

	var keys []string
	for key, _ := range rmap {
		keys = append(keys, key)
	}
	sort.Sort(util.ByContent(keys))

	for _, key := range keys {
		fnames := rmap[key]
		log.Printf("create tarball for %s with %d entrires ", key, len(fnames))
		sort.Sort(util.ByBasename(fnames))
		if ext == ".tar.gz" {
			tarname := dstdir + "/" + key + ext
			err := MakeTarball(tarname, fnames)
			if err != nil {
				log.Fatalf("MakeTarball(%s) failed: %s", tarname, err)
			}
		} else if ext == ".zip" {
			zipname := dstdir + "/" + key + ext
			err := MakeZip(zipname, fnames)
			if err != nil {
				log.Fatalf("MakeZip(%s) failed: %s", zipname, err)
			}
		}
	}
	return
}
