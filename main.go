package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"strings"

	"github.com/richardwilkes/toolbox/atexit"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/formats/icon"
	"github.com/richardwilkes/toolbox/formats/icon/icns"
	"github.com/richardwilkes/toolbox/formats/icon/ico"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/taskqueue"
	"github.com/richardwilkes/toolbox/xio"
	"github.com/richardwilkes/toolbox/xio/fs"
)

func main() {
	docImg, err := loadImage("artwork_prep/doc.png")
	jot.FatalIfErr(err)

	f, err := os.Open("artwork_prep/types")
	jot.FatalIfErr(err)
	fis, err := f.Readdir(-1)
	jot.FatalIfErr(err)
	xio.CloseIgnoringErrors(f)
	list := make([]string, 0, len(fis))
	for _, fi := range fis {
		if strings.HasSuffix(fi.Name(), ".png") {
			list = append(list, fs.TrimExtension(fi.Name()))
		}
	}

	for _, dir := range []string{"icns", "ico", "png", "file_associations"} {
		jot.FatalIfErr(os.RemoveAll(dir))
		jot.FatalIfErr(os.MkdirAll(dir, 0755))
	}
	for _, dir := range []string{"linux", "macos", "windows"} {
		jot.FatalIfErr(os.MkdirAll("file_associations/"+dir, 0755))
	}

	tq := taskqueue.New()
	tq.Submit(processAppFileTask)
	for _, name := range list {
		tq.Submit(newProcessDocFileTask(name, docImg))
	}
	tq.Shutdown()

	atexit.Exit(0)
}

func processAppFileTask() {
	img, err := loadImage("artwork_prep/app.png")
	jot.FatalIfErr(err)
	writeIcons("app", img)
}

func newProcessDocFileTask(name string, docImg image.Image) taskqueue.Task {
	return func() {
		img, err := loadImage("artwork_prep/types/" + name + ".png")
		jot.FatalIfErr(err)
		writeIcons(name+"_doc", icon.Stack(docImg, img))
		writeAssociationFile(name, "linux", "png")
		writeAssociationFile(name, "macos", "icns")
		writeAssociationFile(name, "windows", "ico")
	}
}

func writeAssociationFile(name, platform, fileType string) {
	f, err := os.Create("file_associations/" + platform + "/" + name + "_ext.properties")
	jot.FatalIfErr(err)
	_, err = fmt.Fprintf(f, `extension=%[1]s
mime-type=application/gcs.%[1]s
icon=artifacts/%[2]s/%[1]s_doc.%[2]s
description=%[3]s
`, name, fileType, getDescription(name))
	jot.FatalIfErr(err)
	jot.FatalIfErr(f.Close())
}

func getDescription(name string) string {
	switch name {
	case "adm":
		return "GCS Advantage Modifiers Library"
	case "adq":
		return "GCS Advantages Library"
	case "eqm":
		return "GCS Equipment Modifiers Library"
	case "eqp":
		return "GCS Equipment Library"
	case "gcs":
		return "GURPS Character Sheet"
	case "gct":
		return "GCS Character Template"
	case "not":
		return "GCS Notes Library"
	case "skl":
		return "GCS Skills Library"
	case "spl":
		return "GCS Spells Library"
	default:
		return name
	}
}

func writeIcons(name string, img image.Image) {
	imgs := icon.ScaleTo(img, []image.Point{
		image.Pt(1024, 1024),
		image.Pt(512, 512),
		image.Pt(256, 256),
		image.Pt(128, 128),
		image.Pt(64, 64),
		image.Pt(48, 48),
		image.Pt(32, 32),
		image.Pt(16, 16),
	})

	f, err := os.Create("icns/" + name + ".icns")
	jot.FatalIfErr(err)
	jot.FatalIfErr(icns.Encode(f, getImages([]int{1024, 512, 256, 128, 64, 32, 16}, imgs)...))
	jot.FatalIfErr(f.Close())

	f, err = os.Create("ico/" + name + ".ico")
	jot.FatalIfErr(err)
	jot.FatalIfErr(ico.Encode(f, getImages([]int{256, 48, 32, 16}, imgs)...))
	jot.FatalIfErr(f.Close())

	f, err = os.Create("png/" + name + ".png")
	jot.FatalIfErr(err)
	jot.FatalIfErr(png.Encode(f, getImage(256, imgs)))
	jot.FatalIfErr(f.Close())
}

func loadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errs.Wrap(err)
	}
	defer xio.CloseIgnoringErrors(f)
	img, _, err := image.Decode(f)
	return img, errs.Wrap(err)
}

func getImages(sizes []int, images []image.Image) []image.Image {
	list := make([]image.Image, 0, len(sizes))
	for _, size := range sizes {
		if img := getImage(size, images); img != nil {
			list = append(list, img)
		}
	}
	return list
}

func getImage(size int, images []image.Image) image.Image {
	for _, img := range images {
		if img.Bounds().Dx() == size {
			return img
		}
	}
	return nil
}
