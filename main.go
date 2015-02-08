package main

import (
	//"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"os"
	"path/filepath"
	"sort"
)

const cell_xsize int = 8
const cell_ysize int = 8

type Rating struct {
	occurrence int
	color      color.Color
}
type Ratings []Rating

func (slice Ratings) Len() int {
	return len(slice)
}

func (slice Ratings) Less(i, j int) bool {
	return slice[i].occurrence > slice[j].occurrence
}

func (slice Ratings) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func main() {
	//flag.Parse()
	//root := flag.Arg(0)
	//var err error

	files, _ := filepath.Glob("./eval/*.png")
	for _, img_path := range files {
		fmt.Println(img_path)
		rating := rate_file(img_path)
		fmt.Printf("Rating of file %v is %v \n", img_path, rating)
	}

}

func rate_file(img_path string) (file_rating int) {

	reader, err := os.Open(img_path)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	img, _, err := image.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}

	file_rating = rate_image(img)
	return
}

func rate_image(img image.Image) (image_rating int) {
	bounds := img.Bounds()

	xsize := bounds.Max.X - bounds.Min.X
	ysize := bounds.Max.Y - bounds.Min.X
	fmt.Printf("Width: %v Height: %v \n", xsize, ysize)

	xcells := int(xsize / cell_xsize)
	ycells := int(ysize / cell_ysize)

	for yc := 0; yc < ycells; yc++ {
		for xc := 0; xc < xcells; xc++ {
			image_rating += rate_image_cell(img, xc*cell_xsize, yc*cell_ysize)
		}
	}

	return
}

func rate_image_cell(img image.Image, x, y int) (cell_rating int) {

	var color_map map[color.Color]int
	color_map = make(map[color.Color]int)

	for xp := x; xp < x+cell_xsize; xp++ {
		for yp := y; yp < y+cell_ysize; yp++ {
			color := img.At(xp, yp)
			color_map[color] += 1
		}
	}

	var ratings Ratings

	for k, v := range color_map {
		//fmt.Printf ("Key: %v, value: %v \n", k,v )

		rating := Rating{v, k}
		ratings = append(ratings, rating)

	}

	sort.Sort(ratings)

	var cutted_ratings Ratings
	if ratings.Len() >= 3 {
		cutted_ratings = ratings[2:]
	} else {
		cutted_ratings = make(Ratings, 0)
	}

	for _, r := range cutted_ratings {
		cell_rating += r.occurrence
	}

	if cell_rating != 0 {
		fmt.Println("Cell rating:", cell_rating)
	}
	return
}
