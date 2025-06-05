package webcamdetection

import (
	"bytes"
	"fmt"
	"image"
	"image/png"

	"gocv.io/x/gocv"
)

func matToBytes(mat gocv.Mat) ([]byte, error) {
	img, err := mat.ToImage()
	if err != nil {
		return nil, fmt.Errorf("Error converting mat to img: %v", err)
	}

	buf := new(bytes.Buffer)
	err = png.Encode(buf, img)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
func StartOCRScanner(output chan<- string, stop <-chan struct{}) {
	window := gocv.NewWindow("Test")
	defer window.Close()
	// zoom := 1.0
	// pan := image.Point{X: 0, Y: 0}
	// viewWidth, viewHeight := 800, 600

	for {
		select {
		case <-stop:
			return
		default:
			img := gocv.IMRead("/home/blanco/Pictures/picture_2025-05-30_16-11-40.jpg", gocv.IMReadColor)
			if img.Empty() {
				fmt.Println("Image is empty")
				continue
			}
			defer img.Close()

			hsv := gocv.NewMat()
			gocv.CvtColor(img, &hsv, gocv.ColorBGRToHSV)
			lower := gocv.NewScalar(90, 50, 100, 0)
			upper := gocv.NewScalar(110, 255, 255, 0)
			mask := gocv.NewMat()
			gocv.InRangeWithScalar(hsv, lower, upper, &mask)

			inpainted := gocv.NewMat()
			gocv.Inpaint(img, mask, &inpainted, 3, gocv.Telea)

			gray := gocv.NewMat()
			defer gray.Close()
			gocv.CvtColor(inpainted, &gray, gocv.ColorBGRToGray)

			// blurred := gocv.NewMat()
			// defer blurred.Close()
			// gocv.GaussianBlur(gray, &blurred, image.Pt(5, 5), 0, 0, gocv.BorderDefault)

			rect := gray.Region(image.Rect(69, 230, 271, 271))
			equalized := gocv.NewMat()
			gocv.EqualizeHist(rect, &equalized)

			// resized := gocv.NewMat()
			// gocv.Resize(equalized, &resized, image.Pt(equalized.Cols()*2, equalized.Rows()*2), 0, 0, gocv.InterpolationLinear)
			binary := gocv.NewMat()
			gocv.AdaptiveThreshold(equalized, &binary, 255, gocv.AdaptiveThresholdMean, gocv.ThresholdBinaryInv, 11, 2)

			kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(2, 2))
			cleaned := gocv.NewMat()
			gocv.MorphologyEx(binary, &cleaned, gocv.MorphOpen, kernel)

			final := gocv.NewMat()
			mask.CopyTo(&final)
			// edged := gocv.NewMat()
			// gocv.Canny(blurred, &edged, 50, 200)
			// size := image.Point{
			// 	X: int(float64(edged.Cols()) * zoom),
			// 	Y: int(float64(edged.Rows()) * zoom),
			// }
			// resized := gocv.NewMat()
			// defer resized.Close()
			// gocv.Resize(img, &resized, size, 0, 0, gocv.InterpolationArea)
			//
			// // Clamp view size if resized image is smaller than viewport
			// if resized.Cols() < viewWidth || resized.Rows() < viewHeight {
			// 	gocv.Resize(edged, &resized, image.Pt(viewWidth, viewHeight), 0, 0, gocv.InterpolationLinear)
			// 	zoom = float64(viewWidth) / float64(edged.Cols())
			// 	pan.X = 0
			// 	pan.Y = 0
			// }
			//
			// // Clamp pan to avoid going outside bounds
			// if pan.X < 0 {
			// 	pan.X = 0
			// }
			// if pan.Y < 0 {
			// 	pan.Y = 0
			// }
			// if pan.X+viewWidth > resized.Cols() {
			// 	pan.X = resized.Cols() - viewWidth
			// 	if pan.X < 0 {
			// 		pan.X = 0
			// 	}
			// }
			// if pan.Y+viewHeight > resized.Rows() {
			// 	pan.Y = resized.Rows() - viewHeight
			// 	if pan.Y < 0 {
			// 		pan.Y = 0
			// 	}
			// }
			//
			// tempRoi := resized.Region(image.Rect(pan.X, pan.Y, pan.X+viewWidth, pan.Y+viewHeight))
			// roi := gocv.NewMat()
			// tempRoi.CopyTo(&roi)

			// switch key {
			// case 'q':
			// 	return
			// case '+', '=':
			// 	zoom *= 1.1
			// case '-':
			// 	zoom /= 1.1
			// case 'w':
			// 	pan.Y -= 20
			// case 's':
			// 	pan.Y += 20
			// case 'a':
			// 	pan.X -= 20
			// case 'd':
			// 	pan.X += 20
			// case ']': // increase width
			// 	viewWidth += 20
			// 	if viewWidth > resized.Cols() {
			// 		viewWidth = resized.Cols()
			// 	}
			// case '[': // decrease width
			// 	if viewWidth > 100 {
			// 		viewWidth -= 20
			// 	}
			// case '\'': // increase height (single quote key)
			// 	viewHeight += 20
			// 	if viewHeight > resized.Rows() {
			// 		viewHeight = resized.Rows()
			// 	}
			// case ';': // decrease height
			// 	if viewHeight > 100 {
			// 		viewHeight -= 20
			// 	}
			// }
			//
			// // Clamp pan again because crop size may have changed
			// if pan.X+viewWidth > resized.Cols() {
			// 	pan.X = resized.Cols() - viewWidth
			// 	if pan.X < 0 {
			// 		pan.X = 0
			// 	}
			// }
			// if pan.Y+viewHeight > resized.Rows() {
			// 	pan.Y = resized.Rows() - viewHeight
			// 	if pan.Y < 0 {
			// 		pan.Y = 0
			// 	}
			// }
			window.IMShow(final)

			window.WaitKey(0) //
			bytes, err := matToBytes(final)
			if err != nil {
				fmt.Println("Error converting mat to bytes: ", err)
				continue
			}
			ProcessText(bytes)
		}
	}
}

// func StartOCRScanner(output chan<- string, stop <-chan struct{}) {
// 	// webcam, _ := gocv.VideoCaptureDevice(0)
// 	// defer webcam.Close()
// 	window := gocv.NewWindow("Test")
// 	defer window.Close()
// 	//TODO: Move to separate function
// 	client := gosseract.NewClient()
// 	client.SetLanguage("jpn", "eng")
// 	defer client.Close()
// 	/////
// 	zoom := 1.0
// 	pan := image.Point{X: 0, Y: 0}
// 	viewWidth, viewHeight := 800, 600
// 	for {
// 		select {
// 		case <-stop:
// 			return
// 		default:
// 			// if ok := webcam.Read(&img); !ok || img.Empty() {
// 			// 	continue
// 			// }
//
// 			// img := gocv.NewMat()
// 			img := gocv.IMRead("/home/blanco/Pictures/picture_2025-05-30_16-11-40.jpg", gocv.IMReadColor)
// 			if img.Empty() {
// 				fmt.Println("Image is empty")
// 			}
// 			defer img.Close()
// 			gray := gocv.NewMat()
// 			defer gray.Close()
// 			gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)
//
// 			blurred := gocv.NewMat()
// 			gocv.GaussianBlur(gray, &blurred, image.Pt(5, 5), 0, 0, gocv.BorderDefault)
//
// 			resized := gocv.NewMat()
// 			size := image.Point{
// 				X: int(float64(blurred.Cols()) * zoom),
// 				Y: int(float64(blurred.Rows()) * zoom),
// 			}
// 			gocv.Resize(blurred, &resized, size, 0, 0, gocv.InterpolationArea)
//
// 			// Avoid negative dimensions (can happen if zoomed image is too small)
// 			if resized.Cols() < viewWidth || resized.Rows() < viewHeight {
// 				gocv.Resize(blurred, &resized, image.Pt(viewWidth, viewHeight), 0, 0, gocv.InterpolationLinear)
// 				zoom = float64(viewWidth) / float64(blurred.Cols())
// 				pan.X = 0
// 				pan.Y = 0
// 			}
// 			if pan.X < 0 {
// 				pan.X = 0
// 			}
// 			if pan.Y < 0 {
// 				pan.Y = 0
// 			}
// 			if pan.X+viewWidth > resized.Cols() {
// 				pan.X = resized.Cols() - viewWidth
// 				if pan.X < 0 {
// 					pan.X = 0
// 				}
// 			}
// 			if pan.Y+viewHeight > resized.Rows() {
// 				pan.Y = resized.Rows() - viewHeight
// 				if pan.Y < 0 {
// 					pan.Y = 0
// 				}
// 			}
//
// 			roi := resized.Region(image.Rect(pan.X, pan.Y, pan.X+viewWidth, pan.Y+viewHeight))
// 			roiCopy := gocv.NewMat()
// 			roi.CopyTo(&roiCopy)
// 			roi.Close()
// 			window.IMShow(roiCopy)
// 			key := window.WaitKey(30)
// 			switch key {
// 			case 'q':
// 				return
// 			case '+', '=':
// 				zoom *= 1.1
// 			case '-':
// 				zoom /= 1.1
// 			case 'w':
// 				pan.Y -= 20
// 			case 's':
// 				pan.Y += 20
// 			case 'a':
// 				pan.X -= 20
// 			case 'd':
// 				pan.X += 20
// 			case ']':
// 				viewWidth += 20
// 				if viewWidth > resized.Cols() {
// 					viewWidth = resized.Cols()
// 				}
// 			case '[':
// 				if viewWidth > 100 {
// 					viewWidth -= 20
// 				}
// 			}
// 			if pan.X+viewWidth > resized.Cols() {
// 				pan.X = resized.Cols() - viewWidth
// 				if pan.X < 0 {
// 					pan.X = 0
// 				}
// 			}
// 			if pan.Y+viewHeight > resized.Rows() {
// 				pan.Y = resized.Rows() - viewHeight
// 				if pan.Y < 0 {
// 					pan.Y = 0
// 				}
// 			}
// 			fmt.Printf("WIDTH: %i\nHEIGHT: %i", viewWidth, viewHeight)
// 			// window.IMShow(blurred)
// 			// window.WaitKey(1)
// 			// hsv := gocv.NewMat()
// 			// gocv.CvtColor(img, &hsv, gocv.ColorBGRToHSV)
// 			//
// 			// lowerBlue := gocv.NewScalar(100, 100, 100, 0)
// 			// upperBlue := gocv.NewScalar(140, 255, 255, 0)
// 			// mask := gocv.NewMat()
// 			// gocv.InRangeWithScalar(hsv, lowerBlue, upperBlue, &mask)
// 			// gocv.CvtColor(gray, &gray, gocv.ColorGrayToBGR)
// 			// grayMasked := gocv.NewMat()
// 			// defer grayMasked.Close()
// 			// gray.CopyToWithMask(&grayMasked, mask)
// 			//
// 			// invMask := gocv.NewMat()
// 			// defer invMask.Close()
// 			// gocv.BitwiseNot(mask, &invMask)
// 			//
// 			// nonBlue := gocv.NewMat()
// 			// defer nonBlue.Close()
// 			// img.CopyToWithMask(&nonBlue, invMask)
//
// 			// binary := gocv.NewMat()
// 			// gocv.Threshold(gray, &binary, 0, 255, gocv.ThresholdBinaryInv+gocv.ThresholdOtsu)
// 			//
// 			// denoised := gocv.NewMat()
// 			// gocv.FastNlMeansDenoising(binary, &denoised)
// 			// if nonBlue.Empty() {
// 			// 	fmt.Println("nonBlue is empty")
// 			// }
// 			// if grayMasked.Empty() {
// 			// 	fmt.Println("grayMasked is empty")
// 			// }
// 			// result := gocv.NewMat()
// 			// gocv.Add(nonBlue, grayMasked, &result)
// 			// if result.Empty() {
// 			// 	fmt.Println("processed image is empty")
// 			// }
// 			// defer result.Close()
// 			// 	buf, err := gocv.IMEncode(".png", result)
// 			// 	if err != nil {
// 			// 		fmt.Println("IMEncode error: ", err)
// 			// 		continue
// 			// 	}
// 			// defer buf.Close()
//
// 			// window.IMShow(gray)
// 			// window.WaitKey(1)
//
// 			bytes, err := matToBytes(roiCopy)
// 			if err != nil {
// 				fmt.Println("Error converting mat to bytes: ", err)
// 			}
// 			client.SetImageFromBytes(bytes)
// 			text, err := client.Text()
// 			if err != nil {
// 				fmt.Println("error getting text: ", err)
// 			}
// 			fmt.Println(text)
// 			fmt.Println("####################################")
// 			fmt.Print("\n\n\n\n\n\n")
// 			// webcam.Read(&img)
// 		}
// 	}
// }
