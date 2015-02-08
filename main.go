package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"os"
	"path/filepath"
	"sort"
	"runtime"

	"github.com/disintegration/imaging"
)

const cell_xsize int = 8
const cell_ysize int = 8
const min_xscale int = 256
const min_yscale int = 192

type ColorRating struct {
	occurrence int
	color      color.Color
}
type ColorRatings []ColorRating

func (slice ColorRatings) Len() int {
	return len(slice)
}

func (slice ColorRatings) Less(i, j int) bool {
	return slice[i].occurrence > slice[j].occurrence
}

func (slice ColorRatings) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU()) // allow to utilize all CPU to achieve maximum perfomance
	
	var err error
	var rating int

	var img_path string // image path
	var scale int
	var scale_step int
	var offsets bool

	flag.StringVar(&img_path, "i", "peep.png", "input image")
	flag.IntVar(&scale, "s", 0, "scale from 256 to scale")
	flag.IntVar(&scale_step, "ss", 1, "scale step")
	flag.BoolVar(&offsets, "o", true, "rate all 64 offsets")
	flag.Parse()

	if scale_step <= 0 {
		scale_step = 1
	}
	//------------------------
	img, err := load_image(img_path)
	if err != nil {
		log.Fatal(err)
	}

	mask_path := filepath.Dir(img_path) + "mask_" + filepath.Base(img_path)
	mask, err := load_image(mask_path)
	if err != nil {
		perform_mutations(img, scale, scale_step, offsets)
		//rating = rate_image(img)
	} else {
		perform_mutations_with_mask(img, mask, scale, scale_step, offsets)
		//rating = rate_image_with_mask(img,mask)
	}

	fmt.Printf("Mask: %v \n", mask_path)
	fmt.Printf("Scale: %v \n", scale)
	fmt.Printf("Scale Step: %v \n", scale_step)
	fmt.Printf("Rating of file %v is %v \n", img_path, rating)

	//	for i, s := range flag.Args() {
	//		fmt.Printf(" K: %v, V: %v \n", i, s)
	//	}

	//	files, _ := filepath.Glob("./eval/*.png")
	//	for _, img_path := range files {
	//		ColorRating := rate_file(img_path)
	//		fmt.Printf("ColorRating of file %v is %v \n", img_path, ColorRating)
	//	}
}

func perform_mutations_with_mask(img, mask image.Image, scale, scale_step int, offset bool) {
	fmt.Println("Perform mutations with mask!")
}

func perform_mutations(img image.Image, scale, scale_step int, offsets bool) {
	var new_img image.Image
	fmt.Println("Perform mutations!")
	if offsets {
		for cscale := min_xscale; cscale <= min_xscale+scale; cscale += scale_step {
			for yoff := 0; yoff < 8; yoff++ {
				for xoff := 0; xoff < 8; xoff++ {
					new_img = mutate_image(img, cscale, xoff, yoff)
					rating := rate_image(new_img)
					fmt.Printf("New Image scale: %v, xoff: %v, yoff:%v, rating: %v \n", new_img, cscale, xoff, yoff, rating)
				}
			}
		}

	} else {
		for cscale := min_xscale; cscale <= min_xscale+scale; cscale += scale_step {
			new_img = mutate_image(img, cscale, 0, 0)
			rating := rate_image(new_img)
			fmt.Printf("New Image scale: %v, xoff: %v, yoff:%v, rating: %v \n", cscale, 0, 0, rating)
		}
	}
	//fmt.Printf("New Image: %v \n", new_img)
}

func mutate_image(img image.Image, scale int, xoffset, yoffset int) (new_image image.Image) {
	if scale == min_xscale {
		new_image = imaging.Resize(img, scale, 0 , imaging.Box )		
	} else {
		new_image = imaging.CropCenter(imaging.Resize(img, scale, 0 , imaging.Box ), min_xscale, min_yscale)
	}

	
	return
}

func load_image(img_path string) (img image.Image, err error) {
	reader, err := os.Open(img_path)
	if err != nil {
		return
	}
	defer reader.Close()

	img, _, err = image.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}
	return
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
func rate_image_with_mask(img, mask image.Image) (image_rating int) {
	image_rating = rate_image(img)
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

	var color_ratings ColorRatings

	for k, v := range color_map {
		//fmt.Printf ("Key: %v, value: %v \n", k,v )

		color_rating := ColorRating{v, k}
		color_ratings = append(color_ratings, color_rating)

	}

	sort.Sort(color_ratings)

	var cutted_ratings ColorRatings
	if color_ratings.Len() >= 3 {
		cutted_ratings = color_ratings[2:]
	} else {
		cutted_ratings = make(ColorRatings, 0)
	}

	for _, r := range cutted_ratings {
		cell_rating += r.occurrence
	}

	//	if cell_rating != 0 {
	//		fmt.Println("Cell ColorRating:", cell_rating)
	//	}
	return
}
