/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */
package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"math/big"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		_, _ = fmt.Fprintln(os.Stderr, "需要指定一个文件")
		os.Exit(1)
	}

	var info os.FileInfo
	var err error

	if info, err = os.Stat(os.Args[1]); err != nil {
		panic(err)
	}

	var larger bool = false

	if info.Size() > 102400 {
		larger = true
	} else if info.Size() >= 51200 {
		fmt.Println("文件大小满足条件，跳过")
		os.Exit(2)
	}

	var file *os.File

	if file, err = os.Open(os.Args[1]); err != nil {
		panic(err)
	}

	var img image.Image
	if img, err = jpeg.Decode(file); err != nil {
		panic(err)
	}

	var factor *big.Float = big.NewFloat(1.0)
	var MOD, MAX_TRY = func(larger bool) (*big.Float, int) {
		if larger {
			return big.NewFloat(-0.01), 100
		} else {
			return big.NewFloat(0.01), 10000
		}
	}(larger)

	var bounds image.Rectangle = img.Bounds()
	img.Bounds().Dx()

	var buf bytes.Buffer

	for i := 0; i < MAX_TRY; i++ {
		var currentScale float64
		currentScale, _ = factor.Float64()
		fmt.Println("---")
		fmt.Printf("当前缩放比例：%.2f*\n", currentScale)
		fmt.Printf("当前图片分辨率：%d*%d\n", multiply(bounds.Dx(), factor), multiply(bounds.Dy(), factor))

		var newImg image.Image = resize.Resize(multiply(bounds.Dx(), factor), multiply(bounds.Dy(), factor), img, resize.NearestNeighbor)
		if err = jpeg.Encode(&buf, newImg, nil); err != nil {
			panic(err)
		}
		fmt.Printf("当前图片大小：%.2fkb\n", float64(buf.Len())/1024.0)
		if (larger && buf.Len() <= 102400) || (!larger && buf.Len() >= 51200) {
			if (larger && buf.Len() < 51200) || (!larger && buf.Len() > 102400) {
				panic(errors.New("无法找到一个合适的大小"))
			}
			var output string = strings.TrimSuffix(os.Args[1], filepath.Ext(os.Args[1])) + "-mod" + filepath.Ext(os.Args[1])
			var outputFile *os.File
			if outputFile, err = os.Create(output); err != nil {
				panic(err)
			}
			if _, err = outputFile.ReadFrom(&buf); err != nil {
				panic(err)
			}
			fmt.Println("写入完成")
			return
		}
		buf.Reset()
		factor.Add(factor, MOD)
	}
	panic(errors.New("无法找到一个合适的大小"))
}

func multiply(a int, b *big.Float) uint {
	var bValue float64
	bValue, _ = b.Float64()
	var f = big.NewFloat(bValue)
	var i, _ = f.Mul(
		f, big.NewFloat(float64(a)),
	).Int64()
	return uint(i)
}
