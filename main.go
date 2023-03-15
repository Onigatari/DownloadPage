package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/nfnt/resize"
	"github.com/signintech/gopdf"
	"image"
	"image/jpeg"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func downloadFile(url, fileName string) error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New("Received non 200 response code")
	}
	os.Mkdir("img/", os.ModePerm)
	file, err := os.Create("img/" + fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}

func getImageDimension(imagePath string) (int, int) {
	file, err := os.Open(imagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	image, _, err := image.DecodeConfig(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", imagePath, err)
	}
	return image.Width, image.Height
}

func pageResize(url string) {
	file, err := os.Open(url)
	if err != nil {
		log.Fatal(err)
	}

	img, err := jpeg.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	file.Close()

	m := resize.Resize(1240, 1685, img, resize.Lanczos3)

	out, err := os.Create(url)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	jpeg.Encode(out, m, nil)
}

func removeContents(dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return err
	}

	for _, file := range files {
		err = os.RemoveAll(file)
		if err != nil {
			return err
		}
	}

	return nil
}

func parseURL(textUrl string) string {
	ind := strings.Index(textUrl, ".jpg")
	if ind == -1 {
		return textUrl
	}

	return textUrl[:ind-1]
}

func createFilePDF(mainUrl string, n int) {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA2})
	for ind := 1; ind <= n; ind++ {
		urls := mainUrl + strconv.Itoa(ind) + ".jpg"
		tmp := strconv.Itoa(ind) + ".jpg"
		log.Println(urls)

		err := downloadFile(mainUrl+strconv.Itoa(ind)+".jpg?1664764545", tmp)
		if err != nil {
			log.Fatal(err)
		}

		pageResize("img/" + tmp)
		width, height := getImageDimension("img/" + tmp)
		pdf.AddPage()
		pdf.Image("img/"+tmp, 0, 0, &gopdf.Rect{W: float64(width), H: float64(height)})
	}

	err := pdf.WritePdf("image.pdf")
	if err != nil {
		log.Println("Error Create PDF")
		return
	}

	//err = removeContents("img/")
	if err != nil {
		log.Println("Error Remote Dir")
		return
	}
}

func main() {
	var inputURL string
	var countPage int

	flag.StringVar(&inputURL, "url", "", "Ссылка на скачиваемую книгу")
	flag.IntVar(&countPage, "count", 0, "Кол-во страниц книги")
	flag.Parse()

	bookUrl := parseURL(inputURL)
	createFilePDF(bookUrl, countPage)
	log.Println("Main Finished")
}
