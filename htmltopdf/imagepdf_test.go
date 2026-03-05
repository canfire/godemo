package htmltopdf

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestConvertImagesToPDFFromLocalFiles(t *testing.T) {
	paths := []string{
		filepath.Join("a", "企业微信截图_17708032547521 copy.png"),
		filepath.Join("a", "images.jpeg"),
		filepath.Join("a", "企业微信截图_17708032547521.png"),
	}

	readers := make([]io.ReadCloser, 0, len(paths))
	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			if os.IsNotExist(err) {
				t.Skipf("test image not found: %s", path)
			}
			t.Fatalf("open image %s: %v", path, err)
		}
		readers = append(readers, f)
	}

	defer func() {
		for _, r := range readers {
			r.Close()
		}
	}()

	pdfBytes, err := convertImagesToPDF(readers)
	if err != nil {
		t.Fatalf("convertImagesToPDF failed: %v", err)
	}
	if len(pdfBytes) == 0 {
		t.Fatalf("expected non-empty PDF output")
	}

	// Save to local file
	outputPath := filepath.Join("a", "output.pdf")
	err = os.WriteFile(outputPath, pdfBytes, 0644)
	if err != nil {
		t.Fatalf("failed to save PDF: %v", err)
	}
}
