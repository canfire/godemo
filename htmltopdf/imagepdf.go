package htmltopdf

import (
	"bytes"
	"io"
	"log"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

func PhotoPDF() {
	// 要写入 PDF 的图片列表（顺序 = PDF 页顺序）
	images := []string{
		"a/企业微信截图_17708032547521 copy.png",
		"a/images.jpeg",
		"a/企业微信截图_17708032547521.png",
	}

	// 输出 PDF 文件
	outFile := "output.pdf"

	// import 配置（默认即可）
	conf := model.NewDefaultConfiguration()

	// 将图片写入 PDF
	err := api.ImportImagesFile(images, outFile, nil, conf)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("PDF 生成成功:", outFile)
}

func convertImagesToPDF(imageReaders []io.ReadCloser) ([]byte, error) {
	// 将 ReadCloser 转换为 Reader
	var imgs []io.Reader
	for _, rc := range imageReaders {
		defer rc.Close()

		// 读取到内存
		data, err := io.ReadAll(rc)
		if err != nil {
			return nil, err
		}
		imgs = append(imgs, bytes.NewReader(data))
	}

	// 创建输出缓冲区
	var pdfBuffer bytes.Buffer

	// 创建配置
	conf := model.NewDefaultConfiguration()

	// 创建 Import 配置
	// imp := &pdfcpu.Import{
	// 	PageDim: types.PaperSize["A4"], // 或者不设置使用默认
	// }

	// 第一个参数是现有的 PDF (nil 表示创建新的)
	err := api.ImportImages(nil, &pdfBuffer, imgs, nil, conf)
	if err != nil {
		return nil, err
	}

	return pdfBuffer.Bytes(), nil
}
