package utilsW

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
)

func TarGz(outName string, srcNames []string) error {
	out, err := os.Create(outName)
	if err != nil {
		return err
	}
	defer out.Close()
	gw := gzip.NewWriter(out)
	tw := tar.NewWriter(gw)

	defer gw.Close()
	defer tw.Close()

	for _, filename := range srcNames {
		info, err := os.Stat(filename)
		if err != nil {
			return err
		}
		th, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}
		if err = tw.WriteHeader(th); err != nil {
			return err
		}
		src, err := os.Open(filename)
		if err != nil {
			return err
		}
		if _, err = io.Copy(tw, src); err != nil {
			return err
		}
	}
	return nil
}
