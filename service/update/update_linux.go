package update

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func download(tarGzURL string, timestamp time.Time) (string, error) {
	tarGzFile := fmt.Sprintf("cprobe-%s.tar.gz", timestamp.Format("2006-01-02-15-04-05"))
	fmt.Println("downloading file:", tarGzURL, "save to:", tarGzFile)

	res, err := http.Get(tarGzURL)
	if err != nil {
		return "", fmt.Errorf("failed to download file from %s, error: %v", tarGzURL, err)
	}

	if res.Body == nil {
		return "", fmt.Errorf("downloading file %s error: response body is nil", tarGzURL)
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", fmt.Errorf("downloading %s unexpected status code: %s", tarGzURL, res.Status)
	}

	tarGzFileFD, err := os.OpenFile(tarGzFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return "", err
	}

	defer tarGzFileFD.Close()
	bufWriter := bufio.NewWriter(tarGzFileFD)

	_, err = io.Copy(bufWriter, res.Body)
	if err != nil {
		return "", err
	}

	bufWriter.Flush()
	return tarGzFile, nil
}

func Update(updateFile string) (err error) {
	tarGzFile := updateFile
	timestamp := time.Now()

	if strings.HasPrefix(updateFile, "http://") || strings.HasPrefix(updateFile, "https://") {
		tarGzFile, err = download(updateFile, timestamp)
		if err != nil {
			return
		}
	}

	dstDir := fmt.Sprintf(".%s", timestamp.Format("2006-01-02-15-04-05"))

	nv, err := UnTar(dstDir, tarGzFile)
	if err != nil {
		return err
	}

	if nv == "" {
		return fmt.Errorf("can not find new version binary")
	}

	// old version
	ov, err := os.Executable()
	if err != nil {
		return err
	}

	oldFD, err := os.Stat(ov)
	if err != nil {
		return err
	}

	newFD, err := os.Stat(nv)
	if err != nil {
		return err
	}

	if newFD.Mode().IsDir() {
		return fmt.Errorf("%s is directory", nv)
	}

	fmt.Printf("replace old version: %s with new version: %s\n", ov, nv)

	// replace
	err = os.Rename(nv, ov)
	if err != nil {
		return err
	}

	return os.Chmod(ov, oldFD.Mode().Perm())
}

func UnTar(dst, src string) (target string, err error) {
	fr, err := os.Open(src)
	if err != nil {
		return
	}
	defer fr.Close()

	gr, err := gzip.NewReader(fr)
	if err != nil {
		return
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	for {
		hdr, err := tr.Next()

		switch {
		case err == io.EOF:
			return target, nil
		case err != nil:
			return target, err
		case hdr == nil:
			continue
		}

		dstFileDir := filepath.Join(dst, hdr.Name)

		switch hdr.Typeflag {
		case tar.TypeDir:
			if b := ExistDir(dstFileDir); !b {
				if err := os.MkdirAll(dstFileDir, 0775); err != nil {
					return target, err
				}
			}
		case tar.TypeReg:
			err := os.MkdirAll(filepath.Dir(dstFileDir), 0755)
			if err != nil {
				return target, fmt.Errorf("mkdir:%s, error:%s", filepath.Base(dstFileDir), err)
			}

			file, err := os.OpenFile(dstFileDir, os.O_CREATE|os.O_RDWR, os.FileMode(hdr.Mode))
			if err != nil {
				return target, err
			}

			if strings.HasSuffix(dstFileDir, "cprobe") {
				target = dstFileDir
			}

			_, err = io.Copy(file, tr)
			if err != nil {
				return target, err
			}

			file.Close()
		}
	}

	return target, nil
}

func ExistDir(dirname string) bool {
	fi, err := os.Stat(dirname)
	return (err == nil || os.IsExist(err)) && fi.IsDir()
}
