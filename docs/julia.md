# Julia
## Table
Table containing average time to run on my system for diffrent versions of the Julia program. The average is calculated on the results of three tries.

| Version | Time (sec) |
|:-------:|:----:|
| Original | 11.07 |
| Go CreatePng | 7.66 |
| Concurrent Julia | 2.17 |
| Combined | 1.86 |

## Code
The code for diffrent versions of the program. Only the original and combined have the complete code. Other two show only the functions that have been changed/added compared to the original.

### Original
The original code used as a base case. This is the complete code.
```go
// Stefan Nilsson 2013-02-27

// This program creates pictures of Julia sets (en.wikipedia.org/wiki/Julia_set).
package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"math/cmplx"
	"os"
	"strconv"
)

type ComplexFunc func(complex128) complex128

var Funcs []ComplexFunc = []ComplexFunc{
	func(z complex128) complex128 { return z*z - 0.61803398875 },
	func(z complex128) complex128 { return z*z + complex(0, 1) },
	func(z complex128) complex128 { return z*z + complex(-0.835, -0.2321) },
	func(z complex128) complex128 { return z*z + complex(0.45, 0.1428) },
	func(z complex128) complex128 { return z*z*z + 0.400 },
	func(z complex128) complex128 { return cmplx.Exp(z*z*z) - 0.621 },
	func(z complex128) complex128 { return (z*z+z)/cmplx.Log(z) + complex(0.268, 0.060) },
	func(z complex128) complex128 { return cmplx.Sqrt(cmplx.Sinh(z*z)) + complex(0.065, 0.122) },
}

func main() {
	for n, fn := range Funcs {
		err := CreatePng("picture-"+strconv.Itoa(n)+".png", fn, 1024)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// CreatePng creates a PNG picture file with a Julia image of size n x n.
func CreatePng(filename string, f ComplexFunc, n int) (err error) {
	file, err := os.Create(filename)
	if err != nil {
		return
	}
	defer file.Close()
	err = png.Encode(file, Julia(f, n))
	return
}

// Julia returns an image of size n x n of the Julia set for f.
func Julia(f ComplexFunc, n int) image.Image {
	bounds := image.Rect(-n/2, -n/2, n/2, n/2)
	img := image.NewRGBA(bounds)
	s := float64(n / 4)
	for i := bounds.Min.X; i < bounds.Max.X; i++ {
		for j := bounds.Min.Y; j < bounds.Max.Y; j++ {
			n := Iterate(f, complex(float64(i)/s, float64(j)/s), 256)
			r := uint8(0)
			g := uint8(0)
			b := uint8(n % 32 * 8)
			img.Set(i, j, color.RGBA{r, g, b, 255})
		}
	}
	return img
}

// Iterate sets z_0 = z, and repeatedly computes z_n = f(z_{n-1}), n â‰¥ 1,
// until |z_n| > 2  or n = max and returns this n.
func Iterate(f ComplexFunc, z complex128, max int) (n int) {
	for ; n < max; n++ {
		if real(z)*real(z)+imag(z)*imag(z) > 4 {
			break
		}
		z = f(z)
	}
	return
}
```
### Go CreatePng
Runs the calculations for each image in diffrent goroutines.
```go
func main() {
	start := time.Now()
	wg := new(sync.WaitGroup)
	for n, fn := range Funcs {
		wg.Add(1)
		go func(n int, fn ComplexFunc) {
			err := CreatePng("image/picture-"+strconv.Itoa(n)+".png", fn, 1024)
			if err != nil {
				log.Fatal(err)
			}
			wg.Done()
		}(n, fn)
	}
	wg.Wait()
	fmt.Println("Finished in ", time.Since(start))
}
```
### Concurrent Julia
Divides the image into smaller squares ang gives each square its own goroutine. Size = n / 8 found by trial and error to be sufficient. In this case size = 128 since 1024 / 8 = 128. 
```go
// Julia returns an image of size n x n of the Julia set for f.
func Julia(f ComplexFunc, n int) image.Image {
	bounds := image.Rect(-n/2, -n/2, n/2, n/2)
	img := image.NewRGBA(bounds)
	s := float64(n / 4)

	size := n / 8
	wg := new(sync.WaitGroup)
	for lowerX, upperX := -n/2, -n/2+size; lowerX < n; lowerX, upperX = upperX, upperX+size {
		if upperX > n {
			upperX = n
		}
		for lowerY, upperY := -n/2, -n/2+size; lowerY < n; lowerY, upperY = upperY, upperY+size {
			if upperY > n {
				upperY = n
			}
			wg.Add(1)
			go func(lowerX, upperX, lowerY, upperY int) {
				for i := lowerX; i < upperX; i++ {
					for j := lowerY; j < upperY; j++ {
						n := Iterate(f, complex(float64(i)/s, float64(j)/s), 256)
						r := uint8(0)
						g := uint8(0)
						b := uint8(n % 32 * 8)
						img.Set(i, j, color.RGBA{r, g, b, 255})
					}
				}
				wg.Done()
			}(lowerX, upperX, lowerY, upperY)
		}
	}
	wg.Wait()
	return img
}

func max(x, y int) int {
	if x > y {
		return x
	} else {
		return y
	}
}
```
### Combined
Combines both Go CreatePng and Concurrent Julia in one program. Small performance boost compared to only Concurrent Julia. This is the complete code.
```go
// Emil Hultcrantz 2022-03-28 based on Stefan Nilsson 2013-02-27

// This program creates pictures of Julia sets (en.wikipedia.org/wiki/Julia_set).
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math/cmplx"
	"os"
	"strconv"
	"sync"
	"time"
)

type ComplexFunc func(complex128) complex128

var Funcs []ComplexFunc = []ComplexFunc{
	func(z complex128) complex128 { return z*z - 0.61803398875 },
	func(z complex128) complex128 { return z*z + complex(0, 1) },
	func(z complex128) complex128 { return z*z + complex(-0.835, -0.2321) },
	func(z complex128) complex128 { return z*z + complex(0.45, 0.1428) },
	func(z complex128) complex128 { return z*z*z + 0.400 },
	func(z complex128) complex128 { return cmplx.Exp(z*z*z) - 0.621 },
	func(z complex128) complex128 { return (z*z+z)/cmplx.Log(z) + complex(0.268, 0.060) },
	func(z complex128) complex128 { return cmplx.Sqrt(cmplx.Sinh(z*z)) + complex(0.065, 0.122) },
}

func main() {
	start := time.Now()
	wg := new(sync.WaitGroup)
	for n, fn := range Funcs {
		wg.Add(1)
		go func(n int, fn ComplexFunc) {
			err := CreatePng("image/picture-"+strconv.Itoa(n)+".png", fn, 1024)
			if err != nil {
				log.Fatal(err)
			}
			wg.Done()
		}(n, fn)
	}
	wg.Wait()
	fmt.Println("Finished in ", time.Since(start))
}

// CreatePng creates a PNG picture file with a Julia image of size n x n.
func CreatePng(filename string, f ComplexFunc, n int) (err error) {
	file, err := os.Create(filename)
	if err != nil {
		return
	}
	defer file.Close()
	err = png.Encode(file, Julia(f, n))
	return
}

// Julia returns an image of size n x n of the Julia set for f.
func Julia(f ComplexFunc, n int) image.Image {
	bounds := image.Rect(-n/2, -n/2, n/2, n/2)
	img := image.NewRGBA(bounds)
	s := float64(n / 4)

	size := n / 8
	wg := new(sync.WaitGroup)
	for lowerX, upperX := -n/2, -n/2+size; lowerX < n; lowerX, upperX = upperX, upperX+size {
		if upperX > n {
			upperX = n
		}
		for lowerY, upperY := -n/2, -n/2+size; lowerY < n; lowerY, upperY = upperY, upperY+size {
			if upperY > n {
				upperY = n
			}
			wg.Add(1)
			go func(lowerX, upperX, lowerY, upperY int) {
				for i := lowerX; i < upperX; i++ {
					for j := lowerY; j < upperY; j++ {
						n := Iterate(f, complex(float64(i)/s, float64(j)/s), 256)
						r := uint8(0)
						g := uint8(0)
						b := uint8(n % 32 * 8)
						img.Set(i, j, color.RGBA{r, g, b, 255})
					}
				}
				wg.Done()
			}(lowerX, upperX, lowerY, upperY)
		}
	}
	wg.Wait()
	return img
}

func max(x, y int) int {
	if x > y {
		return x
	} else {
		return y
	}
}

// Iterate sets z_0 = z, and repeatedly computes z_n = f(z_{n-1}), n â‰¥ 1,
// until |z_n| > 2  or n = max and returns this n.
func Iterate(f ComplexFunc, z complex128, max int) (n int) {
	for ; n < max; n++ {
		if real(z)*real(z)+imag(z)*imag(z) > 4 {
			break
		}
		z = f(z)
	}
	return
}
```