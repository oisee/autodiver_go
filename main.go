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
	"runtime"
	"sort"

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

type MutationStats struct {
	rating  int
	xscale  int
	xoffset int
	yoffset int
}
type MutationRating []MutationStats

func main() {
	var err error

	var img_path string // image path
	var scale int
	var scale_step int
	var offsets bool
	var maxcpu bool
	var best_dir string

	//flag.StringVar(&img_path, "i", "geometry.png", "input image")
	//flag.StringVar(&best_dir, "b", "./best", "output directory")
	flag.Usage = usage
	flag.IntVar(&scale, "s", 0, "scale from 256 to (256 + <argument>)")
	flag.IntVar(&scale_step, "ss", 1, "scale step")
	flag.BoolVar(&offsets, "o", true, "rate all 64 offsets")
	flag.BoolVar(&maxcpu, "maxcpu", true, "allow max cpu usage")

	flag.Parse()
	//------------------------
	img_path = flag.Arg(0)
	if img_path == "" {
		flag.Usage()
		log.Fatal()
	}

	best_dir = flag.Arg(1)
	if best_dir == "" {
		best_dir = "./best"
	}

	if maxcpu {
		runtime.GOMAXPROCS(runtime.NumCPU()) // allow to utilize all CPU to achieve maximum perfomance
	}
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

	muta_rating := perform_mutations(img, mask, scale, scale_step, offsets)
	sort.Sort(muta_rating)

	//fmt.Printf("Muta rating: %v \n", muta_rating)

	best := muta_rating[0:8]
	worst := muta_rating[len(muta_rating)-2 : len(muta_rating)-1]

	if exists(best_dir) == false {
		os.Mkdir(best_dir, 0777)
	}

	save_images(img_path, mask_path, img, mask, best, "best", best_dir)
	save_images(img_path, mask_path, img, mask, worst, "worst", best_dir)
}

func save_images(img_path, mask_path string, img, mask image.Image, best MutationRating, postfix, dir string) {
	var err error
	var new_img, new_mask image.Image
	var new_img_path, new_mask_path string
	ext := filepath.Ext(img_path)

	background := find_background(img)

	for i, v := range best {
		new_img_path = fmt.Sprintf("%v/%v_%v_%v_rate%v_s%v_xoff%v_yoff%v%v", dir, filepath.Base(img_path), postfix, i, v.rating, v.xscale, v.xoffset, v.yoffset, ext)
		new_mask_path = fmt.Sprintf("%v/%v_%v_%v_rate%v_s%v_xoff%v_yoff%v%v", dir, filepath.Base(mask_path), postfix, i, v.rating, v.xscale, v.xoffset, v.yoffset, ext)

//		fmt.Printf("New Img Path: %v \n", new_img_path)
//		fmt.Printf("New Mask Path: %v \n", new_mask_path)
		if mask == nil {
			new_img = mutate_image(img, v.xscale, v.xoffset, v.yoffset, background)
			err = imaging.Save(new_img, new_img_path)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			new_img = mutate_image(img, v.xscale, v.xoffset, v.yoffset, background)
			new_mask = mutate_image(mask, v.xscale, v.xoffset, v.yoffset, background)
			err = imaging.Save(new_img, new_img_path)
			if err != nil {
				log.Fatal(err)
			}
			err = imaging.Save(new_mask, new_mask_path)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	return
}

func perform_mutations(img, mask image.Image, scale, scale_step int, offsets bool) (muta_rating MutationRating) {
	var new_img, new_mask image.Image
	var xoffset_max, yoffset_max = 0, 0

	background := find_background(img)

	fmt.Println("image mutation started")
	if offsets {
		xoffset_max = cell_xsize
		yoffset_max = cell_ysize
	}
	for cscale := min_xscale; cscale <= min_xscale+scale; cscale += scale_step {
		fmt.Printf("   Scale:%v\n", scale )
		for yoff := 0; yoff < yoffset_max; yoff++ {
			for xoff := 0; xoff < xoffset_max; xoff++ {
				var rating int
				if mask == nil {
					new_img = mutate_image(img, cscale, xoff, yoff, background)
					rating = rate_image(new_img)
				} else {
					new_img = mutate_image(img, cscale, xoff, yoff, background)
					new_mask = mutate_image(img, cscale, xoff, yoff, color.Black)
					rating = rate_image_with_mask(new_img, new_mask)
				}
				muta_rating = append(muta_rating, MutationStats{rating, cscale, xoff, yoff})
				//fmt.Printf("New Image scale: %v, xoff: %v, yoff:%v, rating: %v \n", cscale, xoff, yoff, rating)
				//fmt.Printf("Muta rating: %v \n", muta_rating[len(muta_rating)-1] )
			}
		}
	}
	fmt.Println("image mutation finished")
	return
}

func mutate_image(img image.Image, scale int, xoffset, yoffset int, background color.Color) (new_image image.Image) {
	if background == nil {
		background = color.Black
	}

	var scaled_image image.Image
	if scale == min_xscale {
		scaled_image = imaging.Resize(img, scale, 0, imaging.Box)
	} else {
		scaled_image = imaging.CropCenter(imaging.Resize(img, scale, 0, imaging.Box), min_xscale, min_yscale)
	}
	if xoffset == 0 && yoffset == 0 {
		new_image = scaled_image
	} else {
		new_image = imaging.New(min_xscale, min_yscale, background)
		new_image = imaging.Paste(new_image, scaled_image, image.Pt(xoffset, yoffset))
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

func find_background(img image.Image) (background color.Color) {
	bounds := img.Bounds()

	var color_map map[color.Color]int
	color_map = make(map[color.Color]int)

	for yp := bounds.Min.Y; yp < bounds.Max.Y; yp++ {
		for xp := bounds.Min.X; xp < bounds.Max.X; xp++ {
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

	background = color_ratings[0].color

	return
}

func rate_image(img image.Image) (image_rating int) {
	bounds := img.Bounds()

	xsize := bounds.Max.X - bounds.Min.X
	ysize := bounds.Max.Y - bounds.Min.X
	//fmt.Printf("Width: %v Height: %v \n", xsize, ysize)

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

func (slice MutationRating) Len() int {
	return len(slice)
}
func (slice MutationRating) Less(i, j int) bool {
	return slice[i].rating < slice[j].rating
}
func (slice MutationRating) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}
func (slice ColorRatings) Len() int {
	return len(slice)
}
func (slice ColorRatings) Less(i, j int) bool {
	return slice[i].occurrence > slice[j].occurrence
}
func (slice ColorRatings) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return false
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "%s image_file_name.png [./best_dir] [--flags]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "%s image_file_name.png \n", os.Args[0])
	fmt.Fprintf(os.Stderr, "%s image_file_name.png -s=64 -ss=4\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "%s image_file_name.png ./best -s=10 -ss=1\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\n")
}
